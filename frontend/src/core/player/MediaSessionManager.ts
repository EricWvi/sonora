import type { Track } from "@/api/types";
import { coverUrl } from "@/utils/cover";
import { msToS } from "@/utils/time";
import { throttle } from "lodash-es";

/** 系统媒体控件发出的操作，由播放控制器处理 */
export interface MediaSessionActions {
  play(): void;
  pause(): void;
  nextOrPrev(type: "next" | "prev"): void;
  setSeek(time: number): void;
}

const supported = () =>
  typeof navigator !== "undefined" && "mediaSession" in navigator;

/**
 * 媒体会话。照搬 SPlayer core/player/MediaSessionManager.ts 的 Web API 部分（Navigator.mediaSession），
 * 不迁入 Electron 原生插件、Discord RPC 与“系统媒体控件”开关（Sonora 没有这项设置）。
 * 另外按播放状态设置 `playbackState`：SPlayer 的网页版只靠 audio 元素的状态，元素经
 * Web Audio 输出时 Chrome 不一定能据此判断。
 */
class MediaSessionManager {
  private currentRate: number = 1;

  /**
   * 初始化媒体会话
   * @param player 处理操作的播放控制器
   */
  public init(player: MediaSessionActions) {
    if (!supported()) return;
    const nav = navigator.mediaSession;
    nav.setActionHandler("play", () => player.play());
    nav.setActionHandler("pause", () => player.pause());
    nav.setActionHandler("previoustrack", () => player.nextOrPrev("prev"));
    nav.setActionHandler("nexttrack", () => player.nextOrPrev("next"));
    nav.setActionHandler("seekto", (e) => {
      if (e.seekTime) player.setSeek(e.seekTime * 1000);
    });
  }

  /**
   * 更新元数据
   * @param song 当前歌曲
   */
  public updateMetadata(song: Track) {
    if (!supported()) return;
    navigator.mediaSession.metadata = new window.MediaMetadata({
      title: song.title,
      artist: song.artists.map((a) => a.name).join("/"),
      album: song.album?.title ?? "",
      artwork: this.buildArtwork(song),
    });
  }

  /** 清除元数据 */
  public clear() {
    if (!supported()) return;
    navigator.mediaSession.metadata = null;
    navigator.mediaSession.playbackState = "none";
  }

  /**
   * 构建专辑封面数组
   */
  private buildArtwork(song: Track) {
    const sizes = [
      ["s", "100x100"],
      ["m", "300x300"],
      ["l", "1024x1024"],
      ["xl", "1920x1920"],
    ] as const;
    return sizes.map(([size, sizes]) => ({
      src: new URL(coverUrl(song.cover, size), location.href).href,
      sizes,
      type: "image/jpeg",
    }));
  }

  /**
   * 更新播放进度
   * @param duration 总时长
   * @param position 当前位置
   * @param immediate 是否立即发送，用于 Seek 操作
   */
  public updateState(
    duration: number,
    position: number,
    immediate: boolean = false,
  ) {
    if (immediate) this.throttledUpdatePositionState.cancel();
    this.throttledUpdatePositionState(duration, position);
  }

  /**
   * 更新播放状态
   */
  public updatePlaybackStatus(isPlaying: boolean) {
    if (!supported()) return;
    navigator.mediaSession.playbackState = isPlaying ? "playing" : "paused";
  }

  /**
   * 限流更新进度状态
   */
  private throttledUpdatePositionState = throttle(
    (duration: number, position: number) => {
      if (!supported() || !(duration > 0)) return;
      navigator.mediaSession.setPositionState({
        duration: msToS(duration),
        position: Math.min(msToS(position), msToS(duration)),
        playbackRate: this.currentRate,
      });
    },
    1000,
  );
}

export const mediaSessionManager = new MediaSessionManager();
