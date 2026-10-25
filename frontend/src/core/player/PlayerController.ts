import { fixturePlayer, source } from "@/api/source";
import type { Track } from "@/api/types";
import {
  useDataStore,
  useMusicStore,
  useSettingStore,
  useStatusStore,
} from "@/stores";
import type { RepeatModeType, ShuffleModeType } from "@/types/play-mode";
import { calculateLyricIndex } from "@/utils/calc";
import { getCoverColor } from "@/utils/color";
import { calculateProgress } from "@/utils/time";
import type { PlayerState } from "@/viewmodels";
import { useLyricManager } from "./LyricManager";

/**
 * 播放控制器。方法名照搬 SPlayer core/player/PlayerController.ts，供照搬的播放器组件调用。
 *
 * 阶段 3.5 只有展示层：控制器改变播放状态与界面，不播放音频，播放时间也不前进。
 * 音频、队列与位置的保存、播放失败跳过等在阶段 4.4 接入。
 */
class PlayerController {
  /**
   * 载入播放器状态：队列、当前曲目、播放位置与是否在播放
   * @param state 播放器状态
   */
  public async loadState(state: PlayerState) {
    const dataStore = useDataStore();
    const statusStore = useStatusStore();
    dataStore.playList = state.queue;
    statusStore.playIndex = state.queue.length ? state.index : -1;
    const song = state.queue[state.index];
    if (!song) {
      useMusicStore().resetMusicData();
      useLyricManager().resetSongLyric();
      statusStore.$patch({
        playStatus: false,
        playLoading: false,
        currentTime: 0,
        duration: 0,
      });
      return;
    }
    await this.setupSong(song, state.positionMs, state.durationMs);
    statusStore.playStatus = state.playing;
  }

  /**
   * 设置当前歌曲的界面：歌曲信息、进度、歌词与封面取色
   * @param song 歌曲
   * @param startSeek 开始位置（毫秒）
   * @param duration 音频时长（毫秒），默认为歌曲的时长
   */
  private async setupSong(
    song: Track,
    startSeek: number,
    duration = song.durationMs,
  ) {
    const musicStore = useMusicStore();
    const statusStore = useStatusStore();
    musicStore.playSong = song;
    statusStore.$patch({
      currentTime: startSeek,
      duration,
      progress: calculateProgress(startSeek, duration),
      lyricIndex: -1,
      lyricLoading: true,
      playLoading: false,
    });
    getCoverColor(musicStore.songCover);
    await useLyricManager().handleLyric(song);
    this.updateLyricIndex();
  }

  /** 按当前时间计算歌词索引 */
  private updateLyricIndex() {
    const musicStore = useMusicStore();
    const statusStore = useStatusStore();
    const settingStore = useSettingStore();
    const offset = statusStore.getSongOffset(musicStore.playSong?.id);
    const useYrc = !!(
      settingStore.showWordLyrics && musicStore.songLyric.yrcData?.length
    );
    const rawLyrics = toRaw(
      useYrc ? musicStore.songLyric.yrcData : musicStore.songLyric.lrcData,
    );
    statusStore.lyricIndex = calculateLyricIndex(
      statusStore.currentTime,
      rawLyrics,
      offset,
    );
  }

  /** 播放 */
  async play() {
    useStatusStore().playStatus = true;
  }

  /** 暂停 */
  async pause() {
    useStatusStore().playStatus = false;
  }

  /** 播放或暂停 */
  async playOrPause() {
    const statusStore = useStatusStore();
    if (statusStore.playStatus) await this.pause();
    else await this.play();
  }

  /**
   * 切歌：上一首/下一首
   * @param type 方向
   */
  public async nextOrPrev(type: "next" | "prev" = "next") {
    const dataStore = useDataStore();
    const statusStore = useStatusStore();
    const playListLength = dataStore.playList.length;
    if (playListLength === 0) {
      window.$message.error("播放列表为空，请添加歌曲");
      return;
    }
    let nextIndex = statusStore.playIndex + (type === "next" ? 1 : -1);
    if (nextIndex >= playListLength) nextIndex = 0;
    if (nextIndex < 0) nextIndex = playListLength - 1;
    await this.togglePlayIndex(nextIndex, true);
  }

  /**
   * 切换到播放列表中的指定歌曲
   * @param index 歌曲索引
   * @param play 是否播放
   */
  public async togglePlayIndex(index: number, play: boolean = false) {
    const dataStore = useDataStore();
    const statusStore = useStatusStore();
    const song = dataStore.playList[index];
    if (!song) return;
    if (statusStore.playIndex === index) {
      if (play) await this.play();
      return;
    }
    statusStore.playIndex = index;
    await this.setupSong(song, 0);
    if (play) await this.play();
  }

  /** 获取总时长 (ms) */
  public getDuration(): number {
    return useStatusStore().duration;
  }

  /** 获取当前播放位置 (ms) */
  public getSeek(): number {
    return useStatusStore().currentTime;
  }

  /**
   * 设置播放进度
   * @param time 播放位置（毫秒）
   */
  public setSeek(time: number) {
    const statusStore = useStatusStore();
    const safeTime = Math.max(0, Math.min(time, this.getDuration()));
    statusStore.currentTime = safeTime;
    statusStore.progress = calculateProgress(safeTime, statusStore.duration);
    this.updateLyricIndex();
  }

  /**
   * 设置音量
   * @param actions 音量值或滚动事件
   */
  public setVolume(actions: number | "up" | "down" | WheelEvent) {
    const statusStore = useStatusStore();
    // 增量
    const increment = 0.05;
    // 直接设置音量
    if (typeof actions === "number") {
      actions = Math.max(0, Math.min(actions, 1));
      statusStore.playVolume = actions;
    }
    // 音量加减
    else if (actions === "up" || actions === "down") {
      statusStore.playVolume = Math.max(
        0,
        Math.min(
          statusStore.playVolume + (actions === "up" ? increment : -increment),
          1,
        ),
      );
    }
    // 滚动事件
    else {
      const deltaY = actions.deltaY;
      const volumeChange = deltaY > 0 ? -increment : increment;
      statusStore.playVolume = Math.max(
        0,
        Math.min(statusStore.playVolume + volumeChange, 1),
      );
    }
  }

  /** 切换静音 */
  public toggleMute() {
    const statusStore = useStatusStore();
    // 是否静音
    const isMuted = statusStore.playVolume === 0;
    if (isMuted) {
      statusStore.playVolume = statusStore.playVolumeMute;
    } else {
      statusStore.playVolumeMute = statusStore.playVolume;
      statusStore.playVolume = 0;
    }
  }

  /**
   * 切换循环模式
   * @param mode 可选，直接设置目标模式。不传则按 List -> One -> Off 顺序轮转
   */
  public toggleRepeat(mode?: RepeatModeType) {
    const statusStore = useStatusStore();
    if (mode) statusStore.repeatMode = mode;
    else statusStore.toggleRepeat();
  }

  /**
   * 切换随机模式
   * @param mode 可选，直接设置目标模式。不传则在 Off 与 On 之间切换
   */
  public async toggleShuffle(mode?: ShuffleModeType) {
    const statusStore = useStatusStore();
    if (mode) statusStore.shuffleMode = mode;
    else statusStore.toggleShuffle();
  }
}

let instance: PlayerController | null = null;

export const usePlayerController = (): PlayerController => {
  if (!instance) instance = new PlayerController();
  return instance;
};

/**
 * 启动时载入播放器：开发时由 fixture 指定状态（`?player=`），并载入喜爱的歌曲。
 * 浏览器中保存的队列与位置在阶段 4.4 恢复。
 */
export async function initPlayer() {
  const dataStore = useDataStore();
  const loadLikes = source.likeIds().then(
    (ids) => (dataStore.userLikeData.songs = ids),
    (error: unknown) => console.error("加载喜爱的歌曲失败", error),
  );
  if (fixturePlayer) await usePlayerController().loadState(fixturePlayer);
  await loadLikes;
}
