import { defineStore } from "pinia";
import { ApiError, loginUrl, source } from "@/api/source";
import { connection } from "@/api/connection";
import { useDataStore } from "@/stores/data";
import type { NeteaseStatus, Playlist, User } from "@/api/types";

/** 当前登录用户与其歌单。 */
export const useUserStore = defineStore("sonora-user", () => {
  const user = ref<User | null>(null);
  const playlists = ref<Playlist[]>([]);
  /** 本浏览器已退出登录，等待用户重新登录。 */
  const signedOut = ref(false);
  /** 网易云账号，与 Sonora 的登录互相独立：退出一个不影响另一个。 */
  const netease = ref<NeteaseStatus>({ state: "disabled", profile: null });
  /** 网易云上自己创建的歌单与收藏的歌单 */
  const onlinePlaylists = ref<{ created: Playlist[]; collected: Playlist[] }>({
    created: [],
    collected: [],
  });
  const onlineSignedIn = computed(() => netease.value.state === "signedIn");

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

  // 单独读取：接口服务慢或不可用时，不耽误页面载入。
  async function loadNetease(): Promise<void> {
    try {
      netease.value = await source.neteaseStatus();
      if (!onlineSignedIn.value) return clearOnline();
      const [lists, likeIds] = await Promise.all([
        source.onlinePlaylists(),
        source.onlineLikeIds(),
      ]);
      onlinePlaylists.value = lists;
      useDataStore().userLikeData.onlineSongs = likeIds;
    } catch (error) {
      if (!unauthorized(error)) console.error("❌ 读取网易云账号失败:", error);
    }
  }

  function clearOnline(): void {
    onlinePlaylists.value = { created: [], collected: [] };
    useDataStore().userLikeData.onlineSongs = [];
  }

  async function logoutNetease(): Promise<void> {
    await source.neteaseLogout();
    netease.value = { state: "signedOut", profile: null };
    clearOnline();
  }

  async function load(): Promise<void> {
    void loadNetease();
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

  return {
    user,
    playlists,
    signedOut,
    netease,
    onlinePlaylists,
    onlineSignedIn,
    load,
    loadPlaylists,
    loadNetease,
    logoutNetease,
    logout,
  };
});
