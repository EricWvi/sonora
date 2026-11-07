import { describe, expect, it } from "vitest";
import type { FmState, Track } from "@/api/types";
import { fmCurrentTrack, fmTracks, fmTracksToAppend } from "./fm";

const track = (id: number) => ({ id, title: `歌曲 ${id}` }) as Track;

const state: FmState = {
  albums: [
    { id: 1, tracks: [track(11), track(12), track(13)] },
    { id: 2, tracks: [track(21)] },
  ],
  trackId: null,
  groupIds: [],
};

describe("漫游 FM 的队列", () => {
  it("歌曲带所属专辑，按播放顺序", () => {
    expect(fmTracks(state).map((t) => [t.id, t.fmAlbumId])).toEqual([
      [11, 1],
      [12, 1],
      [13, 1],
      [21, 2],
    ]);
  });

  it("从记录的歌曲继续，没有记录或歌曲不在其中时从头开始", () => {
    expect(fmCurrentTrack(state)?.id).toBe(11);
    expect(fmCurrentTrack({ ...state, trackId: 12 })?.id).toBe(12);
    expect(fmCurrentTrack({ ...state, trackId: 99 })?.id).toBe(11);
    expect(fmCurrentTrack({ ...state, albums: [] })).toBeUndefined();
  });

  it("只接上队列里还没有的专辑，当前专辑从记录的歌曲开始", () => {
    const ids = (s: FmState, queued: number[]) =>
      fmTracksToAppend(s, new Set(queued)).map((t) => t.id);
    expect(ids(state, [1])).toEqual([21]);
    expect(ids(state, [1, 2])).toEqual([]);
    expect(ids({ ...state, trackId: 12 }, [])).toEqual([12, 13, 21]);
    expect(ids({ ...state, trackId: 99 }, [])).toEqual([11, 12, 13, 21]);
  });
});
