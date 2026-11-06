import { defineStore } from "pinia";
import { ApiError, loginUrl, source } from "@/api/source";
import { connection } from "@/api/connection";
import { useDataStore } from "@/stores/data";
import { resumeAlbumImports } from "@/utils/albumImport";
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
  /** 已经读过一次网易云账号状态（成功或失败）。在这之前 `netease` 只是初始值 */
  const neteaseChecked = ref(false);
  /** 网易云上收藏的专辑与关注的艺术家，决定在线专辑页与艺术家页按钮的状态（同 SPlayer 的 userLikeData） */
  const onlineCollected = ref<{ albums: number[]; artists: number[] }>({
    albums: [],
    artists: [],
  });

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
      try {
        netease.value = await source.neteaseStatus();
      } finally {
        neteaseChecked.value = true;
      }
      if (!onlineSignedIn.value) return clearOnline();
      const [lists, likeIds] = await Promise.all([
        source.onlinePlaylists(),
        source.onlineLikeIds(),
      ]);
      onlinePlaylists.value = lists;
      useDataStore().userLikeData.onlineSongs = likeIds;
      void loadOnlineCollected();
    } catch (error) {
      if (!unauthorized(error)) console.error("❌ 读取网易云账号失败:", error);
    }
  }

  /** 读取收藏的专辑与关注的艺术家。要翻页，所以不拖慢账号的载入。 */
  async function loadOnlineCollected(): Promise<void> {
    try {
      const [albums, artists] = await Promise.all([
        source.onlineAlbums(),
        source.onlineArtists(),
      ]);
      onlineCollected.value = {
        albums: albums.items.map((item) => item.id),
        artists: artists.items.map((item) => item.id),
      };
    } catch (error) {
      if (!unauthorized(error)) console.error("❌ 读取网易云收藏失败:", error);
    }
  }

  function clearOnline(): void {
    onlinePlaylists.value = { created: [], collected: [] };
    onlineCollected.value = { albums: [], artists: [] };
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
      if (me.canImport) void resumeAlbumImports();
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

  /** 在线歌单写入后重新读取。 */
  async function loadOnlinePlaylists(): Promise<void> {
    try {
      onlinePlaylists.value = await source.onlinePlaylists();
    } catch (error) {
      if (!unauthorized(error)) console.error("❌ 读取在线歌单失败:", error);
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
    neteaseChecked,
    onlineCollected,
    load,
    loadPlaylists,
    loadOnlinePlaylists,
    loadNetease,
    logoutNetease,
    logout,
  };
});
