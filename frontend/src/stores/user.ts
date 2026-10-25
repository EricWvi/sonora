import { defineStore } from "pinia";
import { ApiError, loginUrl, source } from "@/api/source";
import type { Playlist, User } from "@/api/types";

/** 当前登录用户与其歌单。 */
export const useUserStore = defineStore("sonora-user", () => {
  const user = ref<User | null>(null);
  const playlists = ref<Playlist[]>([]);
  /** 本浏览器已退出登录，等待用户重新登录。 */
  const signedOut = ref(false);

  // 会话已失效：回到登录页。
  function unauthorized(error: unknown): boolean {
    if (!(error instanceof ApiError) || error.status !== 401) return false;
    location.assign(loginUrl);
    return true;
  }

  async function load(): Promise<void> {
    try {
      const [me, list] = await Promise.all([source.me(), source.playlists()]);
      user.value = me;
      playlists.value = list.items;
    } catch (error) {
      if (!unauthorized(error)) throw error;
    }
  }

  async function logout(): Promise<void> {
    await source.logout();
    user.value = null;
    playlists.value = [];
    signedOut.value = true;
  }

  return { user, playlists, signedOut, load, logout };
});
