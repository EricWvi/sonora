import {
  AUDIO_EVENTS,
  AudioErrorCode,
  BaseAudioPlayer,
} from "./BaseAudioPlayer";
import type { AudioEventType } from "./IPlaybackEngine";

/**
 * 基于 HTMLAudioElement 的播放器实现
 *
 * 照搬 SPlayer core/audio-player/AudioElementPlayer.ts：使用原生 HTML5 Audio 元素播放，
 * 通过 MediaElementAudioSourceNode 连接到 Web Audio API 音频图谱。去掉倍速与音高；
 * 延迟补偿只保留手动补偿（阶段 5 的决定），不减去 outputLatency 与 baseLatency。
 */
export class AudioElementPlayer extends BaseAudioPlayer {
  /** 内部 Audio 元素 */
  private audioElement: HTMLAudioElement;
  /** MediaElementAudioSourceNode 用于连接 Web Audio API */
  private sourceNode: MediaElementAudioSourceNode | null = null;

  /** Seek 锁，用于在 seek 过程中返回稳定的 currentTime */
  private isInternalSeeking = false;
  /** 目标时间缓存，用于在 seek 过程中返回稳定的 currentTime */
  private targetSeekTime = 0;

  /** 预加载下一首的元素，保留到下一次预加载，切歌时主元素复用它下载的数据 */
  private preloadElement: HTMLAudioElement | null = null;
  private preloadUrl: string | null = null;

  constructor() {
    super();
    this.audioElement = new Audio();
    this.audioElement.crossOrigin = "anonymous";
    this.bindInternalEvents();

    this.audioElement.addEventListener("seeked", () => {
      this.isInternalSeeking = false;
    });
  }

  /**
   * 当音频图谱初始化完成时调用
   * 创建 MediaElementAudioSourceNode 并连接到输入节点
   */
  protected onGraphInitialized(): void {
    if (!this.audioCtx || !this.inputNode) return;

    try {
      if (!this.sourceNode) {
        this.sourceNode = this.audioCtx.createMediaElementSource(
          this.audioElement,
        );
      } else {
        this.sourceNode.disconnect();
      }

      // 连接: Source -> Input
      this.sourceNode.connect(this.inputNode);
    } catch (error) {
      console.error("[AudioElementPlayer] SourceNode 创建失败", error);
    }
  }

  /**
   * 加载音频资源
   * @param url 音频地址
   */
  public async load(url: string): Promise<void> {
    this.isInternalSeeking = false;
    this.audioElement.src = url;
    this.audioElement.load();
  }

  /**
   * 预加载下一首。Chrome 中同一地址、同样 crossOrigin 的媒体元素共享已下载的数据，
   * 主元素切到这首时不再等待开头的下载；用 fetch 预取不起作用（见
   * scripts/experiments/preload-next.mjs）。preload 为 metadata，只下载到能开始播放为止。
   * @param url 下一首的地址，null 时释放
   */
  public preload(url: string | null): void {
    if (url === this.preloadUrl) return;
    if (this.preloadElement) {
      this.preloadElement.removeAttribute("src");
      this.preloadElement.load();
      this.preloadElement = null;
    }
    this.preloadUrl = url;
    if (!url) return;
    const element = new Audio();
    element.crossOrigin = "anonymous";
    element.preload = "metadata";
    element.src = url;
    this.preloadElement = element;
  }

  /**
   * 执行底层播放
   * @returns 播放 Promise
   */
  protected async doPlay(): Promise<void> {
    return this.audioElement.play();
  }

  /**
   * 执行底层暂停
   */
  protected doPause(): void {
    this.audioElement.pause();
  }

  /**
   * 跳转到指定时间
   * @param time 目标时间（秒）
   */
  public async seek(time: number): Promise<void> {
    this.isInternalSeeking = true;
    this.targetSeekTime = time;

    this.cancelPendingPause();
    this.doSeek(time);
  }

  /**
   * 停止播放并清理当前音频源
   * 彻底移除 src，防止旧链接后续继续触发 canplay 等事件
   */
  public stop(): void {
    super.stop();
    this.isInternalSeeking = false;
    this.audioElement.removeAttribute("src");
    this.audioElement.load();
  }

  /**
   * 执行底层 Seek
   * @param time 目标时间（秒）
   */
  protected doSeek(time: number): void {
    if (Number.isFinite(time)) {
      this.audioElement.currentTime = time;
    }
  }

  /**
   * 设置音频输出设备
   * @param deviceId 设备 ID
   */
  protected async doSetSinkId(deviceId: string): Promise<void> {
    if (typeof this.audioElement.setSinkId === "function") {
      await this.audioElement.setSinkId(deviceId);
    }
  }

  /** 获取当前音频源地址 */
  public get src(): string {
    return this.audioElement.src || "";
  }

  /** 获取音频总时长（秒） */
  public get duration(): number {
    return this.audioElement.duration || 0;
  }

  /**
   * 获取当前播放时间（秒）
   * 如果正在 Seek，返回目标时间以避免进度跳回
   */
  public get currentTime(): number {
    if (this.isInternalSeeking) {
      return this.targetSeekTime;
    }
    // 基础时间 + 手动延迟补偿（SPlayer 只在延迟策略为 playback 时补偿，并减去自动延迟）
    return (
      (this.audioElement.currentTime || 0) + this.audioDelayCompensation / 1000
    );
  }

  /** 获取是否暂停状态 */
  public get paused(): boolean {
    return this.audioElement.paused;
  }

  /**
   * 获取错误码
   * @returns 错误码 (0: 无错误, 1: ABORTED, 2: NETWORK, 3: DECODE, 4: SRC_NOT_SUPPORTED)
   */
  public getErrorCode(): number {
    if (!this.audioElement.error) return 0;
    switch (this.audioElement.error.code) {
      case MediaError.MEDIA_ERR_ABORTED:
        return AudioErrorCode.ABORTED;
      case MediaError.MEDIA_ERR_NETWORK:
        return AudioErrorCode.NETWORK;
      case MediaError.MEDIA_ERR_DECODE:
        return AudioErrorCode.DECODE;
      case MediaError.MEDIA_ERR_SRC_NOT_SUPPORTED:
        return AudioErrorCode.SRC_NOT_SUPPORTED;
      default:
        return 0;
    }
  }

  /**
   * 监听原生 DOM 事件并转发为标准事件
   * 将 HTMLAudioElement 的事件转换为 BaseAudioPlayer 的统一事件格式
   */
  private bindInternalEvents() {
    const events: AudioEventType[] = Object.values(AUDIO_EVENTS);

    events.forEach((eventType) => {
      this.audioElement.addEventListener(eventType, (e) => {
        if (eventType === AUDIO_EVENTS.ERROR) {
          this.dispatch(AUDIO_EVENTS.ERROR, {
            originalEvent: e,
            errorCode: this.getErrorCode(),
          });
        } else {
          this.dispatch(eventType);
        }
      });
    });
  }
}
