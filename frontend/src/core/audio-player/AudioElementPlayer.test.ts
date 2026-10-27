import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { AudioElementPlayer } from "./AudioElementPlayer";

/** 测试环境没有 HTMLAudioElement，只模拟播放位置与 seeked 事件 */
class FakeAudio extends EventTarget {
  currentTime = 0;
  crossOrigin: string | null = null;
}

describe("AudioElementPlayer 的音频与歌词同步补偿", () => {
  let audio: FakeAudio;

  beforeEach(() => {
    vi.stubGlobal(
      "Audio",
      vi.fn(function () {
        audio = new FakeAudio();
        return audio;
      }),
    );
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("没有补偿时为元素的播放位置", () => {
    const player = new AudioElementPlayer();
    audio.currentTime = 12.5;
    expect(player.currentTime).toBe(12.5);
  });

  it("正值让播放位置提前，负值推后", () => {
    const player = new AudioElementPlayer();
    audio.currentTime = 12.5;
    player.setAudioDelayCompensation(300);
    expect(player.currentTime).toBeCloseTo(12.8);
    player.setAudioDelayCompensation(-1000);
    expect(player.currentTime).toBeCloseTo(11.5);
  });

  it("跳转中返回目标时间，跳转完成后再加上补偿", async () => {
    const player = new AudioElementPlayer();
    player.setAudioDelayCompensation(200);
    await player.seek(30);
    expect(audio.currentTime).toBe(30);
    expect(player.currentTime).toBe(30);
    audio.dispatchEvent(new Event("seeked"));
    expect(player.currentTime).toBeCloseTo(30.2);
  });
});
