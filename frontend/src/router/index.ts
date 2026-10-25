import {
  createRouter,
  createWebHashHistory,
  type NavigationGuard,
  type RouteRecordRaw,
} from "vue-router";
import HomeView from "@/views/HomeView.vue";
import EmptyView from "@/views/EmptyView.vue";

// 路由名沿用 SPlayer：详情页以 `?id=` 指定对象，缺少时转到 404。
const requireId: NavigationGuard = (to) => (to.query.id ? true : "/404");

const routes: RouteRecordRaw[] = [
  { path: "/", name: "home", component: HomeView },
  {
    path: "/songs",
    name: "songs",
    component: EmptyView,
    meta: { title: "歌曲" },
  },
  {
    path: "/albums",
    name: "albums",
    component: EmptyView,
    meta: { title: "专辑" },
  },
  {
    path: "/album",
    name: "album",
    beforeEnter: requireId,
    component: EmptyView,
    meta: { title: "专辑详情" },
  },
  {
    path: "/artists",
    name: "artists",
    component: EmptyView,
    meta: { title: "艺术家" },
  },
  {
    path: "/artist",
    name: "artist",
    beforeEnter: requireId,
    redirect: (to) => ({ name: "artist-songs", query: to.query }),
    children: [
      {
        path: "songs",
        name: "artist-songs",
        component: EmptyView,
        meta: { title: "艺术家 · 歌曲" },
      },
      {
        path: "albums",
        name: "artist-albums",
        component: EmptyView,
        meta: { title: "艺术家 · 专辑" },
      },
    ],
  },
  {
    path: "/like-songs",
    name: "like-songs",
    component: EmptyView,
    meta: { title: "我喜欢的音乐" },
  },
  {
    path: "/history",
    name: "history",
    component: EmptyView,
    meta: { title: "最近播放" },
  },
  {
    path: "/playlist",
    name: "playlist",
    beforeEnter: requireId,
    component: EmptyView,
    meta: { title: "歌单详情" },
  },
  {
    path: "/search",
    name: "search",
    beforeEnter: (to) => (to.query.keyword ? true : "/404"),
    redirect: (to) => ({ name: "search-songs", query: to.query }),
    children: ["songs", "albums", "artists", "playlists"].map((path) => ({
      path,
      name: `search-${path}`,
      component: EmptyView,
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
