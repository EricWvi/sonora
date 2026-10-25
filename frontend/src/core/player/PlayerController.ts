import { fixturePlayer, source, type DataSource } from "@/api/source";
import type { Track } from "@/api/types";
import {
  AudioErrorCode,
  type IPlaybackEngine,
} from "@/core/audio-player/IPlaybackEngine";
import {
  useDataStore,
  useMusicStore,
  useSettingStore,
  useStatusStore,
} from "@/stores";
import type { RepeatModeType, ShuffleModeType } from "@/types/play-mode";
import { calculateLyricIndex } from "@/utils/calc";
import { calculateProgress } from "@/utils/time";
import type { PlayerState } from "@/viewmodels";
import { debounce, throttle, type DebouncedFunc } from "lodash-es";
import { browserDeps } from "./browserDeps";
import { mediaSessionManager } from "./MediaSessionManager";
import {
  indexAfterMove,
  indexAfterRemove,
  insertAfter,
  newEntry,
  reserveEntryIds,
  shuffled,
  without,
  type QueueEntry,
} from "./queue";
import type { PlayerStorage, SavedQueue } from "./storage";
import { keepAwake } from "./wakeLock";

/**
 * 播放状态（路线图 5.2）：
 *
 * - `idle`：队列为空；
 * - `resolving`：取当前条目的音频地址；
 * - `loading`：音频已交给引擎，等待可以播放；
 * - `playing` / `paused`；
 * - `ended`：不循环时播完最后一首，停在末尾；
 * - `error`：播放失败，随后自动跳过或停止。
 *
 * 从浏览器恢复的队列处于 `paused`，但还没有加载音频，按播放时才加载。
 */
export type PlaybackState =
  "idle" | "resolving" | "loading" | "playing" | "paused" | "ended" | "error";

/** 播放控制器用到的外部能力；单元测试换成假的实现 */
export interface PlaybackDeps {
  engine: IPlaybackEngine;
  /**
   * 解析播放地址。Sonora 的地址固定为 `DataSource.audioUrl`，留作异步是为了让
   * 过期的解析可以被丢弃，也便于测试切歌竞态
   */
  resolveAudio(track: Track): Promise<string>;
  source: Pick<DataSource, "recordPlay" | "audioStatus">;
  storage: PlayerStorage;
  message: Pick<Window["$message"], "success" | "error" | "warning">;
  /** 切换歌曲时更新歌词、封面取色与系统媒体信息，歌词载入后完成 */
  setupSong(song: Track): Promise<void>;
  /** 清空队列时清除歌词与系统媒体信息 */
  resetSong(): void;
  /** 会话失效，回到登录页 */
  unauthorized(): void;
}

/** 连续失败多少首后停止（阶段 2 的“播放行为”） */
const MAX_FAIL_SKIP = 3;
/** 播放位置的保存间隔（毫秒） */
const POSITION_SAVE_INTERVAL = 5000;

const sleep = (ms: number) => new Promise((r) => setTimeout(r, ms));

/**
 * 播放控制器，也是路线图中的 Playback：播放状态、队列与播放模式的唯一所有者，
 * 组件只读 store、只调用这里的方法。方法名与流程照搬 SPlayer core/player/PlayerController.ts
 * 与 PlayModeManager.ts，照搬的组件因此不用改。
 *
 * 只用 Web Audio（HTMLAudioElement 经 MediaElementSource），不迁入 MPV、FFmpeg、自动混音、
 * 倍速、AB 循环、定时关闭、私人 FM 与心动模式。与 SPlayer 的差异：
 *
 * - 队列条目有独立身份，允许重复入队（见 queue.ts）；随机时当前曲目放在第一位，
 *   增删同时作用于原始顺序；
 * - 不循环时播完最后一首就停止，SPlayer 会绕回第一首；
 * - 播放失败不重试，直接跳过，连续 3 首失败或无歌可跳时停止并提示，不清空队列；
 * - 播放超过 min(30 秒, 时长的一半) 才记入最近播放，SPlayer 开始播放即记录；
 * - 解析或加载中暂停，加载完成后保持暂停。
 */
export class PlayerController {
  private state: PlaybackState = "idle";
  /** 用户是否想要播放；解析与加载完成后据此决定是否开始播放 */
  private wantPlay = false;
  /** 当前加载的音频已可以播放 */
  private ready = false;
  /** 播放中缓冲 */
  private buffering = false;
  /** 当前播放请求标识；切歌、清空时递增，过期请求的结果被丢弃 */
  public currentRequestToken = 0;
  /** 开始播放的尝试序号；暂停时递增，丢弃之前尝试的结果 */
  private playAttempt = 0;
  /** 连续跳过计数 */
  private failSkipCount = 0;
  /** 本次播放已听的时长（毫秒），跳转不计入 */
  private listenedMs = 0;
  /** 上次进度更新时的位置；跳转后为 null */
  private lastTickMs: number | null = null;
  /** 本次播放已记入最近播放 */
  private recorded = false;
  /** 上次保存播放位置的时间 */
  private lastPositionSave = 0;
  /** 播放进度更新回调 */
  private onTimeUpdate: DebouncedFunc<() => void>;
  /** 保存队列，合并连续的修改 */
  private saveQueue: DebouncedFunc<() => void>;

  constructor(private deps: PlaybackDeps) {
    this.onTimeUpdate = throttle(() => this.handleTimeUpdate(), 200);
    this.saveQueue = debounce(() => {
      if (this.state === "idle") return;
      void this.deps.storage
        .saveQueue(this.snapshot())
        .catch((error) => console.error("保存播放队列失败", error));
    }, 300);
    this.bindAudioEvents();
  }

  /** 当前播放状态 */
  public get playbackState(): PlaybackState {
    return this.state;
  }

  private get engine() {
    return this.deps.engine;
  }

  /** 当前条目 */
  private currentEntry(): QueueEntry | undefined {
    const dataStore = useDataStore();
    return dataStore.playList[useStatusStore().playIndex];
  }

  /**
   * 进入新状态，并把播放状态写入 store。store 中的 playStatus、playLoading
   * 只在这里写入
   */
  private setState(state: PlaybackState) {
    this.state = state;
    if (state !== "playing") this.buffering = false;
    this.sync();
  }

  private sync() {
    const statusStore = useStatusStore();
    const settingStore = useSettingStore();
    const state = this.state;
    const active =
      state === "resolving" || state === "loading" || state === "playing";
    statusStore.$patch({
      playStatus: this.wantPlay && active,
      playLoading:
        state === "resolving" ||
        state === "loading" ||
        (state === "playing" && this.buffering),
    });
    mediaSessionManager.updatePlaybackStatus(statusStore.playStatus);
    keepAwake(settingStore.preventSleep && statusStore.playStatus);
  }

  /** 保存的队列 */
  private snapshot(): SavedQueue {
    const dataStore = useDataStore();
    return {
      playList: toRaw(dataStore.playList),
      originalPlayList: toRaw(dataStore.originalPlayList),
      playIndex: useStatusStore().playIndex,
      playPlaylistId: useMusicStore().playPlaylistId,
    };
  }

  /**
   * 保存当前条目的播放位置
   * @param positionMs 位置，默认为当前位置
   */
  public savePosition(positionMs = useStatusStore().currentTime) {
    const entry = this.currentEntry();
    if (!entry) return;
    this.lastPositionSave = Date.now();
    void this.deps.storage
      .savePosition({ entryId: entry.entryId, positionMs })
      .catch((error) => console.error("保存播放位置失败", error));
  }

  /** 立即写入尚未保存的队列 */
  public flushPersistence() {
    this.saveQueue.flush();
  }

  /**
   * 设置播放列表；随机模式下同时设置原始顺序
   * @param list 播放列表
   * @param original 原始顺序，不在随机模式时为空
   */
  private setLists(list: QueueEntry[], original: QueueEntry[] = []) {
    const dataStore = useDataStore();
    dataStore.setPlayList(list);
    dataStore.setOriginalPlayList(original);
    this.saveQueue();
  }

  /**
   * 载入播放器状态：队列、当前曲目、播放位置与是否在播放。用于 fixture（`?player=`）：
   * 不加载音频，“播放中”只是显示状态，播放时间停在指定的位置
   * @param state 播放器状态
   */
  public async loadState(state: PlayerState) {
    const statusStore = useStatusStore();
    const list = state.queue.map(newEntry);
    this.currentRequestToken++;
    this.engine.stop();
    useDataStore().setPlayList(list);
    useDataStore().setOriginalPlayList([]);
    statusStore.playIndex = list.length ? state.index : -1;
    const song = list[state.index];
    if (!song) {
      this.resetPlayer();
      return;
    }
    await this.setupSongUI(song, state.positionMs, state.durationMs);
    this.wantPlay = state.playing;
    this.setState(state.playing ? "playing" : "paused");
  }

  /**
   * 恢复浏览器中保存的队列与位置，处于暂停，不加载音频
   * @returns 是否有保存的队列
   */
  public async restore(): Promise<boolean> {
    const saved = await this.deps.storage.load();
    if (!saved || !saved.queue.playList.length) return false;
    const { queue, position } = saved;
    const settingStore = useSettingStore();
    const statusStore = useStatusStore();
    reserveEntryIds(queue.playList);
    reserveEntryIds(queue.originalPlayList);
    const index = Math.min(
      Math.max(queue.playIndex, 0),
      queue.playList.length - 1,
    );
    this.currentRequestToken++;
    useDataStore().setPlayList(queue.playList);
    useDataStore().setOriginalPlayList(
      statusStore.shuffleMode === "on" ? queue.originalPlayList : [],
    );
    statusStore.playIndex = index;
    useMusicStore().playPlaylistId = queue.playPlaylistId;
    const entry = queue.playList[index];
    const positionMs =
      settingStore.memoryLastSeek && position?.entryId === entry.entryId
        ? position.positionMs
        : 0;
    this.wantPlay = false;
    this.setState("paused");
    await this.setupSongUI(entry, positionMs);
    return true;
  }

  /**
   * 设置歌曲 UI 状态
   * @param song 歌曲
   * @param startSeek 开始位置（毫秒）
   * @param duration 音频时长（毫秒），默认为歌曲的时长
   * @returns 歌词载入后完成
   */
  private setupSongUI(
    song: Track,
    startSeek: number,
    duration = song.durationMs,
  ): Promise<void> {
    const musicStore = useMusicStore();
    const statusStore = useStatusStore();
    musicStore.playSong = song;
    statusStore.$patch({
      currentTime: startSeek,
      duration,
      progress: calculateProgress(startSeek, duration),
      lyricIndex: -1,
      lyricLoading: true,
    });
    const token = this.currentRequestToken;
    return this.deps.setupSong(song).then(() => {
      if (token === this.currentRequestToken) this.updateLyricIndex();
    });
  }

  /** 按当前时间计算歌词索引 */
  private updateLyricIndex() {
    const statusStore = useStatusStore();
    statusStore.lyricIndex = this.lyricIndexAt(statusStore.currentTime);
  }

  /**
   * 指定位置的歌词索引
   * @param currentTime 播放位置（毫秒）
   */
  private lyricIndexAt(currentTime: number) {
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
    return calculateLyricIndex(currentTime, rawLyrics, offset);
  }

  /** 重新开始统计本次播放的收听时长 */
  private resetListening() {
    this.listenedMs = 0;
    this.lastTickMs = null;
    this.recorded = false;
  }

  /**
   * 初始化并播放当前条目
   * @param options 配置
   * @param options.autoPlay 是否自动播放
   * @param options.seek 初始播放进度（毫秒）
   */
  public async playSong(
    options: { autoPlay?: boolean; seek?: number } = {
      autoPlay: true,
      seek: 0,
    },
  ) {
    const statusStore = useStatusStore();
    // 生成新的请求标识
    const requestToken = ++this.currentRequestToken;
    const { autoPlay = true, seek = 0 } = options;
    const entry = this.currentEntry();
    // 立即停止当前播放
    this.engine.stop();
    this.playAttempt++;
    this.ready = false;
    if (!entry) {
      this.wantPlay = false;
      this.setState("idle");
      return;
    }
    this.wantPlay = autoPlay;
    this.resetListening();
    this.setState("resolving");
    // 立即更新 UI（歌曲信息、封面、歌词等），无需等待网络请求
    void this.setupSongUI(entry, seek);
    this.savePosition(seek);
    let url: string;
    try {
      url = await this.deps.resolveAudio(entry);
      if (!url) throw new Error("AUDIO_SOURCE_EMPTY");
    } catch (error) {
      if (requestToken !== this.currentRequestToken) return;
      console.error("❌ 解析播放地址失败:", error);
      return this.handlePlaybackError(requestToken, undefined);
    }
    // 检查请求是否过期
    if (requestToken !== this.currentRequestToken) return;
    this.setState("loading");
    this.engine.setVolume(statusStore.playVolume);
    // 应用输出设备：音频图谱建立后输出走 AudioContext
    if (useSettingStore().playDevice !== "default")
      void this.toggleOutputDevice().catch((error) =>
        console.warn("切换输出设备失败", error),
      );
    try {
      // 解析期间可能拖动了进度
      await this.engine.play(url, {
        autoPlay: false,
        seek: statusStore.currentTime / 1000,
      });
    } catch (error) {
      if (requestToken !== this.currentRequestToken) return;
      console.error("❌ 加载音频失败:", error);
      return this.handlePlaybackError(requestToken, undefined);
    }
  }

  /** 音频可以播放：按用户的意图开始播放或保持暂停 */
  private onReady() {
    const statusStore = useStatusStore();
    this.ready = true;
    const duration = Math.floor(this.engine.duration * 1000);
    if (duration > 0) {
      statusStore.duration = duration;
      statusStore.progress = calculateProgress(
        statusStore.currentTime,
        duration,
      );
    }
    if (this.wantPlay) void this.startPlayback();
    else this.setState("paused");
  }

  /** 让已加载的音频开始播放 */
  private async startPlayback() {
    const settingStore = useSettingStore();
    const token = this.currentRequestToken;
    const attempt = ++this.playAttempt;
    const fadeTime = settingStore.getFadeTime
      ? settingStore.getFadeTime / 1000
      : 0;
    try {
      await this.engine.resume({ fadeIn: !!fadeTime, fadeDuration: fadeTime });
    } catch (error) {
      if (token !== this.currentRequestToken || attempt !== this.playAttempt)
        return;
      const name = error instanceof Error ? error.name : "";
      // 浏览器的自动播放限制：保持暂停，等用户点击
      if (name === "NotAllowedError") {
        this.wantPlay = false;
        this.setState("paused");
        return;
      }
      if (name === "AbortError") return;
      console.error("❌ 播放失败:", error);
      return this.handlePlaybackError(token, this.engine.getErrorCode());
    }
    if (token !== this.currentRequestToken || attempt !== this.playAttempt)
      return;
    this.setState("playing");
  }

  /**
   * 统一音频事件绑定
   */
  private bindAudioEvents() {
    const engine = this.engine;
    // 加载完成
    engine.addEventListener("canplay", () => {
      if (this.state === "loading" && !this.ready) this.onReady();
    });
    // 播放中缓冲
    engine.addEventListener("waiting", () => {
      if (this.state !== "playing") return;
      this.buffering = true;
      this.sync();
    });
    engine.addEventListener("playing", () => {
      if (!this.buffering) return;
      this.buffering = false;
      this.sync();
    });
    // 播放结束
    engine.addEventListener("ended", () => {
      if (this.state !== "playing") return;
      console.log(`⏹️ [${useMusicStore().playSong?.id}] 歌曲结束`);
      this.onTimeUpdate.cancel();
      this.handleTimeUpdate();
      // 自动播放下一首
      void this.nextOrPrev("next", true, true);
    });
    // 进度更新
    engine.addEventListener("timeupdate", this.onTimeUpdate);
    // 错误处理
    engine.addEventListener("error", (e) => {
      if (
        this.state !== "loading" &&
        this.state !== "playing" &&
        this.state !== "paused"
      )
        return;
      void this.handlePlaybackError(
        this.currentRequestToken,
        e.detail.errorCode,
      );
    });
  }

  /** 进度更新：进度、歌词、系统媒体信息、最近播放与位置保存 */
  private handleTimeUpdate() {
    if (this.state !== "playing" && this.state !== "paused") return;
    const statusStore = useStatusStore();
    const currentTime = Math.floor(this.engine.currentTime * 1000);
    const duration =
      Math.floor(this.engine.duration * 1000) || statusStore.duration;
    // 更新状态
    statusStore.$patch({
      currentTime,
      duration,
      progress: calculateProgress(currentTime, duration),
      lyricIndex: this.lyricIndexAt(currentTime),
    });
    // 成功播放一段距离后，重置失败跳过计数
    if (currentTime > 500 && this.failSkipCount > 0) {
      this.failSkipCount = 0;
    }
    // 更新系统 MediaSession
    mediaSessionManager.updateState(duration, currentTime);
    this.countListening(currentTime, duration);
    if (Date.now() - this.lastPositionSave >= POSITION_SAVE_INTERVAL)
      this.savePosition(currentTime);
  }

  /**
   * 累计收听时长，超过 min(30 秒, 时长的一半) 时记入最近播放，每次播放只记一次
   * @param currentTime 当前位置（毫秒）
   * @param duration 时长（毫秒）
   */
  private countListening(currentTime: number, duration: number) {
    if (this.state === "playing" && this.lastTickMs !== null) {
      const delta = currentTime - this.lastTickMs;
      // 跳转产生的位移不算收听
      if (delta > 0 && delta < 2000) this.listenedMs += delta;
    }
    this.lastTickMs = currentTime;
    if (this.recorded || !(duration > 0)) return;
    if (this.listenedMs < Math.min(30_000, duration / 2)) return;
    const entry = this.currentEntry();
    if (!entry) return;
    this.recorded = true;
    this.deps.source
      .recordPlay(entry.id)
      .catch((error) => console.error("记录最近播放失败", error));
  }

  /**
   * 统一错误处理：查明原因并提示，自动跳到下一首；连续失败 3 首或没有可跳的歌时停止。
   * 未登录时回到登录页，不算失败
   * @param token 出错的播放请求
   * @param errCode 错误码
   */
  private async handlePlaybackError(
    token: number,
    errCode: number | undefined,
  ) {
    if (token !== this.currentRequestToken) return;
    // 用户主动中止
    if (
      errCode === AudioErrorCode.ABORTED ||
      errCode === AudioErrorCode.DOM_ABORT
    )
      return;
    const entry = this.currentEntry();
    if (!entry) return;
    const wasPlaying = this.wantPlay;
    this.wantPlay = false;
    this.playAttempt++;
    this.engine.stop();
    this.setState("error");
    const status = await this.deps.source.audioStatus(entry.id);
    if (token !== this.currentRequestToken) return;
    if (status === 401) {
      this.deps.unauthorized();
      return;
    }
    const reason = errorReason(status, errCode);
    console.error(`❌ [${entry.id}] 播放失败：${reason}`, { status, errCode });
    const dataStore = useDataStore();
    const statusStore = useStatusStore();
    this.failSkipCount++;
    // 连续跳过 3 首 -> 停止播放
    if (this.failSkipCount >= MAX_FAIL_SKIP) {
      this.failSkipCount = 0;
      this.deps.message.error(`${reason}，播放失败次数过多，已停止播放`);
      return;
    }
    // 列表只有一首 -> 停止播放
    if (dataStore.playList.length <= 1) {
      this.failSkipCount = 0;
      this.deps.message.error(`${reason}，当前已无可播放歌曲`);
      return;
    }
    const last = statusStore.playIndex >= dataStore.playList.length - 1;
    if (last && statusStore.repeatMode === "off") {
      this.failSkipCount = 0;
      this.deps.message.error(`${reason}，已停止播放`);
      return;
    }
    this.deps.message.error(`${reason}，已自动跳过`);
    // 添加延迟，避免快速切歌导致卡死
    await sleep(500);
    if (token !== this.currentRequestToken) return;
    statusStore.playIndex = last ? 0 : statusStore.playIndex + 1;
    this.saveQueue();
    await this.playSong({ autoPlay: wasPlaying });
  }

  /** 用户选择了歌曲，重新统计连续失败 */
  private resetFailures() {
    this.failSkipCount = 0;
  }

  /** 播放 */
  async play() {
    const statusStore = useStatusStore();
    // 如果已经在播放，直接返回
    if (this.state === "playing") return;
    if (!this.currentEntry()) return;
    // 解析或加载中：加载完成后开始播放
    if (
      this.state === "resolving" ||
      (this.state === "loading" && !this.ready)
    ) {
      this.wantPlay = true;
      this.sync();
      return;
    }
    // 如果没有源，尝试重新初始化当前歌曲
    if (!this.engine.src || this.state === "error" || this.state === "idle") {
      this.resetFailures();
      await this.playSong({ autoPlay: true, seek: statusStore.currentTime });
      return;
    }
    // 播完后重新开始，这是新的一次播放
    if (this.state === "ended") this.resetListening();
    this.wantPlay = true;
    this.sync();
    await this.startPlayback();
  }

  /** 暂停 */
  async pause() {
    const settingStore = useSettingStore();
    this.wantPlay = false;
    this.playAttempt++;
    if (this.state === "playing" || (this.state === "loading" && this.ready)) {
      // 计算渐出时间
      const fadeTime = settingStore.getFadeTime
        ? settingStore.getFadeTime / 1000
        : 0;
      this.engine.pause({ fadeOut: !!fadeTime, fadeDuration: fadeTime });
      this.setState("paused");
      this.savePosition(this.getSeek());
      return;
    }
    this.sync();
  }

  /** 播放/暂停切换 */
  async playOrPause() {
    if (this.wantPlay) await this.pause();
    else await this.play();
  }

  /**
   * 切歌：上一首/下一首
   * @param type 方向
   * @param play 是否立即播放
   * @param autoEnd 是否是自动结束触发的
   */
  public async nextOrPrev(
    type: "next" | "prev" = "next",
    play: boolean = true,
    autoEnd: boolean = false,
  ) {
    const dataStore = useDataStore();
    const statusStore = useStatusStore();
    // 播放列表是否为空
    const playListLength = dataStore.playList.length;
    if (playListLength === 0) {
      this.deps.message.error("播放列表为空，请添加歌曲");
      return;
    }
    if (!autoEnd) this.resetFailures();
    // 单曲循环
    // 如果是自动结束触发的单曲循环，则重播当前歌曲
    if (statusStore.repeatMode === "one" && autoEnd) {
      await this.playSong({ autoPlay: play, seek: 0 });
      return;
    }
    // 计算索引
    let nextIndex = statusStore.playIndex + (type === "next" ? 1 : -1);
    // 边界处理 (索引越界)
    if (nextIndex >= playListLength) {
      // 不循环：播完最后一首后停止
      if (autoEnd && statusStore.repeatMode === "off") {
        this.finish();
        return;
      }
      nextIndex = 0;
    }
    if (nextIndex < 0) nextIndex = playListLength - 1;
    // 更新状态并播放
    statusStore.playIndex = nextIndex;
    this.saveQueue();
    await this.playSong({ autoPlay: play });
  }

  /** 播完最后一首，停在末尾 */
  private finish() {
    const statusStore = useStatusStore();
    this.wantPlay = false;
    this.setState("ended");
    statusStore.currentTime = statusStore.duration;
    statusStore.progress = 100;
    // 下次打开时从头开始
    this.savePosition(0);
  }

  /** 获取总时长 (ms) */
  public getDuration(): number {
    const statusStore = useStatusStore();
    const duration = this.engine.duration;
    return duration > 0 ? Math.floor(duration * 1000) : statusStore.duration;
  }

  /** 获取当前播放位置 (ms) */
  public getSeek(): number {
    const statusStore = useStatusStore();
    if (this.state !== "playing" && this.state !== "paused")
      return statusStore.currentTime;
    const currentTime = this.engine.currentTime;
    return currentTime > 0
      ? Math.floor(currentTime * 1000)
      : statusStore.currentTime;
  }

  /**
   * 设置进度
   * @param time 时间 (ms)
   */
  public setSeek(time: number) {
    this.onTimeUpdate.cancel();
    const statusStore = useStatusStore();
    const safeTime = Math.max(0, Math.min(time, this.getDuration()));
    statusStore.currentTime = safeTime;
    statusStore.progress = calculateProgress(safeTime, statusStore.duration);
    this.updateLyricIndex();
    // 跳转的位移不算收听
    this.lastTickMs = null;
    // 解析中还没有音频，加载时从这里开始
    if (this.engine.src && this.state !== "resolving") {
      this.engine.seek(safeTime / 1000);
    }
    if (this.state === "ended") this.setState("paused");
    mediaSessionManager.updateState(this.getDuration(), safeTime, true);
    this.savePosition(safeTime);
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
    this.engine.setVolume(statusStore.playVolume);
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
    this.engine.setVolume(statusStore.playVolume);
  }

  /**
   * 更新播放列表并播放
   * @param data 歌曲列表
   * @param song 指定播放的歌曲，默认为第一首
   * @param pid 列表来源，如 `album-3`（见 music store 的 playPlaylistId）
   * @param options 配置项
   * @param options.showTip 是否显示提示
   * @param options.play 是否播放
   */
  public async updatePlayList(
    data: Track[],
    song?: Track,
    pid?: string,
    options: { showTip?: boolean; play?: boolean } = {
      showTip: true,
      play: true,
    },
  ) {
    const statusStore = useStatusStore();
    const musicStore = useMusicStore();
    if (!data || !data.length) return;
    this.resetFailures();
    const entries = data.map(newEntry);
    const songIndex = song ? data.findIndex((s) => s.id === song.id) : -1;
    const start = entries[Math.max(songIndex, 0)];
    // 处理随机模式
    if (statusStore.shuffleMode === "on") {
      this.setLists(shuffled(entries, start), entries);
    } else {
      this.setLists(entries);
    }
    const newIndex = useDataStore().playList.indexOf(start);
    musicStore.playPlaylistId = pid ?? "";
    if (song && musicStore.playSong.id === song.id && this.state !== "idle") {
      // 如果是同一首歌，仅更新索引
      statusStore.playIndex = newIndex;
      this.saveQueue();
      // 如果需要播放
      if (options.play) await this.play();
    } else {
      statusStore.playIndex = newIndex;
      this.saveQueue();
      await this.playSong({ autoPlay: options.play });
    }
    if (options.showTip) this.deps.message.success("已开始播放");
  }

  /**
   * 清空播放列表
   */
  public async cleanPlayList() {
    const statusStore = useStatusStore();
    const musicStore = useMusicStore();
    this.currentRequestToken++;
    this.playAttempt++;
    this.saveQueue.cancel();
    // 重置状态
    this.engine.stop();
    this.wantPlay = false;
    this.ready = false;
    this.setState("idle");
    statusStore.resetPlayStatus();
    musicStore.resetMusicData();
    this.deps.resetSong();
    // 清空播放列表
    useDataStore().setPlayList([]);
    useDataStore().setOriginalPlayList([]);
    await this.deps.storage.clear();
  }

  /** 没有当前曲目时的界面状态 */
  private resetPlayer() {
    const statusStore = useStatusStore();
    useMusicStore().resetMusicData();
    this.deps.resetSong();
    this.wantPlay = false;
    this.setState("idle");
    statusStore.$patch({ currentTime: 0, duration: 0, progress: 0 });
  }

  /**
   * 添加下一首歌曲。队列中已有这首歌时仍插入一个新条目
   * @param song 歌曲
   * @param play 是否立即播放
   */
  public async addNextSong(song: Track, play: boolean = false) {
    const dataStore = useDataStore();
    const musicStore = useMusicStore();
    const current = this.currentEntry();
    if (current && musicStore.playSong.id === song.id) {
      await this.play();
      this.deps.message.success("已开始播放");
      return;
    }
    const entry = newEntry(song);
    const { list, index } = insertAfter(dataStore.playList, current, entry);
    const original = dataStore.originalPlayList.length
      ? insertAfter(dataStore.originalPlayList, current, entry).list
      : [];
    this.setLists(list, original);
    // 当前曲目之后插入，当前播放索引不变
    if (play) {
      await this.togglePlayIndex(index, true);
    } else {
      this.deps.message.success("已添加至下一首播放");
    }
  }

  /**
   * 切换播放索引
   * @param index 播放索引
   * @param play 是否立即播放
   */
  public async togglePlayIndex(index: number, play: boolean = false) {
    const dataStore = useDataStore();
    const statusStore = useStatusStore();
    // 若超出播放列表
    if (index < 0 || index >= dataStore.playList.length) return;
    this.resetFailures();
    // 相同歌曲且需要播放
    if (statusStore.playIndex === index && this.state !== "idle") {
      if (play) await this.play();
      return;
    }
    // 更改状态
    statusStore.playIndex = index;
    this.saveQueue();
    await this.playSong({ autoPlay: play });
  }

  /**
   * 移除指定歌曲
   * @param index 歌曲索引
   */
  public removeSongIndex(index: number) {
    const dataStore = useDataStore();
    const statusStore = useStatusStore();
    // 获取数据
    const { playList } = dataStore;
    // 若超出播放列表
    if (index < 0 || index >= playList.length) return;
    // 仅剩一首
    if (playList.length === 1) {
      void this.cleanPlayList();
      return;
    }
    const entry = playList[index];
    // 是否为当前播放歌曲
    const isCurrentPlay = statusStore.playIndex === index;
    statusStore.playIndex = indexAfterRemove(
      statusStore.playIndex,
      index,
      playList.length,
    );
    this.setLists(
      without(playList, entry),
      dataStore.originalPlayList.length
        ? without(dataStore.originalPlayList, entry)
        : [],
    );
    // 若为当前播放
    if (isCurrentPlay) {
      void this.playSong({ autoPlay: this.wantPlay });
    }
  }

  /**
   * 移动歌曲
   * @param fromIndex 移动前索引
   * @param toIndex 移动后索引
   */
  public async moveSong(fromIndex: number, toIndex: number) {
    const dataStore = useDataStore();
    const statusStore = useStatusStore();
    // 若索引相同
    if (fromIndex === toIndex) return;
    // 若索引超出播放列表
    if (fromIndex < 0 || fromIndex >= dataStore.playList.length) return;
    if (toIndex < 0 || toIndex >= dataStore.playList.length) return;
    // 复制播放列表
    const list = [...dataStore.playList];
    const [movedSong] = list.splice(fromIndex, 1);
    list.splice(toIndex, 0, movedSong);
    // 更新播放索引
    statusStore.playIndex = indexAfterMove(
      statusStore.playIndex,
      fromIndex,
      toIndex,
    );
    // 更新播放列表；原始顺序不变，关闭随机后恢复
    this.setLists(list, toRaw(dataStore.originalPlayList));
  }

  /**
   * 切换循环模式
   * @param mode 可选，直接设置目标模式。如果不传，则按 List -> One -> Off 顺序轮转
   */
  public toggleRepeat(mode?: RepeatModeType) {
    const statusStore = useStatusStore();
    if (mode) {
      if (statusStore.repeatMode === mode) return;
      statusStore.repeatMode = mode;
    } else {
      statusStore.toggleRepeat();
    }
    const modeText: Record<RepeatModeType, string> = {
      list: "列表循环",
      one: "单曲循环",
      off: "不循环",
    };
    this.deps.message.success(modeText[statusStore.repeatMode], {
      showIcon: false,
    });
  }

  /**
   * 切换随机模式：开启时打乱队列，当前曲目放在第一位；关闭时恢复原始顺序
   * @param mode 可选，直接设置目标模式。如果不传则按 Off -> On -> Off 顺序轮转
   */
  public async toggleShuffle(mode?: ShuffleModeType) {
    const dataStore = useDataStore();
    const statusStore = useStatusStore();
    const nextMode = mode ?? (statusStore.shuffleMode === "off" ? "on" : "off");
    if (nextMode === statusStore.shuffleMode) return;
    statusStore.shuffleMode = nextMode;
    const current = this.currentEntry();
    const list = toRaw(dataStore.playList);
    if (nextMode === "on") {
      this.setLists(shuffled(list, current), list);
      if (current) statusStore.playIndex = 0;
      this.deps.message.success("随机播放已开启", { showIcon: false });
    } else {
      const original = toRaw(dataStore.originalPlayList);
      const restored = original.length ? original : list;
      this.setLists(restored);
      if (current) {
        const idx = restored.findIndex((e) => e.entryId === current.entryId);
        statusStore.playIndex = idx !== -1 ? idx : 0;
      }
      this.deps.message.success("随机播放已关闭", { showIcon: false });
    }
    this.saveQueue();
  }

  /**
   * 获取频谱数据
   */
  public getSpectrumData(): Uint8Array | null {
    return this.engine.getFrequencyData();
  }

  /**
   * 获取低频音量 [0.0-1.0]
   * 用于驱动背景动画等视觉效果
   */
  public getLowFrequencyVolume(): number {
    return this.engine.getLowFrequencyVolume();
  }

  /**
   * 切换输出设备
   * @param deviceId 设备 ID
   */
  public async toggleOutputDevice(deviceId?: string) {
    const settingStore = useSettingStore();
    const device = deviceId ?? settingStore.playDevice;
    await this.engine.setSinkId(device);
  }

  /**
   * 更新均衡器
   * @param options 均衡器选项
   * @param options.bands 频带增益
   */
  public updateEq(options?: { bands?: number[] }) {
    if (options?.bands) {
      options.bands.forEach((val, idx) => this.engine.setFilterGain(idx, val));
    }
  }

  /**
   * 禁用均衡器
   */
  public disableEq() {
    for (let i = 0; i < 10; i++) this.engine.setFilterGain(i, 0);
  }
}

/**
 * 播放失败的原因
 * @param status 音频地址的 HTTP 状态，0 为服务不可达
 * @param errCode 音频元素的错误码
 */
function errorReason(status: number, errCode: number | undefined): string {
  if (status === 0) return "无法连接服务器";
  if (status === 404) return "歌曲已被删除";
  if (status >= 400) return `播放失败（${status}）`;
  switch (errCode) {
    case AudioErrorCode.SRC_NOT_SUPPORTED:
      return "浏览器不支持这首歌的格式";
    case AudioErrorCode.DECODE:
      return "音频无法解码";
    case AudioErrorCode.NETWORK:
      return "网络错误";
    default:
      return "播放失败";
  }
}

let instance: PlayerController | null = null;

/**
 * 获取 PlayerController 实例
 * @returns PlayerController
 */
export const usePlayerController = (): PlayerController => {
  instance ??= new PlayerController(browserDeps());
  return instance;
};

/**
 * 启动时载入播放器：恢复浏览器中保存的队列与位置（不自动播放，除非开启了“自动播放”），
 * 载入喜爱的歌曲，并应用均衡器与输出设备。开发时由 fixture 指定状态（`?player=`）。
 */
export async function initPlayer() {
  const player = usePlayerController();
  const dataStore = useDataStore();
  const statusStore = useStatusStore();
  const settingStore = useSettingStore();
  const loadLikes = source.likeIds().then(
    (ids) => (dataStore.userLikeData.songs = ids),
    (error: unknown) => console.error("加载喜爱的歌曲失败", error),
  );
  mediaSessionManager.init(player);
  // 均衡器
  watch(
    () => ({ enabled: statusStore.eqEnabled, bands: [...statusStore.eqBands] }),
    ({ enabled, bands }) => {
      if (enabled) player.updateEq({ bands });
      else player.disableEq();
    },
    { immediate: true },
  );
  // 输出设备
  watch(
    () => settingStore.playDevice,
    () =>
      player
        .toggleOutputDevice()
        .catch((error) => console.warn("切换输出设备失败", error)),
  );
  // 关闭页面前保存
  window.addEventListener("pagehide", () => {
    player.flushPersistence();
    player.savePosition(player.getSeek());
  });
  if (fixturePlayer) await player.loadState(fixturePlayer);
  else if ((await player.restore()) && settingStore.autoPlay)
    void player.play();
  await loadLikes;
}
