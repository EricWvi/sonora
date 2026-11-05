import {
  createRouter,
  createWebHashHistory,
  type NavigationGuard,
  type RouteRecordRaw,
} from "vue-router";
import HomeView from "@/views/Home/index.vue";

// 路由名沿用 SPlayer：详情页以 `?id=` 指定对象，缺少时转到 404。
const requireId: NavigationGuard = (to) => (to.query.id ? true : "/404");

const routes: RouteRecordRaw[] = [
  { path: "/", name: "home", component: HomeView },
  {
    path: "/songs",
    name: "songs",
    component: () => import("@/views/Library/songs.vue"),
    meta: { title: "歌曲" },
  },
  {
    path: "/albums",
    name: "albums",
    component: () => import("@/views/Library/albums.vue"),
    meta: { title: "专辑" },
  },
  {
    path: "/album",
    name: "album",
    beforeEnter: requireId,
    component: () => import("@/views/List/album.vue"),
    meta: { title: "专辑详情" },
  },
  {
    path: "/artists",
    name: "artists",
    component: () => import("@/views/Library/artists.vue"),
    meta: { title: "艺术家" },
  },
  {
    path: "/artist",
    name: "artist",
    beforeEnter: requireId,
    component: () => import("@/views/Artist/layout.vue"),
    redirect: (to) => ({ name: "artist-songs", query: to.query }),
    children: [
      {
        path: "songs",
        name: "artist-songs",
        component: () => import("@/views/Artist/songs.vue"),
        meta: { title: "艺术家 · 歌曲" },
      },
      {
        path: "albums",
        name: "artist-albums",
        component: () => import("@/views/Artist/albums.vue"),
        meta: { title: "艺术家 · 专辑" },
      },
    ],
  },
  {
    path: "/like-songs",
    name: "like-songs",
    component: () => import("@/views/List/liked.vue"),
    meta: { title: "我喜欢的音乐" },
  },
  {
    path: "/collection",
    name: "collection",
    component: () => import("@/views/Collection/index.vue"),
    meta: { title: "我的收藏" },
  },
  {
    path: "/online",
    name: "online-list",
    component: () => import("@/views/List/onlineList.vue"),
    meta: { title: "在线歌单" },
  },
  {
    path: "/history",
    name: "history",
    component: () => import("@/views/History.vue"),
    meta: { title: "最近播放" },
  },
  {
    path: "/playlist",
    name: "playlist",
    beforeEnter: requireId,
    component: () => import("@/views/List/playlist.vue"),
    meta: { title: "歌单详情" },
  },
  {
    path: "/search",
    name: "search",
    beforeEnter: (to) => (to.query.keyword ? true : "/404"),
    component: () => import("@/views/Search/layout.vue"),
    redirect: (to) => ({ name: "search-songs", query: to.query }),
    children: (
      [
        ["songs", () => import("@/views/Search/songs.vue")],
        ["albums", () => import("@/views/Search/albums.vue")],
        ["artists", () => import("@/views/Search/artists.vue")],
        ["playlists", () => import("@/views/Search/playlists.vue")],
      ] as const
    ).map(([path, component]) => ({
      path,
      name: `search-${path}`,
      component,
      meta: { title: "搜索" },
    })),
  },
  {
    path: "/404",
    name: "404",
    component: () => import("@/views/Status/404.vue"),
  },
  {
    path: "/500",
    name: "500",
    component: () => import("@/views/Status/500.vue"),
  },
  { path: "/:pathMatch(.*)*", redirect: "/404" },
];

const router = createRouter({ history: createWebHashHistory(), routes });

// 切换页面时回到顶部（只改 hash 时除外），与 SPlayer 一致。
router.afterEach((to, from) => {
  if (to.fullPath.split("#")[0] === from.fullPath.split("#")[0]) return;
  requestAnimationFrame(() => {
    const container = document.querySelector(
      "#main-content .n-scrollbar-container",
    );
    if (container) container.scrollTop = 0;
  });
});

export default router;
