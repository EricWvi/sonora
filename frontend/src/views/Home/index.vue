<!-- 首页。问候语照搬 SPlayer views/Home/index.vue；快捷入口沿用 SPlayer 首页的 SongListCard，
栏目沿用 HomeOnline 的标题与列表。内容按阶段 2：快捷入口（我喜欢的音乐、歌曲）、
“最近添加的歌曲”、“最近添加的专辑”、“最近播放”；栏目的显示与顺序由 homePageSections
设置决定。 -->
<template>
  <div class="home">
    <div v-if="settingStore.showHomeGreeting" class="welcome">
      <n-h1>{{ greetings }}</n-h1>
      <n-text depth="3">由此开启好心情 ~</n-text>
    </div>
    <!-- 快捷入口 -->
    <div class="main-rec">
      <div class="main-rec-grid">
        <SongListCard
          :data="likedSongs"
          :height="90"
          title="我喜欢的音乐"
          description="发现你独特的音乐品味"
          size="small"
          :hiddenCover="settingStore.hiddenCovers.home"
          @click="router.push({ name: 'like-songs' })"
        />
        <SongListCard
          :data="librarySongs"
          :height="90"
          title="歌曲"
          :description="`曲库中的 ${libraryTotal} 首歌曲`"
          size="small"
          :hiddenCover="settingStore.hiddenCovers.home"
          @click="router.push({ name: 'songs' })"
        />
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
import { getGreeting } from "@/utils/time";
import SongListCard from "@/components/Card/SongListCard.vue";
import CoverList from "@/components/List/CoverList.vue";
import SongList from "@/components/List/SongList.vue";
import ListError from "@/components/List/ListError.vue";

const router = useRouter();
const settingStore = useSettingStore();
const userStore = useUserStore();

// 问候语
const greetings = computed(() => {
  const greeting = getGreeting();
  const name = userStore.user?.name ?? "";
  return name ? `${greeting}，${name}` : greeting;
});

// 首页数据
const likedSongs = shallowRef<Track[]>([]);
const librarySongs = shallowRef<Track[]>([]);
const libraryTotal = ref<number>(0);
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
      // 默认新入库在前，“最近添加的歌曲”取前 12 首
      source.tracks({ limit: 50 }),
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
    librarySongs.value = songs.items;
    libraryTotal.value = songs.total;
    newSongs.value = songs.items.slice(0, 12);
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
  if (isActivated.value) getHomeData();
  isActivated.value = true;
});
onMounted(getHomeData);
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
