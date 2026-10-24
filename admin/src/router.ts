import { createRouter, createWebHistory } from "vue-router";
import { ensureStatus, session } from "./session";
// Views are bundled, not lazy-loaded: the console is small, and a lazy
// chunk cannot load while the service is down, which would hide the
// unreachable state.
import Setup from "./views/Setup.vue";
import Login from "./views/Login.vue";
import Layout from "./views/Layout.vue";
import Overview from "./views/Overview.vue";
import Upload from "./views/Upload.vue";
import Tracks from "./views/Tracks.vue";
import Albums from "./views/Albums.vue";
import Artists from "./views/Artists.vue";

const router = createRouter({
  history: createWebHistory("/admin/"),
  routes: [
    {
      path: "/setup",
      component: Setup,
      meta: { public: true },
    },
    {
      path: "/login",
      component: Login,
      meta: { public: true },
    },
    {
      path: "/",
      component: Layout,
      children: [
        {
          path: "",
          component: Overview,
          meta: { title: "概况" },
        },
        {
          path: "upload",
          component: Upload,
          meta: { title: "上传" },
        },
        {
          path: "tracks",
          component: Tracks,
          meta: { title: "歌曲" },
        },
        {
          path: "albums",
          component: Albums,
          meta: { title: "专辑" },
        },
        {
          path: "artists",
          component: Artists,
          meta: { title: "艺术家" },
        },
      ],
    },
    { path: "/:rest(.*)*", redirect: "/" },
  ],
});

// The page follows the server state: setup until an admin exists, then
// sign-in until there is a session.
router.beforeEach(async (to) => {
  await ensureStatus();
  if (!session.loaded) return true; // App shows the unreachable state
  if (!session.initialized) return to.path === "/setup" ? true : "/setup";
  if (!session.admin) {
    if (to.path === "/login") return true;
    return {
      path: "/login",
      query: to.fullPath === "/" ? {} : { next: to.fullPath },
    };
  }
  if (to.meta.public) return "/";
  return true;
});

router.afterEach((to) => {
  const title = to.matched.map((r) => r.meta.title).find(Boolean) as
    string | undefined;
  document.title = title ? `${title} · Sonora 管理后台` : "Sonora 管理后台";
});

export default router;
