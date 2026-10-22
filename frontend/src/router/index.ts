import { createRouter, createWebHashHistory } from "vue-router";
import HomeView from "@/views/HomeView.vue";
import EmptyView from "@/views/EmptyView.vue";

export const navigation = [
  { path: "/", name: "home", title: "为我推荐", icon: "Home" },
  { path: "/discover", name: "discover", title: "发现音乐", icon: "Discover" },
  { path: "/radio", name: "radio", title: "播客电台", icon: "Record" },
  { path: "/liked", name: "liked", title: "我喜欢的音乐", icon: "Favorite" },
  { path: "/collection", name: "collection", title: "我的收藏", icon: "Star" },
  { path: "/history", name: "history", title: "最近播放", icon: "History" },
];

export default createRouter({
  history: createWebHashHistory(),
  routes: [
    ...navigation.map((item) => ({
      path: item.path,
      name: item.name,
      component: item.name === "home" ? HomeView : EmptyView,
      meta: { title: item.title },
    })),
    {
      path: "/search",
      name: "search",
      component: EmptyView,
      meta: { title: "搜索" },
    },
    { path: "/:pathMatch(.*)*", redirect: "/" },
  ],
});
