import { loginUrl, source, usingFixture } from "@/api/source";
import { isOnline } from "@/utils/track";
import { AudioElementPlayer } from "@/core/audio-player/AudioElementPlayer";
import { SilentAudioPlayer } from "@/core/audio-player/SilentAudioPlayer";
import { useMusicStore } from "@/stores";
import { getCoverColor } from "@/utils/color";
import { useLyricManager } from "./LyricManager";
import { mediaSessionManager } from "./MediaSessionManager";
import type { PlaybackDeps } from "./PlayerController";
import { browserStorage, memoryStorage } from "./storage";

/**
 * 浏览器中的播放控制器依赖。fixture 没有音频，换成不出声的引擎，队列只保存在内存中，
 * 不改动浏览器中保存的队列
 */
export function browserDeps(): PlaybackDeps {
  return {
    engine: usingFixture ? new SilentAudioPlayer() : new AudioElementPlayer(),
    resolveAudio: async (track) =>
      isOnline(track)
        ? source.onlineSongUrl(track.id)
        : source.audioUrl(track.id),
    source,
    storage: usingFixture ? memoryStorage() : browserStorage,
    message: {
      success: (...args) => window.$message.success(...args),
      error: (...args) => window.$message.error(...args),
      warning: (...args) => window.$message.warning(...args),
    },
    setupSong(song) {
      getCoverColor(useMusicStore().songCover);
      mediaSessionManager.updateMetadata(song);
      return useLyricManager().handleLyric(song);
    },
    resetSong() {
      useLyricManager().resetSongLyric();
      mediaSessionManager.clear();
    },
    unauthorized: () => location.assign(loginUrl),
    setTitle(title) {
      document.title = title;
    },
  };
}
