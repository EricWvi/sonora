<!-- 首页。问候语照搬 SPlayer views/Home/index.vue；快捷入口沿用 SPlayer 首页的 SongListCard，
栏目沿用 HomeOnline 的标题与列表。内容：快捷入口（每日推荐、我喜欢的音乐、私人 FM）、
“最近添加的歌曲”、“最近添加的专辑”、“最近播放”；栏目的显示与顺序由 homePageSections
设置决定。快捷入口的布局照搬 HomeOnline（左边两张上下叠放，右边是私人 FM）；与 SPlayer 不同：
卡片固定显示，账号是页面载入后才读到的，按登录状态切换卡片会闪动，
没有歌曲时用默认封面，没有登录网易云时点开每日推荐，在该页提示登录。 -->
<template>
  <div class="home">
    <div v-if="settingStore.showHomeGreeting" class="welcome">
      <n-h1>{{ greetings }}</n-h1>
      <n-text depth="3">由此开启好心情 ~</n-text>
    </div>
    <!-- 快捷入口 -->
    <div class="main-rec">
      <div class="main-rec-grid">
        <n-flex :size="20" class="rec-list" justify="space-between" vertical>
          <SongListCard
            :data="dailySongs"
            :cover="dailySongs.length ? undefined : defaultCovers.song"
            :title="dailySongsTitle"
            :height="90"
            description="根据你的音乐口味 · 每日更新"
            size="small"
            :hiddenCover="settingStore.hiddenCovers.home"
            @click="router.push({ name: 'daily-songs' })"
          />
          <SongListCard
            :data="likedSongs"
            :cover="likedSongs.length ? undefined : defaultCovers.song"
            :height="90"
            title="我喜欢的音乐"
            description="发现你独特的音乐品味"
            size="small"
            :hiddenCover="settingStore.hiddenCovers.home"
            @click="router.push({ name: 'like-songs' })"
          />
        </n-flex>
        <!-- 私人 FM -->
        <PersonalFM />
      </div>
    </div>
    <ListError v-if="error" :error="error" @retry="getHomeData" />
    <!-- 栏目 -->
    <div
      v-for="item in sortedSections"
      v-else
      :key="item.key"
      class="rec-public"
    >
      <n-flex
        class="title"
        align="center"
        justify="space-between"
        @click="router.push({ name: item.route })"
      >
        <n-h3 prefix="bar">
          <n-text>{{ item.name }}</n-text>
          <SvgIcon :size="26" name="Right" />
        </n-h3>
      </n-flex>
      <!-- 列表 -->
      <CoverList
        v-if="item.key !== 'history'"
        :data="item.key === 'song' ? newSongs : recentAlbums"
        :loading="loading"
        :loadingNum="6"
        :type="item.key"
        :hiddenCover="settingStore.hiddenCovers.home"
        fill-rows
      />
      <SongList
        v-else
        :data="recentSongs"
        :loading="loading"
        height="auto"
        :hidden-cover="settingStore.hiddenCovers.home"
        disabledSort
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import type { Album, Track } from "@/api/types";
import { ApiError, source } from "@/api/source";
import { useSettingStore, useUserStore } from "@/stores";
import { defaultCovers } from "@/utils/cover";
import { getGreeting } from "@/utils/time";
import SvgIcon from "@/components/Global/SvgIcon.vue";
import { NText } from "naive-ui";
import SongListCard from "@/components/Card/SongListCard.vue";
import CoverList from "@/components/List/CoverList.vue";
import SongList from "@/components/List/SongList.vue";
import ListError from "@/components/List/ListError.vue";
import PersonalFM from "@/components/Player/PlayerComponents/PersonalFM.vue";

const router = useRouter();
const settingStore = useSettingStore();
const userStore = useUserStore();

// 问候语。与 SPlayer 不同（Eric，2026-11-07）：登录了网易云时用网易云的昵称，否则用 Sonora 的名字。
// 网易云的账号状态比用户信息读得晚，读到之前不显示名字，免得名字先出现再被换掉。
const greetings = computed(() => {
  const greeting = getGreeting();
  if (!userStore.neteaseChecked) return greeting;
  const name =
    (userStore.onlineSignedIn && userStore.netease.profile?.nickname) ||
    userStore.user?.name ||
    "";
  return name ? `${greeting}，${name}` : greeting;
});

// 日推标题：照搬 SPlayer HomeOnline，日历图标里是今天的日期
const dailySongsTitle = computed(() => {
  if (settingStore.hiddenCovers.home) return "每日推荐";
  return h("div", { class: "date" }, [
    h("div", { class: "date-icon" }, [
      h(SvgIcon, { name: "Calendar-Empty", size: 30, depth: 2 }),
      h(NText, null, () => new Date().getDate()),
    ]),
    h(NText, { class: "name text-hidden" }, () => "每日推荐"),
  ]);
});

// 每日推荐：单独读取，接口服务慢或失败时不影响首页其他内容，卡片只是没有封面
const dailySongs = shallowRef<Track[]>([]);
const getDailySongs = async () => {
  if (!userStore.onlineSignedIn) {
    dailySongs.value = [];
    return;
  }
  try {
    dailySongs.value = (await source.onlineDailySongs()).items;
  } catch (e) {
    console.error("获取每日推荐失败：", e);
  }
};
// 网易云账号是页面载入后才读到的
watch(() => userStore.onlineSignedIn, getDailySongs);

// 首页数据
const likedSongs = shallowRef<Track[]>([]);
const newSongs = shallowRef<Track[]>([]);
const recentAlbums = shallowRef<Album[]>([]);
const recentSongs = shallowRef<Track[]>([]);
const loading = ref<boolean>(true);
const error = ref<string>("");

// 栏目：路由为标题指向的页面
const sections = {
  song: { route: "songs" },
  album: { route: "albums" },
  history: { route: "history" },
} as const;

// 根据设置过滤和排序栏目
const sortedSections = computed(() =>
  settingStore.homePageSections
    .filter((section) => section.visible && section.key in sections)
    .sort((a, b) => a.order - b.order)
    .map((section) => ({ ...section, ...sections[section.key] })),
);

// 获取首页数据
const getHomeData = async () => {
  loading.value = true;
  error.value = "";
  try {
    const [likes, songs, albums, history] = await Promise.all([
      source.mixedLikes({ limit: 50 }),
      // 默认新入库在前
      source.tracks({ limit: 12 }),
      // 与 SPlayer 不同：最近添加的专辑不含标记为单曲的专辑
      source.albums({
        sort: "createTime",
        order: "desc",
        limit: 12,
        hideSingles: true,
      }),
      source.mixedHistory({ limit: 10 }),
    ]);
    likedSongs.value = likes.items;
    newSongs.value = songs.items;
    recentAlbums.value = albums.items;
    recentSongs.value = history.items.map((entry) => entry.track);
  } catch (e) {
    error.value = e instanceof ApiError ? e.message : "加载失败";
  } finally {
    loading.value = false;
  }
};

// 每次进入都重新读取；keep-alive 首次激活与挂载同时发生，跳过
const isActivated = ref<boolean>(false);
onActivated(() => {
  if (isActivated.value) {
    getHomeData();
    getDailySongs();
  }
  isActivated.value = true;
});
onMounted(() => {
  getHomeData();
  getDailySongs();
});
</script>

<style lang="scss" scoped>
.home {
  width: 100%;
  max-width: 1500px;
  margin: 0 auto;
  .welcome {
    margin-top: 8px;
    margin-bottom: 20px;
    .n-h1 {
      margin: 0;
      font-weight: bold;
    }
  }
}
.main-rec {
  .main-rec-grid {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 20px;
  }
  // 卡片标题由渲染函数生成，要用 :deep
  :deep(.date) {
    display: flex;
    align-items: center;
    margin-bottom: 4px;
    .date-icon {
      position: relative;
      display: flex;
      align-items: center;
      justify-content: center;
      margin-right: 4px;
      .n-text {
        position: absolute;
        font-size: 12px;
        color: var(--primary-hex);
        line-height: normal;
        margin-top: 4px;
        transform: scale(0.8);
      }
    }
    .name {
      font-size: 18px;
      font-weight: bold;
    }
  }
}
.title {
  margin-top: 28px;
  padding: 0 4px;
  width: max-content;
  .n-h {
    margin: 0;
    display: flex;
    align-items: center;
    cursor: pointer;
    .n-icon {
      opacity: 0;
      transform: translateX(4px);
      transition:
        opacity 0.3s,
        transform 0.3s;
    }
    &:hover {
      .n-icon {
        opacity: 1;
        transform: translateX(0);
      }
    }
  }
}
.rec-public {
  .song-list {
    margin-top: 20px;
  }
}
</style>
