import { TypedEventTarget } from "@/utils/TypedEventTarget";
import { AudioEffectManager } from "./AudioEffectManager";
import type {
  AudioEventMap,
  FadeCurve,
  IPlaybackEngine,
  PauseOptions,
  PlayOptions,
} from "./IPlaybackEngine";
import {
  getSharedAudioContext,
  getSharedMasterInput,
  type IExtendedAudioContext,
} from "./SharedAudioContext";

export { AUDIO_EVENTS, AudioErrorCode } from "./IPlaybackEngine";

const SEEK_FADE_TIME = 0.05;

/**
 * 音频播放器抽象基类
 *
 * 照搬 SPlayer core/audio-player/BaseAudioPlayer.ts：管理 AudioContext、音量增益、EQ 连接、
 * 以及通用的淡入淡出/Seek 逻辑。去掉 ReplayGain、倍速、延迟补偿与自动混音的滤波。
 */
export abstract class BaseAudioPlayer
  extends TypedEventTarget<AudioEventMap>
  implements IPlaybackEngine
{
  /** 核心上下文 */
  protected audioCtx: IExtendedAudioContext | null = null;
  /** 主输出增益节点 (控制音量) */
  protected gainNode: GainNode | null = null;
  /** 输入节点 (子类将源连接到此处) */
  protected inputNode: GainNode | null = null;

  protected effectManager: AudioEffectManager | null = null;

  /** 初始化状态 */
  protected isInitialized = false;
  /** 目标音量 (0-1) */
  protected volume: number = 1;

  /** 音频与歌词同步补偿（毫秒） */
  protected audioDelayCompensation = 0;
  /** 均衡器增益，音频图谱建立前设置的值在建立时应用 */
  private filterGains: number[] = [];
  /** 存储淡出暂停的定时器 ID */
  private fadeTimer: ReturnType<typeof setTimeout> | null = null;

  /**
   * 初始化音频图谱
   * 链路: InputNode(子类 Source) -> EffectManager -> GainNode -> Destination
   */
  public init() {
    if (this.isInitialized) return;

    try {
      this.audioCtx = getSharedAudioContext();
      this.inputNode = this.audioCtx.createGain();
      this.inputNode.gain.value = 1; // 直通
      this.gainNode = this.audioCtx.createGain();
      this.effectManager = new AudioEffectManager(this.audioCtx);

      // 连接链路: Input -> EQ/Spectrum -> MasterGain -> Speaker
      // AudioEffectManager.connect 接受输入节点，内部串联后返回输出节点
      const processedNode = this.effectManager.connect(this.inputNode);
      processedNode.connect(this.gainNode);
      this.gainNode.connect(getSharedMasterInput());

      // 应用初始音量与均衡器
      this.gainNode.gain.value = this.volume;
      this.filterGains.forEach((value, index) =>
        this.effectManager?.setFilterGain(index, value),
      );

      this.isInitialized = true;

      // 通知子类连接其特定的源
      this.onGraphInitialized();
    } catch (e) {
      console.error("初始化 AudioContext 失败", e);
    }
  }

  /**
   * 供子类重写，当 AudioContext 初始化完成时调用
   *
   * 子类应在此处创建 SourceNode 并连接到 this.inputNode
   */
  protected abstract onGraphInitialized(): void;

  /**
   * 播放音频
   * @throws 如果播放失败，则抛出原始错误
   */
  public async play(url?: string, options: PlayOptions = {}) {
    this.cancelPendingPause();
    const shouldPlay = options.autoPlay ?? true;

    if (url) {
      await this.load(url);
    }

    if (!this.isInitialized) this.init();

    // 恢复播放位置
    if (options.seek && options.seek > 0) {
      this.doSeek(options.seek);
    }

    if (!shouldPlay) return;

    // 页面从未被点击时，AudioContext 无法恢复，resume() 不会完成；与 audio 元素一样按
    // 自动播放限制报 NotAllowedError（阶段 2 的 Chrome 音频实验）
    if (
      this.audioCtx?.state !== "running" &&
      navigator.userActivation &&
      !navigator.userActivation.hasBeenActive
    ) {
      throw new DOMException("需要先与页面交互", "NotAllowedError");
    }
    // SPlayer 只在 suspended 时恢复。暂停刚请求挂起时状态仍是 running，随后才变为
    // suspended，这里总是恢复，排在挂起之后执行
    if (this.audioCtx) {
      await this.audioCtx.resume();
    }

    const duration = options.fadeIn ? (options.fadeDuration ?? 0.5) : 0;

    // 修复：如果是渐入，强制从 0 开始
    if (duration > 0 && this.gainNode && this.audioCtx) {
      this.gainNode.gain.setValueAtTime(0, this.audioCtx.currentTime);
    }

    this.applyFadeTo(this.volume, duration, options.fadeCurve);

    try {
      await this.doPlay();
    } catch (e) {
      console.error("播放失败", e);
      throw e;
    }
  }

  public async resume(options?: {
    fadeIn?: boolean;
    fadeDuration?: number;
    fadeCurve?: FadeCurve;
  }): Promise<void> {
    await this.play(undefined, options);
  }

  public async pause(options: PauseOptions = {}) {
    this.cancelPendingPause();

    const duration = options.fadeOut ? (options.fadeDuration ?? 0.5) : 0;

    const performPause = async () => {
      await this.doPause();

      if (this.audioCtx && this.audioCtx.state === "running") {
        try {
          await this.audioCtx.suspend();
        } catch (e) {
          console.warn("挂起 AudioContext 失败", e);
        }
      }

      this.fadeTimer = null;
    };

    if (duration > 0) {
      this.applyFadeTo(0, duration, options.fadeCurve);

      this.fadeTimer = setTimeout(() => {
        performPause();
      }, duration * 1000);
    } else {
      await performPause();
    }
  }

  /**
   * 跳转进度
   * @param time 目标时间 (秒)
   */
  public async seek(time: number, immediate = false) {
    this.cancelPendingPause();
    // 如果已经暂停，直接跳转
    if (this.paused) {
      this.doSeek(time);
      return;
    }

    if (!immediate) {
      this.applyFadeTo(0, SEEK_FADE_TIME);
      await new Promise((resolve) =>
        setTimeout(resolve, SEEK_FADE_TIME * 1000),
      );
    }

    await this.doSeek(time);

    this.applyFadeTo(this.volume, immediate ? 0 : SEEK_FADE_TIME);
  }

  /**
   * 停止播放并重置
   */
  public stop() {
    this.cancelPendingPause();
    // 捕获可能产生的异步错误
    Promise.resolve(this.pause({ fadeOut: false })).catch(() => {});
    Promise.resolve(this.doSeek(0)).catch(() => {});
  }

  /**
   * 设置音量
   * @param value 0.0 - 1.0
   */
  public setVolume(value: number) {
    this.volume = Math.max(0, Math.min(1, value));
    this.applyFadeTo(this.volume, 0);
  }

  /**
   * 获取当前音量
   * @returns 当前音量值 (0.0 - 1.0)
   */
  public getVolume(): number {
    return this.volume;
  }

  /**
   * 应用音量渐变
   * @param targetValue 目标音量
   * @param duration 持续时间 (秒)
   * @param curve 渐变曲线
   */
  protected applyFadeTo(
    targetValue: number,
    duration: number,
    curve: FadeCurve = "linear",
  ) {
    if (!this.gainNode || !this.audioCtx) return;

    const currentTime = this.audioCtx.currentTime;
    // 取消之前计划的音量变化
    this.gainNode.gain.cancelScheduledValues(currentTime);

    // 设定当前值为起点 ，防止爆音
    const currentValue = this.gainNode.gain.value;
    this.gainNode.gain.setValueAtTime(currentValue, currentTime);

    if (duration <= 0) {
      const safeStartTime = currentTime + 0.02;
      if (Number.isFinite(safeStartTime) && safeStartTime > currentTime) {
        this.gainNode.gain.linearRampToValueAtTime(targetValue, safeStartTime);
      } else {
        this.gainNode.gain.setValueAtTime(targetValue, currentTime);
      }
      return;
    }

    // 稍微延后一点点开始 Ramp，给 AudioContext 内部队列一点喘息时间
    const safeStartTime = currentTime + 0.02;
    this.gainNode.gain.setValueAtTime(currentValue, safeStartTime);

    if (curve === "equalPower") {
      const steps = Math.max(2, Math.floor(duration * 60));
      const curveData = new Float32Array(steps);

      for (let i = 0; i < steps; i++) {
        const t = i / (steps - 1);
        let val = 0;

        if (targetValue > currentValue) {
          const factor = Math.sin((t * Math.PI) / 2);
          val = currentValue + (targetValue - currentValue) * factor;
        } else {
          const factor = Math.cos((t * Math.PI) / 2);
          val = targetValue + (currentValue - targetValue) * factor;
        }
        curveData[i] = val;
      }
      this.gainNode.gain.setValueCurveAtTime(
        curveData,
        safeStartTime,
        duration,
      );
    } else if (curve === "exponential") {
      let safeTarget = targetValue;
      if (safeTarget <= 0.001) safeTarget = 0.001;

      if (currentValue < 0.001) {
        this.gainNode.gain.linearRampToValueAtTime(
          targetValue,
          safeStartTime + duration,
        );
      } else {
        this.gainNode.gain.exponentialRampToValueAtTime(
          safeTarget,
          safeStartTime + duration,
        );
        if (targetValue === 0) {
          this.gainNode.gain.setValueAtTime(0, safeStartTime + duration);
        }
      }
    } else {
      this.gainNode.gain.linearRampToValueAtTime(
        targetValue,
        safeStartTime + duration,
      );
    }
  }

  /**
   * 设置音频与歌词同步补偿
   * @param offset 补偿（毫秒），正值让 currentTime 提前
   */
  public setAudioDelayCompensation(offset: number): void {
    this.audioDelayCompensation = offset;
  }

  /**
   * 设置输出设备
   */
  public async setSinkId(device: string) {
    // SPlayer 遇到 default 直接返回，切换到其他设备后就无法改回默认设备；"" 表示默认设备
    const deviceId = device === "default" ? "" : device;
    // 优先尝试 AudioContext 级设置
    if (this.audioCtx && typeof this.audioCtx.setSinkId === "function") {
      try {
        await this.audioCtx.setSinkId(deviceId);
        return;
      } catch (e) {
        console.warn("AudioContext setSinkId 失败, 尝试后备", e);
      }
    }
    // 回退逻辑由子类实现，例如设置 HTMLAudioElement.setSinkId
    await this.doSetSinkId(deviceId);
  }

  /**
   * 取消正在进行的暂停计划
   * 用于在淡出期间点击了播放或跳转，取消之前的暂停指令
   */
  protected cancelPendingPause() {
    if (this.fadeTimer) {
      clearTimeout(this.fadeTimer);
      this.fadeTimer = null;
    }
  }

  /** 获取频率数据 */
  public getFrequencyData(): Uint8Array {
    return this.effectManager
      ? this.effectManager.getFrequencyData()
      : new Uint8Array(0);
  }

  /** 获取低频音量 */
  public getLowFrequencyVolume(): number {
    return this.effectManager ? this.effectManager.getLowFrequencyVolume() : 0;
  }

  /** 设置滤波器增益 */
  public setFilterGain(index: number, value: number) {
    this.filterGains[index] = value;
    this.effectManager?.setFilterGain(index, value);
  }

  /** 加载资源 */
  public abstract load(url: string): Promise<void>;

  /** 预加载下一首的音频，null 释放 */
  public abstract preload(url: string | null): void;

  /** 执行底层播放 */
  protected abstract doPlay(): Promise<void>;

  /** 执行底层暂停 */
  protected abstract doPause(): void | Promise<void>;

  /** 执行底层 Seek */
  protected abstract doSeek(time: number): void | Promise<void>;

  /** 子类回退实现：设置输出设备 */
  protected abstract doSetSinkId(deviceId: string): Promise<void>;

  public abstract get src(): string;
  public abstract get duration(): number;
  public abstract get currentTime(): number;
  public abstract get paused(): boolean;
  public abstract getErrorCode(): number;
}
