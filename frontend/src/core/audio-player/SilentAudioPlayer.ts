import { TypedEventTarget } from "@/utils/TypedEventTarget";
import type {
  AudioEventMap,
  IPlaybackEngine,
  PlayOptions,
} from "./IPlaybackEngine";

/**
 * 不出声的播放引擎，供 fixture 使用：fixture 没有音频，播放时间停在跳转点，
 * 播放器的截图对照因此可以重复。事件顺序与 HTMLAudioElement 相同。
 */
export class SilentAudioPlayer
  extends TypedEventTarget<AudioEventMap>
  implements IPlaybackEngine
{
  private url = "";
  private time = 0;
  private isPaused = true;
  private volume = 1;

  public async play(url?: string, options: PlayOptions = {}) {
    if (url) {
      this.url = url;
      this.time = 0;
      this.dispatch("loadstart");
      queueMicrotask(() => {
        if (this.url === url) this.dispatch("canplay");
      });
    }
    if (options.seek && options.seek > 0) this.time = options.seek;
    if (!(options.autoPlay ?? true) || !this.isPaused) return;
    this.isPaused = false;
    this.dispatch("play");
    this.dispatch("playing");
  }

  public async resume() {
    await this.play();
  }

  public pause() {
    if (this.isPaused) return;
    this.isPaused = true;
    this.dispatch("pause");
  }

  public stop() {
    this.pause();
    this.url = "";
    this.time = 0;
  }

  public seek(time: number) {
    this.time = time;
    this.dispatch("seeked");
  }

  public get duration() {
    return 0;
  }

  public get currentTime() {
    return this.time;
  }

  public get paused() {
    return this.isPaused;
  }

  public get src() {
    return this.url;
  }

  public setVolume(value: number) {
    this.volume = value;
  }

  public getVolume() {
    return this.volume;
  }

  public async setSinkId() {}

  public setAudioDelayCompensation() {}

  public preload() {}

  public setFilterGain() {}

  public getFrequencyData() {
    return new Uint8Array(0);
  }

  public getLowFrequencyVolume() {
    return 0;
  }

  public getErrorCode() {
    return 0;
  }
}
