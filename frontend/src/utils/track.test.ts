import { describe, expect, it } from "vitest";
import type { Track } from "@/api/types";
import { sameTrack, sortTracks, trackKey } from "./track";

const track = (over: Partial<Track>): Track => ({
  id: 1,
  title: "",
  artists: [],
  album: null,
  trackNo: 0,
  discNo: 0,
  year: 0,
  durationMs: 0,
  format: "mp3",
  cover: null,
  hasLyrics: false,
  ...over,
});

describe("歌曲身份", () => {
  it("本地与在线歌曲的 ID 相同时不是同一首", () => {
    const local = track({ id: 7 });
    const online = track({ id: 7, source: "online" });
    expect(sameTrack(local, online)).toBe(false);
    expect(sameTrack(online, track({ id: 7, source: "online" }))).toBe(true);
    expect(trackKey(local)).not.toBe(trackKey(online));
  });
});

describe("混合列表排序", () => {
  const list = [
    track({ id: 1, title: "歌10", durationMs: 3 }),
    track({ id: 2, title: "歌2", durationMs: 1, source: "online" }),
    track({ id: 3, title: "歌1", durationMs: 2 }),
  ];
  const ids = (l: Track[]) => l.map((t) => t.id);

  it("默认保持服务端的喜欢时间顺序", () => {
    expect(ids(sortTracks(list, "default", "asc"))).toEqual([1, 2, 3]);
    expect(ids(sortTracks(list, "title", "default"))).toEqual([1, 2, 3]);
  });
  it("标题按数字大小比较，在线歌曲一起排序", () => {
    expect(ids(sortTracks(list, "title", "asc"))).toEqual([3, 2, 1]);
  });
  it("时长可以倒序", () => {
    expect(ids(sortTracks(list, "duration", "desc"))).toEqual([1, 3, 2]);
  });
  it("在线歌曲没有入库时间，保持列表顺序", () => {
    expect(ids(sortTracks(list, "createTime", "desc"))).toEqual([1, 2, 3]);
  });
});
