import type { TypedEventTarget } from "@/utils/TypedEventTarget";

// 照搬 SPlayer core/audio-player/IPlaybackEngine.ts 与 BaseAudioPlayer.ts 中的事件定义。
// Sonora 只有 Web Audio 一种引擎，不迁入 MPV、FFmpeg、自动混音的滤波与交叉淡入淡出。

export const AUDIO_EVENTS = {
  PLAY: "play",
  PAUSE: "pause",
  ENDED: "ended",
  TIME_UPDATE: "timeupdate",
  ERROR: "error",
  CAN_PLAY: "canplay",
  LOAD_START: "loadstart",
  SEEKED: "seeked",
  WAITING: "waiting",
  VOLUME_CHANGE: "volumechange",
  PLAYING: "playing",
  SEEKING: "seeking",
  EMPTIED: "emptied",
} as const;

export type AudioEventType = (typeof AUDIO_EVENTS)[keyof typeof AUDIO_EVENTS];

export interface AudioErrorDetail {
  originalEvent?: Event;
  errorCode: number;
}

export type AudioEventMap = {
  [K in AudioEventType]: K extends typeof AUDIO_EVENTS.ERROR
    ? CustomEvent<AudioErrorDetail>
    : CustomEvent<undefined>;
};

export enum AudioErrorCode {
  /** 用户中止 */
  ABORTED = 1,
  /** 网络错误 */
  NETWORK = 2,
  /** 解码错误 */
  DECODE = 3,
  /** 格式不支持 */
  SRC_NOT_SUPPORTED = 4,
  /** DOMException: AbortError */
  DOM_ABORT = 20,
}

export type FadeCurve = "linear" | "exponential" | "equalPower";

export interface PlayOptions {
  /** 是否自动播放 */
  autoPlay?: boolean;
  /** 是否渐入 */
  fadeIn?: boolean;
  /** 渐入时长（秒） */
  fadeDuration?: number;
  fadeCurve?: FadeCurve;
  /** 开始播放的位置（秒） */
  seek?: number;
}

export interface PauseOptions {
  /** 是否渐出 */
  fadeOut?: boolean;
  /** 渐出时长（秒） */
  fadeDuration?: number;
  fadeCurve?: FadeCurve;
}

/**
 * 播放引擎。播放控制器只经这个接口操作音频，单元测试换成假的实现。
 */
export interface IPlaybackEngine extends TypedEventTarget<AudioEventMap> {
  /**
   * 加载并播放音频
   * @param url 音频地址；省略时继续播放当前音频
   * @throws 播放失败时抛出原始错误（如自动播放限制的 NotAllowedError）
   */
  play(url?: string, options?: PlayOptions): Promise<void>;
  /** 恢复播放 */
  resume(options?: { fadeIn?: boolean; fadeDuration?: number }): Promise<void>;
  /** 暂停 */
  pause(options?: PauseOptions): void;
  /** 停止播放并清除当前音频，旧音频之后不再触发事件 */
  stop(): void;
  /**
   * 跳转
   * @param time 时间（秒）
   */
  seek(time: number): void;
  /** 音频总时长（秒），未知时为 0 */
  readonly duration: number;
  /** 当前播放时间（秒） */
  readonly currentTime: number;
  readonly paused: boolean;
  /** 当前音频地址，没有时为空字符串 */
  readonly src: string;
  /** 设置音量 (0.0 - 1.0) */
  setVolume(value: number): void;
  getVolume(): number;
  /** 设置输出设备 */
  setSinkId(deviceId: string): Promise<void>;
  /** 设置均衡器第 index 段的增益（dB） */
  setFilterGain(index: number, value: number): void;
  /** 频谱数据，没有音频图谱时为空数组 */
  getFrequencyData(): Uint8Array;
  /** 低频音量 [0.0-1.0] */
  getLowFrequencyVolume(): number;
  /** 当前音频的错误码，见 AudioErrorCode；没有错误时为 0 */
  getErrorCode(): number;
}
