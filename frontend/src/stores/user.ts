import { defineStore } from "pinia";
import { ApiError, loginUrl, source } from "@/api/source";
import { connection } from "@/api/connection";
import type { Playlist, User } from "@/api/types";

/** 当前登录用户与其歌单。 */
export const useUserStore = defineStore("sonora-user", () => {
  const user = ref<User | null>(null);
  const playlists = ref<Playlist[]>([]);
  /** 本浏览器已退出登录，等待用户重新登录。 */
  const signedOut = ref(false);

  // 会话已失效：回到登录页。本浏览器已退出时停在退出页，否则 Authelia 会立即重新登录。
  function toLogin() {
    if (!signedOut.value) location.assign(loginUrl);
  }
  connection.onUnauthorized(toLogin);

  function unauthorized(error: unknown): boolean {
    if (!(error instanceof ApiError) || error.status !== 401) return false;
    toLogin();
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

  /** 歌单写入后重新读取侧栏与“添加到歌单”的列表。 */
  async function loadPlaylists(): Promise<void> {
    try {
      playlists.value = (await source.playlists()).items;
    } catch (error) {
      if (!unauthorized(error)) console.error("❌ 读取歌单失败:", error);
    }
  }

  // 服务恢复后重新读取用户与歌单（启动时不可达则一直为空）
  connection.onReconnect(() => {
    if (!signedOut.value)
      load().catch((error: unknown) =>
        console.error("重新读取用户数据失败", error),
      );
  });

  async function logout(): Promise<void> {
    await source.logout();
    user.value = null;
    playlists.value = [];
    signedOut.value = true;
  }

  return { user, playlists, signedOut, load, loadPlaylists, logout };
});
