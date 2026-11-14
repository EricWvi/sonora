import { createPinia, setActivePinia } from "pinia";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import type { FmState, Track } from "@/api/types";
import {
  AudioErrorCode,
  type AudioEventMap,
  type IPlaybackEngine,
  type PlayOptions,
} from "@/core/audio-player/IPlaybackEngine";
import {
  useDataStore,
  useMusicStore,
  useSettingStore,
  useStatusStore,
} from "@/stores";
import { TypedEventTarget } from "@/utils/TypedEventTarget";
import { PlayerController, type PlaybackDeps } from "./PlayerController";
import { memoryStorage } from "./storage";

// 控制器模块经 store 与默认依赖间接导入数据源；测试不访问服务端。
vi.mock("@/api/source", () => ({
  source: { coverUrl: () => "", audioUrl: (id: number) => `/audio/${id}` },
  fixturePlayer: null,
  usingFixture: true,
  loginUrl: "/auth/login",
}));

/** 可控的假引擎：测试决定何时可以播放、何时出错与播放进度 */
class FakeEngine
  extends TypedEventTarget<AudioEventMap>
  implements IPlaybackEngine
{
  src = "";
  time = 0;
  length = 0;
  isPaused = true;
  /** 加载过的地址 */
  loads: string[] = [];
  /** 下一次 resume 抛出的错误 */
  resumeError: Error | null = null;
  volume = 1;
  rate = 1;

  replayGain = 1;

  setRate(rate: number) {
    this.rate = rate;
  }
  setReplayGain(gain: number) {
    this.replayGain = gain;
  }

  async play(url?: string, options: PlayOptions = {}) {
    if (url) {
      this.src = url;
      this.loads.push(url);
      this.time = options.seek ?? 0;
      this.length = 0;
    }
    if (options.autoPlay ?? true) await this.resume();
  }
  async resume() {
    if (this.resumeError) {
      const error = this.resumeError;
      this.resumeError = null;
      throw error;
    }
    this.isPaused = false;
    this.dispatch("play");
  }
  pause() {
    this.isPaused = true;
  }
  stop() {
    this.isPaused = true;
    this.src = "";
    this.time = 0;
  }
  seek(time: number) {
    this.time = time;
  }
  get duration() {
    return this.length;
  }
  get currentTime() {
    return this.time;
  }
  get paused() {
    return this.isPaused;
  }
  setVolume(value: number) {
    this.volume = value;
  }
  getVolume() {
    return this.volume;
  }
  async setSinkId() {}
  setAudioDelayCompensation() {}
  /** 预加载的地址 */
  preloaded: string | null = null;
  preload(url: string | null) {
    this.preloaded = url;
  }
  setFilterGain() {}
  getFrequencyData() {
    return new Uint8Array(0);
  }
  getLowFrequencyVolume() {
    return 0;
  }
  getErrorCode() {
    return 0;
  }

  /** 音频可以播放 */
  ready(seconds = 200) {
    this.length = seconds;
    this.dispatch("canplay");
  }
  /** 播放前进若干秒 */
  advance(seconds: number) {
    this.time += seconds;
    this.dispatch("timeupdate");
  }
  /** 播放到结尾 */
  end() {
    this.time = this.length;
    this.isPaused = true;
    this.dispatch("timeupdate");
    this.dispatch("ended");
  }
  fail(errorCode: number) {
    this.dispatch("error", { errorCode });
  }
}

function track(id: number, durationMs = 200_000): Track {
  return {
    id,
    title: `歌曲 ${id}`,
    artists: [],
    album: null,
    trackNo: id,
    discNo: 1,
    year: 0,
    durationMs,
    format: "mp3",
    cover: null,
    hasLyrics: false,
  };
}

/** 解析由测试逐个完成，用来制造先点后返回的竞态 */
function controllableResolve() {
  const pending = new Map<number, (url: string) => void>();
  const resolveAudio = (t: Track) =>
    new Promise<string>((resolve) => pending.set(t.id, resolve));
  const finish = (id: number) => {
    const resolve = pending.get(id);
    if (!resolve) throw new Error(`歌曲 ${id} 没有在解析`);
    pending.delete(id);
    resolve(`/audio/${id}`);
  };
  return { resolveAudio, finish, pending };
}

let engine: FakeEngine;
let storage: ReturnType<typeof memoryStorage>;
let deps: PlaybackDeps & {
  message: {
    success: ReturnType<typeof vi.fn>;
    error: ReturnType<typeof vi.fn>;
  };
  source: {
    recordPlay: ReturnType<typeof vi.fn>;
    audioStatus: ReturnType<typeof vi.fn>;
  };
  unauthorized: ReturnType<typeof vi.fn>;
  setTitle: ReturnType<typeof vi.fn>;
};

function makeDeps(overrides: Partial<PlaybackDeps> = {}) {
  deps = {
    engine,
    resolveAudio: async (t) => `/audio/${t.id}`,
    source: {
      recordPlay: vi.fn(async () => {}),
      audioStatus: vi.fn(async () => 200),
    },
    storage,
    message: { success: vi.fn(), error: vi.fn(), warning: vi.fn() },
    setupSong: async () => {},
    resetSong: () => {},
    unauthorized: vi.fn(),
    setTitle: vi.fn(),
    ...overrides,
  } as typeof deps;
  return deps;
}

function newPlayer(overrides: Partial<PlaybackDeps> = {}) {
  return new PlayerController(makeDeps(overrides));
}

/** 等待挂起的 Promise 回调执行完 */
const settle = () => vi.advanceTimersByTimeAsync(0);

const ids = () => useDataStore().playList.map((e) => e.id);

beforeEach(() => {
  vi.useFakeTimers();
  setActivePinia(createPinia());
  engine = new FakeEngine();
  storage = memoryStorage();
});

afterEach(() => {
  vi.useRealTimers();
});

describe("切歌竞态", () => {
  it("先点 A 后点 B，A 较晚返回也不覆盖 B", async () => {
    const { resolveAudio, finish } = controllableResolve();
    const player = newPlayer({ resolveAudio });
    const list = [track(1), track(2)];
    void player.updatePlayList(list, list[0]);
    await settle();
    void player.togglePlayIndex(1, true);
    await settle();
    finish(2);
    await settle();
    finish(1);
    await settle();
    expect(engine.loads).toEqual(["/audio/2"]);
    expect(useMusicStore().playSong.id).toBe(2);
    expect(useStatusStore().playIndex).toBe(1);
    engine.ready();
    await settle();
    expect(player.playbackState).toBe("playing");
    expect(engine.paused).toBe(false);
  });

  it("加载中切歌后，旧音频的就绪不影响新歌", async () => {
    const player = newPlayer();
    void player.updatePlayList([track(1), track(2)]);
    await settle();
    expect(player.playbackState).toBe("loading");
    void player.nextOrPrev("next");
    await settle();
    expect(engine.loads).toEqual(["/audio/1", "/audio/2"]);
    engine.ready();
    await settle();
    expect(useMusicStore().playSong.id).toBe(2);
    expect(player.playbackState).toBe("playing");
  });

  it("A 的歌词较晚载入时不改写 B 的歌词索引", async () => {
    let finishLyricA: () => void = () => {};
    const player = newPlayer({
      setupSong: (song) =>
        song.id === 1
          ? new Promise<void>((r) => (finishLyricA = r))
          : Promise.resolve(),
    });
    void player.updatePlayList([track(1), track(2)]);
    await settle();
    void player.togglePlayIndex(1, true);
    await settle();
    useStatusStore().lyricIndex = 7;
    finishLyricA();
    await settle();
    expect(useStatusStore().lyricIndex).toBe(7);
  });
});

describe("解析与加载中暂停", () => {
  it("解析中暂停，解析完成后不自动播放", async () => {
    const { resolveAudio, finish } = controllableResolve();
    const player = newPlayer({ resolveAudio });
    void player.updatePlayList([track(1)]);
    await settle();
    expect(useStatusStore().playStatus).toBe(true);
    await player.pause();
    expect(useStatusStore().playStatus).toBe(false);
    finish(1);
    await settle();
    engine.ready();
    await settle();
    expect(player.playbackState).toBe("paused");
    expect(engine.paused).toBe(true);
    expect(useStatusStore().playLoading).toBe(false);
  });

  it("加载中暂停，就绪后保持暂停；再点播放才开始", async () => {
    const player = newPlayer();
    void player.updatePlayList([track(1)]);
    await settle();
    await player.playOrPause();
    engine.ready();
    await settle();
    expect(player.playbackState).toBe("paused");
    expect(engine.paused).toBe(true);
    await player.playOrPause();
    expect(player.playbackState).toBe("playing");
    expect(engine.paused).toBe(false);
  });

  it("加载中暂停后又播放，就绪后开始播放", async () => {
    const player = newPlayer();
    void player.updatePlayList([track(1)]);
    await settle();
    await player.pause();
    await player.play();
    engine.ready();
    await settle();
    expect(player.playbackState).toBe("playing");
  });

  it("解析中拖动进度，加载时从拖到的位置开始", async () => {
    const { resolveAudio, finish } = controllableResolve();
    const player = newPlayer({ resolveAudio });
    void player.updatePlayList([track(1)]);
    await settle();
    player.setSeek(42_000);
    finish(1);
    await settle();
    expect(engine.currentTime).toBe(42);
  });

  it("浏览器拒绝自动播放时停在暂停，不算失败", async () => {
    const player = newPlayer();
    engine.resumeError = new DOMException("blocked", "NotAllowedError");
    void player.updatePlayList([track(1)]);
    await settle();
    engine.ready();
    await settle();
    expect(player.playbackState).toBe("paused");
    expect(useStatusStore().playStatus).toBe(false);
    expect(deps.message.error).not.toHaveBeenCalled();
  });
});

describe("队列编辑", () => {
  async function playing(tracks: Track[], index = 0) {
    const player = newPlayer();
    void player.updatePlayList(tracks, tracks[index]);
    await settle();
    engine.ready();
    await settle();
    return player;
  }

  it("同一首歌可以多次入队，条目身份不同", async () => {
    const player = await playing([track(1), track(2)]);
    await player.addNextSong(track(2));
    await player.addNextSong(track(2));
    expect(ids()).toEqual([1, 2, 2, 2]);
    const entryIds = useDataStore().playList.map((e) => e.entryId);
    expect(new Set(entryIds).size).toBe(4);
    // 删除一个条目不影响同一首歌的其他条目
    player.removeSongIndex(1);
    expect(ids()).toEqual([1, 2, 2]);
    expect(useStatusStore().playIndex).toBe(0);
  });

  it("下一首播放正在播放的歌时直接播放，不插入", async () => {
    const player = await playing([track(1), track(2)]);
    await player.pause();
    await player.addNextSong(track(1));
    expect(ids()).toEqual([1, 2]);
    expect(player.playbackState).toBe("playing");
  });

  it("批量下一首播放按给定顺序插到当前曲目之后", async () => {
    const player = await playing([track(1), track(2), track(3)], 1);
    player.addNextSongs([track(7), track(8), track(9)]);
    expect(ids()).toEqual([1, 2, 7, 8, 9, 3]);
    expect(useStatusStore().playIndex).toBe(1);
    expect(player.playbackState).toBe("playing");
  });

  it("移除最后一首（不是当前曲目）时当前曲目不变", async () => {
    const player = await playing([track(1), track(2), track(3)], 1);
    player.removeSongIndex(2);
    expect(ids()).toEqual([1, 2]);
    expect(useStatusStore().playIndex).toBe(1);
    expect(engine.loads).toEqual(["/audio/2"]);
  });

  it("移除当前曲目时播放新的当前曲目，保持播放状态", async () => {
    const player = await playing([track(1), track(2), track(3)], 1);
    player.removeSongIndex(1);
    await settle();
    expect(useMusicStore().playSong.id).toBe(3);
    expect(useStatusStore().playStatus).toBe(true);
  });

  it("拖动后当前曲目跟随移动", async () => {
    const player = await playing([track(1), track(2), track(3), track(4)], 1);
    await player.moveSong(1, 3);
    expect(ids()).toEqual([1, 3, 4, 2]);
    expect(useStatusStore().playIndex).toBe(3);
    await player.moveSong(0, 3);
    expect(ids()).toEqual([3, 4, 2, 1]);
    expect(useStatusStore().playIndex).toBe(2);
  });

  it("清空队列后回到空闲", async () => {
    const player = await playing([track(1), track(2)]);
    await player.cleanPlayList();
    expect(player.playbackState).toBe("idle");
    expect(useDataStore().playList).toEqual([]);
    expect(storage.queue).toBeNull();
  });
});

describe("不可播放的歌曲", () => {
  const online = (id: number, playable: boolean): Track => ({
    ...track(id),
    source: "online",
    playable,
  });

  async function start(tracks: Track[], song?: Track) {
    const player = newPlayer();
    void player.updatePlayList(tracks, song);
    await settle();
    engine.ready();
    await settle();
    return player;
  }

  it("不进入队列，从第一首可播放的歌开始", async () => {
    await start([
      online(1, false),
      online(2, true),
      online(3, false),
      online(4, true),
    ]);
    expect(ids()).toEqual([2, 4]);
    expect(useMusicStore().playSong.id).toBe(2);
  });

  it("双击不可播放的歌只提示，不改变队列", async () => {
    const player = await start([track(1), track(2)]);
    const bad = online(9, false);
    await player.updatePlayList([bad, online(3, true)], bad);
    expect(deps.message.warning).toHaveBeenCalledTimes(1);
    expect(ids()).toEqual([1, 2]);
  });

  it("整个列表都不可播放时只提示", async () => {
    const player = await start([track(1)]);
    await player.updatePlayList([online(8, false), online(9, false)]);
    expect(deps.message.warning).toHaveBeenCalledTimes(1);
    expect(ids()).toEqual([1]);
  });

  it("下一首播放只加入可播放的歌", async () => {
    const player = await start([track(1)]);
    await player.addNextSong(online(9, false));
    player.addNextSongs([online(8, false), online(7, true)]);
    expect(ids()).toEqual([1, 7]);
    expect(deps.message.success).toHaveBeenLastCalledWith(
      "已添加 1 首至下一首播放",
    );
  });
});

describe("播放速度、AB 循环与自动关闭", () => {
  async function playing(tracks: Track[]) {
    const player = newPlayer();
    void player.updatePlayList(tracks);
    await settle();
    engine.ready();
    await settle();
    return player;
  }

  it("播放速度限制在范围内，换歌后保持", async () => {
    const player = await playing([track(1), track(2)]);
    player.setRate(5);
    expect(useStatusStore().playRate).toBe(2);
    expect(engine.rate).toBe(2);
    player.setRate(1.5);
    engine.rate = 1;
    void player.nextOrPrev("next");
    await settle();
    expect(engine.rate).toBe(1.5);
  });

  it("到达 B 点时回到 A 点，换歌后清除", async () => {
    const player = await playing([track(1), track(2)]);
    useStatusStore().abLoop = { enable: true, pointA: 10, pointB: 20 };
    engine.length = 200;
    engine.time = 19;
    engine.dispatch("timeupdate");
    await vi.advanceTimersByTimeAsync(300);
    expect(engine.time).toBe(19);
    engine.time = 20.1;
    engine.dispatch("timeupdate");
    await vi.advanceTimersByTimeAsync(300);
    expect(engine.time).toBe(10);
    void player.nextOrPrev("next");
    await settle();
    expect(useStatusStore().abLoop).toEqual({
      enable: false,
      pointA: null,
      pointB: null,
    });
  });

  it("B 点不大于 A 点时不循环", async () => {
    await playing([track(1)]);
    useStatusStore().abLoop = { enable: true, pointA: 20, pointB: 20 };
    engine.length = 200;
    engine.time = 50;
    engine.dispatch("timeupdate");
    await vi.advanceTimersByTimeAsync(300);
    expect(engine.time).toBe(50);
  });

  it("不等歌曲结束时，时间到就暂停", async () => {
    const player = await playing([track(1), track(2)]);
    useStatusStore().autoClose.waitSongEnd = false;
    player.startAutoCloseTimer(1, 2);
    await vi.advanceTimersByTimeAsync(3000);
    expect(player.playbackState).toBe("paused");
    expect(useStatusStore().autoClose.enable).toBe(false);
  });

  it("等歌曲结束时，播完这首才停，不切到下一首", async () => {
    const player = await playing([track(1), track(2)]);
    useStatusStore().autoClose.waitSongEnd = true;
    player.startAutoCloseTimer(1, 2);
    await vi.advanceTimersByTimeAsync(3000);
    expect(player.playbackState).toBe("playing");
    engine.dispatch("ended");
    await settle();
    expect(useMusicStore().playSong.id).toBe(1);
    expect(player.playbackState).toBe("ended");
    expect(useStatusStore().autoClose.enable).toBe(false);
  });

  it("关闭自动关闭后不再暂停（与 SPlayer 不同）", async () => {
    const player = await playing([track(1)]);
    useStatusStore().autoClose.waitSongEnd = false;
    player.startAutoCloseTimer(1, 60);
    player.stopAutoCloseTimer();
    await vi.advanceTimersByTimeAsync(120_000);
    expect(player.playbackState).toBe("playing");
    expect(useStatusStore().autoClose.remainTime).toBe(60);
  });
});

describe("播放模式", () => {
  async function playing(tracks: Track[], index = 0) {
    const player = newPlayer();
    void player.updatePlayList(tracks, tracks[index]);
    await settle();
    engine.ready();
    await settle();
    return player;
  }

  it("不循环时播完最后一首停在末尾", async () => {
    const player = await playing([track(1), track(2)], 1);
    engine.end();
    await settle();
    expect(player.playbackState).toBe("ended");
    expect(useStatusStore().playIndex).toBe(1);
    expect(engine.loads).toEqual(["/audio/2"]);
    // 手动下一首仍绕回第一首
    void player.nextOrPrev("next");
    await settle();
    expect(useMusicStore().playSong.id).toBe(1);
  });

  it("列表循环播完最后一首回到第一首", async () => {
    const player = await playing([track(1), track(2)], 1);
    player.toggleRepeat("list");
    engine.end();
    await settle();
    expect(useStatusStore().playIndex).toBe(0);
    expect(player.playbackState).toBe("loading");
  });

  it("单曲循环重播当前条目", async () => {
    const player = await playing([track(1), track(2)]);
    player.toggleRepeat("one");
    engine.end();
    await settle();
    expect(useStatusStore().playIndex).toBe(0);
    expect(engine.loads).toEqual(["/audio/1", "/audio/1"]);
  });

  it("随机时当前曲目放在第一位，关闭后恢复原顺序，期间的增删保留", async () => {
    const list = [1, 2, 3, 4, 5, 6].map((id) => track(id));
    const player = await playing(list, 2);
    await player.toggleShuffle();
    expect(useDataStore().playList[0].id).toBe(3);
    expect(useStatusStore().playIndex).toBe(0);
    expect([...ids()].sort()).toEqual([1, 2, 3, 4, 5, 6]);
    // 随机期间加入与删除
    await player.addNextSong(track(9));
    expect(ids()[1]).toBe(9);
    player.removeSongIndex(ids().indexOf(5));
    await player.toggleShuffle();
    expect(ids()).toEqual([1, 2, 3, 9, 4, 6]);
    expect(useStatusStore().playIndex).toBe(2);
    expect(useDataStore().originalPlayList).toEqual([]);
  });

  it("随机模式下播放列表，点击的歌放在第一位", async () => {
    const player = newPlayer();
    useStatusStore().shuffleMode = "on";
    const list = [1, 2, 3, 4].map((id) => track(id));
    void player.updatePlayList(list, list[2]);
    await settle();
    expect(ids()[0]).toBe(3);
    expect(useDataStore().originalPlayList.map((e) => e.id)).toEqual([
      1, 2, 3, 4,
    ]);
  });
});

describe("下一首歌曲预载", () => {
  async function playing(tracks: Track[], index = 0) {
    const player = newPlayer();
    void player.updatePlayList(tracks, tracks[index]);
    await settle();
    return player;
  }

  it("当前曲开始播放后才预加载下一首", async () => {
    useStatusStore().repeatMode = "off";
    await playing([track(1), track(2), track(3)]);
    expect(engine.preloaded).toBeNull();
    engine.ready();
    await settle();
    expect(engine.preloaded).toBe("/audio/2");
  });

  it("切到预加载的那首时保留，开始播放后预加载再下一首", async () => {
    useStatusStore().repeatMode = "off";
    const player = await playing([track(1), track(2), track(3)]);
    engine.ready();
    await settle();
    engine.end();
    await settle();
    // 队列变化时 initPlayer 会重新计算；这时当前曲还没开始，不释放
    player.preloadNext();
    expect(engine.loads).toEqual(["/audio/1", "/audio/2"]);
    expect(engine.preloaded).toBe("/audio/2");
    engine.ready();
    await settle();
    expect(engine.preloaded).toBe("/audio/3");
  });

  it("不循环时最后一首不预加载，列表循环时预加载第一首", async () => {
    useStatusStore().repeatMode = "off";
    const player = await playing([track(1), track(2)], 1);
    engine.ready();
    await settle();
    expect(engine.preloaded).toBeNull();
    useStatusStore().repeatMode = "list";
    player.preloadNext();
    await settle();
    expect(engine.preloaded).toBe("/audio/1");
  });

  it("单曲循环或关闭设置时释放", async () => {
    useStatusStore().repeatMode = "list";
    const player = await playing([track(1), track(2)]);
    engine.ready();
    await settle();
    expect(engine.preloaded).toBe("/audio/2");
    useStatusStore().repeatMode = "one";
    player.preloadNext();
    expect(engine.preloaded).toBeNull();
    useStatusStore().repeatMode = "list";
    player.preloadNext();
    await settle();
    expect(engine.preloaded).toBe("/audio/2");
    useSettingStore().useNextPrefetch = false;
    player.preloadNext();
    expect(engine.preloaded).toBeNull();
  });

  it("插入下一首后预加载新的下一首", async () => {
    useStatusStore().repeatMode = "off";
    const player = await playing([track(1), track(2)]);
    engine.ready();
    await settle();
    await player.addNextSong(track(5));
    player.preloadNext();
    await settle();
    expect(engine.preloaded).toBe("/audio/5");
  });
});

describe("播放失败", () => {
  async function failing(tracks: Track[], status = 404) {
    const player = newPlayer();
    deps.source.audioStatus.mockResolvedValue(status);
    void player.updatePlayList(tracks);
    await settle();
    return player;
  }

  it("失败后跳到下一首，连续 3 首失败后停止并提示", async () => {
    const player = await failing([1, 2, 3, 4, 5].map((id) => track(id)));
    for (let i = 0; i < 3; i++) {
      engine.fail(AudioErrorCode.SRC_NOT_SUPPORTED);
      await vi.advanceTimersByTimeAsync(600);
    }
    expect(engine.loads).toEqual(["/audio/1", "/audio/2", "/audio/3"]);
    expect(player.playbackState).toBe("error");
    expect(useStatusStore().playStatus).toBe(false);
    expect(deps.message.error).toHaveBeenLastCalledWith(
      "歌曲已被删除，播放失败次数过多，已停止播放",
    );
    // 队列保留
    expect(ids()).toEqual([1, 2, 3, 4, 5]);
  });

  it("播放一段后计数重置", async () => {
    await failing(
      [1, 2, 3, 4, 5].map((id) => track(id)),
      200,
    );
    engine.fail(AudioErrorCode.DECODE);
    await vi.advanceTimersByTimeAsync(600);
    engine.fail(AudioErrorCode.DECODE);
    await vi.advanceTimersByTimeAsync(600);
    engine.ready();
    await settle();
    engine.advance(1);
    engine.fail(AudioErrorCode.DECODE);
    await vi.advanceTimersByTimeAsync(600);
    expect(engine.loads).toHaveLength(4);
    expect(deps.message.error).toHaveBeenLastCalledWith(
      "音频无法解码，已自动跳过",
    );
  });

  it("只有一首时停止，不清空队列", async () => {
    const player = await failing([track(1)], 0);
    engine.fail(AudioErrorCode.NETWORK);
    await settle();
    expect(player.playbackState).toBe("error");
    expect(useStatusStore().playStatus).toBe(false);
    expect(ids()).toEqual([1]);
    expect(deps.message.error).toHaveBeenCalledWith(
      "无法连接服务器，当前已无可播放歌曲",
    );
    // 再点播放会重试
    await player.play();
    expect(engine.loads).toEqual(["/audio/1", "/audio/1"]);
  });

  it("在线歌曲的文件取不到时不归咎于格式", async () => {
    const online = (id: number): Track => ({ ...track(id), source: "online" });
    const player = newPlayer();
    void player.updatePlayList([online(1), online(2)]);
    await settle();
    engine.fail(AudioErrorCode.SRC_NOT_SUPPORTED);
    await settle();
    expect(deps.message.error).toHaveBeenLastCalledWith(
      "这首歌暂时无法播放，已自动跳过",
    );
    await vi.advanceTimersByTimeAsync(500);
    expect(useMusicStore().playSong.id).toBe(2);
  });

  it("未登录时回到登录页，不跳过", async () => {
    await failing([track(1), track(2)], 401);
    engine.fail(AudioErrorCode.SRC_NOT_SUPPORTED);
    await vi.advanceTimersByTimeAsync(600);
    expect(deps.unauthorized).toHaveBeenCalled();
    expect(engine.loads).toEqual(["/audio/1"]);
  });

  it("等待跳过时保持播放中；期间暂停，跳到下一首但不播放", async () => {
    const player = await failing([1, 2, 3].map((id) => track(id)));
    const statusStore = useStatusStore();
    engine.fail(AudioErrorCode.SRC_NOT_SUPPORTED);
    await settle();
    expect(player.playbackState).toBe("error");
    expect(statusStore.playStatus).toBe(true);
    expect(statusStore.playLoading).toBe(false);
    await player.playOrPause();
    expect(statusStore.playStatus).toBe(false);
    await vi.advanceTimersByTimeAsync(600);
    expect(engine.loads).toEqual(["/audio/1", "/audio/2"]);
    expect(statusStore.playIndex).toBe(1);
    engine.ready();
    await settle();
    expect(player.playbackState).toBe("paused");
    expect(engine.paused).toBe(true);
  });

  it("等待跳过时用户切了歌，不再自动跳过", async () => {
    const player = await failing([1, 2, 3].map((id) => track(id)));
    engine.fail(AudioErrorCode.SRC_NOT_SUPPORTED);
    await settle();
    void player.togglePlayIndex(2, true);
    await vi.advanceTimersByTimeAsync(600);
    expect(engine.loads).toEqual(["/audio/1", "/audio/3"]);
  });
});

describe("页面标题", () => {
  it("播放时为当前歌曲，暂停、加载与出错时恢复", async () => {
    const player = newPlayer();
    const song = {
      ...track(1),
      artists: [
        { id: 1, name: "甲" },
        { id: 2, name: "乙" },
      ],
    };
    void player.updatePlayList([song, track(2)]);
    await settle();
    expect(deps.setTitle).toHaveBeenLastCalledWith("Sonora");
    engine.ready();
    await settle();
    expect(deps.setTitle).toHaveBeenLastCalledWith("歌曲 1 - 甲/乙 | Sonora");
    await player.pause();
    expect(deps.setTitle).toHaveBeenLastCalledWith("Sonora");
    await player.play();
    expect(deps.setTitle).toHaveBeenLastCalledWith("歌曲 1 - 甲/乙 | Sonora");
    void player.nextOrPrev("next");
    await settle();
    expect(deps.setTitle).toHaveBeenLastCalledWith("Sonora");
    engine.ready();
    await settle();
    expect(deps.setTitle).toHaveBeenLastCalledWith("歌曲 2 -  | Sonora");
    engine.fail(AudioErrorCode.DECODE);
    await settle();
    expect(deps.setTitle).toHaveBeenLastCalledWith("Sonora");
  });
});

describe("最近播放", () => {
  it("收听超过 30 秒才记录，跳转不计入", async () => {
    const player = newPlayer();
    void player.updatePlayList([track(1, 300_000)]);
    await settle();
    engine.ready(300);
    await settle();
    // 跳到 200 秒不算收听
    player.setSeek(200_000);
    engine.time = 200;
    for (let i = 0; i < 29; i++) {
      engine.advance(1);
      await vi.advanceTimersByTimeAsync(250);
    }
    expect(deps.source.recordPlay).not.toHaveBeenCalled();
    for (let i = 0; i < 3; i++) {
      engine.advance(1);
      await vi.advanceTimersByTimeAsync(250);
    }
    expect(deps.source.recordPlay).toHaveBeenCalledTimes(1);
    expect(deps.source.recordPlay).toHaveBeenCalledWith(1);
  });

  it("短歌超过一半即记录，每次播放只记一次", async () => {
    const player = newPlayer();
    void player.updatePlayList([track(1, 40_000)]);
    await settle();
    engine.ready(40);
    await settle();
    for (let i = 0; i < 25; i++) {
      engine.advance(1);
      await vi.advanceTimersByTimeAsync(250);
    }
    expect(deps.source.recordPlay).toHaveBeenCalledTimes(1);
  });
});

describe("保存与恢复", () => {
  it("恢复队列、条目身份与播放位置，处于暂停且不加载音频", async () => {
    const player = newPlayer();
    useStatusStore().shuffleMode = "off";
    const list = [1, 2, 3].map((id) => track(id));
    void player.updatePlayList(list, list[1]);
    await settle();
    engine.ready();
    await settle();
    await player.addNextSong(track(1));
    player.setSeek(61_000);
    player.flushPersistence();
    await settle();
    const saved = useDataStore().playList.map((e) => e.entryId);

    // 重启：新的 pinia 与控制器，同一份保存的数据
    setActivePinia(createPinia());
    engine = new FakeEngine();
    const restored = newPlayer();
    expect(await restored.restore()).toBe(true);
    expect(ids()).toEqual([1, 2, 1, 3]);
    expect(useDataStore().playList.map((e) => e.entryId)).toEqual(saved);
    expect(useStatusStore().playIndex).toBe(1);
    expect(useMusicStore().playSong.id).toBe(2);
    expect(useStatusStore().currentTime).toBe(61_000);
    expect(restored.playbackState).toBe("paused");
    expect(useStatusStore().playStatus).toBe(false);
    expect(engine.loads).toEqual([]);

    // 继续播放从保存的位置开始
    await restored.play();
    expect(engine.loads).toEqual(["/audio/2"]);
    expect(engine.currentTime).toBe(61);
    // 新条目不与恢复的条目重复
    await restored.addNextSong(track(7));
    const entryIds = useDataStore().playList.map((e) => e.entryId);
    expect(new Set(entryIds).size).toBe(entryIds.length);
  });

  it("恢复随机模式下的原始顺序", async () => {
    const player = newPlayer();
    const list = [1, 2, 3, 4].map((id) => track(id));
    void player.updatePlayList(list);
    await settle();
    await player.toggleShuffle("on");
    player.flushPersistence();
    await settle();

    setActivePinia(createPinia());
    useStatusStore().shuffleMode = "on";
    const restored = newPlayer();
    await restored.restore();
    await restored.toggleShuffle("off");
    expect(ids()).toEqual([1, 2, 3, 4]);
    expect(useStatusStore().playIndex).toBe(0);
  });

  it("位置属于别的条目时从头开始", async () => {
    const player = newPlayer();
    void player.updatePlayList([track(1), track(2)]);
    await settle();
    player.setSeek(30_000);
    void player.togglePlayIndex(1);
    await settle();
    player.flushPersistence();
    await settle();
    storage.position = { entryId: -1, positionMs: 99_000 };

    setActivePinia(createPinia());
    const restored = newPlayer();
    await restored.restore();
    expect(useStatusStore().currentTime).toBe(0);
  });
});

describe("漫游 FM", () => {
  const album = (id: number, ...ids: number[]) => ({
    id,
    tracks: ids.map((t) => track(t)),
  });
  const fmState = (
    albums: FmState["albums"],
    trackId: number | null = null,
  ): FmState => ({ albums, trackId, groupIds: [] });

  /** 开始时当前专辑是 1（歌曲 11、12），下一张是 2（歌曲 21）；听完 1 之后接上 3 */
  function setup(first = fmState([album(1, 11, 12), album(2, 21)])) {
    const player = newPlayer({
      source: {
        recordPlay: vi.fn(async () => {}),
        audioStatus: vi.fn(async () => 200),
        fm: vi.fn(async () => first),
        fmFinish: vi.fn(async () => fmState([album(2, 21), album(3, 31)])),
        fmSkip: vi.fn(async () => fmState([album(2, 21), album(3, 31)])),
        fmPosition: vi.fn(async () => {}),
      } as unknown as PlaybackDeps["source"],
    });
    return {
      player,
      source: deps.source as unknown as Record<
        string,
        ReturnType<typeof vi.fn>
      >,
    };
  }
  const queue = ids;

  it("从记录的歌曲开始，关闭随机，不提示", async () => {
    const { player, source } = setup(
      fmState([album(1, 11, 12), album(2, 21)], 12),
    );
    useStatusStore().shuffleMode = "on";
    await player.playFm();
    await settle();
    expect(useMusicStore().playPlaylistId).toBe("fm");
    expect(useMusicStore().playSong.id).toBe(12);
    expect(useStatusStore().shuffleMode).toBe("off");
    expect(queue()).toEqual([11, 12, 21]);
    expect(deps.message.success).not.toHaveBeenCalled();
    expect(source.fmPosition).toHaveBeenCalledWith(12);
  });

  it("专辑内切歌不计数，放完最后一首才计一次并接上后面的专辑", async () => {
    const { player, source } = setup();
    await player.playFm();
    await player.nextOrPrev("next");
    await settle();
    expect(useMusicStore().playSong.id).toBe(12);
    expect(source.fmFinish).not.toHaveBeenCalled();
    await player.nextOrPrev("next");
    await settle();
    expect(source.fmFinish).toHaveBeenCalledTimes(1);
    expect(source.fmFinish).toHaveBeenCalledWith(1);
    expect(useMusicStore().playSong.id).toBe(21);
    expect(queue()).toEqual([11, 12, 21, 31]);
  });

  it("“听完”先于新专辑的收听位置到达服务端", async () => {
    const { player, source } = setup();
    await player.playFm();
    await player.nextOrPrev("next");
    await settle();
    source.fmPosition.mockClear();
    await player.nextOrPrev("next");
    await settle();
    expect(source.fmFinish.mock.invocationCallOrder[0]).toBeLessThan(
      source.fmPosition.mock.invocationCallOrder[0],
    );
  });

  it("单曲循环与不循环都不影响 FM 接下去", async () => {
    const { player } = setup(fmState([album(1, 11), album(2, 21)]));
    useStatusStore().repeatMode = "one";
    await player.playFm();
    await player.nextOrPrev("next", true, true);
    await settle();
    expect(useMusicStore().playSong.id).toBe(21);
    useStatusStore().repeatMode = "off";
    await player.nextOrPrev("next", true, true);
    await settle();
    expect(useMusicStore().playSong.id).toBe(31);
  });

  it("用户插入的歌不算专辑结束", async () => {
    const { player, source } = setup();
    await player.playFm();
    await player.addNextSong(track(99));
    await player.nextOrPrev("next");
    await settle();
    expect(useMusicStore().playSong.id).toBe(99);
    expect(source.fmFinish).not.toHaveBeenCalled();
  });

  it("取不到下一张专辑时停住，不绕回开头", async () => {
    const { player, source } = setup(fmState([album(1, 11)]));
    source.fmFinish.mockRejectedValue(new Error("offline"));
    await player.playFm();
    await settle();
    await player.nextOrPrev("next", true, true);
    await settle();
    expect(useMusicStore().playSong.id).toBe(11);
    expect(deps.message.error).toHaveBeenCalledWith("暂时取不到下一张专辑");
    expect(player.playbackState).toBe("ended");
  });

  it("跳过专辑后播放下一张；没在听 FM 时不打断当前播放", async () => {
    const { player, source } = setup();
    await player.updatePlayList([track(1), track(2)], track(1), "album-9");
    await settle();
    await player.fmSkipAlbum(1);
    await settle();
    expect(source.fmSkip).toHaveBeenCalledWith(1);
    expect(useMusicStore().playPlaylistId).toBe("album-9");
    expect(useMusicStore().playSong.id).toBe(1);

    await player.playFm();
    await settle();
    await player.fmSkipAlbum();
    await settle();
    expect(useMusicStore().playSong.id).toBe(21);
    expect(queue()).toEqual([21, 31]);
    expect(source.fmFinish).not.toHaveBeenCalled();
  });

  it("换去听别的列表后不再通知 FM，切歌也不受影响", async () => {
    const { player, source } = setup();
    await player.playFm();
    await settle();
    source.fmPosition.mockClear();
    await player.updatePlayList([track(1), track(2)], track(1), "album-9");
    await player.nextOrPrev("next");
    await settle();
    expect(useMusicStore().playSong.id).toBe(2);
    expect(source.fmPosition).not.toHaveBeenCalled();
    expect(source.fmFinish).not.toHaveBeenCalled();
  });

  it("FM 中不能开启随机", async () => {
    const { player } = setup();
    await player.playFm();
    await player.toggleShuffle("on");
    expect(useStatusStore().shuffleMode).toBe("off");
  });
});

describe("补全歌词", () => {
  it("更新队列中的条目；正在播放时立即重新读取，进度不变", async () => {
    const setupSong = vi.fn(async () => {});
    const player = newPlayer({ setupSong });
    await player.updatePlayList([track(1), track(2)], track(1));
    await settle();
    useStatusStore().currentTime = 5000;
    setupSong.mockClear();

    player.lyricsAdded(track(2));
    expect(useDataStore().playList[1].hasLyrics).toBe(true);
    expect(setupSong).not.toHaveBeenCalled();

    player.lyricsAdded(track(1));
    expect(useDataStore().playList[0].hasLyrics).toBe(true);
    expect(setupSong).toHaveBeenCalledTimes(1);
    expect(useStatusStore().currentTime).toBe(5000);
  });
});

describe("音量平衡", () => {
  const withGain = (id: number, gainDb: number | null): Track => ({
    ...track(id),
    gainDb,
  });

  it("按歌曲的 gainDb 施加增益，换歌时跟着换", async () => {
    const player = newPlayer();
    void player.updatePlayList([withGain(1, -6), withGain(2, null)]);
    await settle();
    expect(engine.replayGain).toBeCloseTo(10 ** (-6 / 20));
    void player.nextOrPrev("next");
    await settle();
    expect(engine.replayGain).toBe(1);
  });

  it("在线歌曲不做均衡", async () => {
    const player = newPlayer();
    void player.updatePlayList([
      { ...withGain(1, -6), source: "online", playable: true },
    ]);
    await settle();
    expect(engine.replayGain).toBe(1);
  });

  it("关闭后恢复原音量，重新开启时作用于当前歌曲", async () => {
    const player = newPlayer();
    void player.updatePlayList([withGain(1, 3)]);
    await settle();
    expect(engine.replayGain).toBeCloseTo(10 ** (3 / 20));
    useSettingStore().enableReplayGain = false;
    expect(player.applyReplayGain()).toBe(1);
    expect(engine.replayGain).toBe(1);
    useSettingStore().enableReplayGain = true;
    player.applyReplayGain();
    expect(engine.replayGain).toBeCloseTo(10 ** (3 / 20));
  });
});
