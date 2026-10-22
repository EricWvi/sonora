<template>
  <n-menu
    :value="String(route.name)"
    :options="options"
    :indent="0"
    :root-indent="26"
    :collapsed="collapsed"
    :collapsed-width="64"
    :collapsed-icon-size="22"
    @update:value="navigate"
  />
</template>

<script setup lang="ts">
import { NButton, NText, type MenuOption } from "naive-ui";
import SvgIcon from "@/components/Global/SvgIcon.vue";
import { navigation } from "@/router";

const props = defineProps<{ collapsed: boolean }>();
const emit = defineEmits<{ navigate: [] }>();
const route = useRoute();
const router = useRouter();
const options = computed<MenuOption[]>(() => [
  ...navigation.flatMap((item, index) => [
    ...(index === 3 ? [{ key: "divider", type: "divider" as const }] : []),
    {
      key: item.name,
      label:
        item.name === "liked"
          ? () =>
              h("div", { class: "user-liked" }, [
                h(NText, null, () => item.title),
                h(NButton, {
                  round: true,
                  strong: true,
                  secondary: true,
                  renderIcon: () => h(SvgIcon, { name: "HeartBit" }),
                  onClick: (event: Event) => event.stopPropagation(),
                  "aria-label": "心动模式",
                }),
              ])
          : item.title,
      icon: () =>
        h(SvgIcon, {
          name: item.icon,
          size: 22,
          style: index < 3 ? { transform: "translateY(-1px)" } : undefined,
        }),
    },
  ]),
  { key: "divider-two", type: "divider" },
  ...[
    { key: "user-playlists", title: "创建的歌单", icon: "PlaylistAdd" },
    { key: "liked-playlists", title: "收藏的歌单", icon: "PlaylistAddCheck" },
  ].map((item) => ({
    key: item.key,
    icon: props.collapsed ? () => h(SvgIcon, { name: item.icon }) : undefined,
    label: () =>
      h("div", { class: "user-list" }, [
        h(NText, { depth: 3 }, () => item.title),
        ...(item.key === "user-playlists"
          ? ["Menu", "Add"].map((name) =>
              h(NButton, {
                round: true,
                strong: true,
                secondary: true,
                renderIcon: () => h(SvgIcon, { name }),
                onClick: (event: Event) => event.stopPropagation(),
                "aria-label": name === "Add" ? "创建歌单" : "歌单模式",
              }),
            )
          : []),
      ]),
    children: [],
  })),
]);

// 导航仅切换展示页，不触发音乐服务请求。
function navigate(name: string): void {
  if (!navigation.some((item) => item.name === name)) return;
  void router.push({ name });
  emit("navigate");
}
</script>

<style lang="scss" scoped src="@/style/menu.scss"></style>
<style lang="scss" src="@/style/menu-labels.scss"></style>
