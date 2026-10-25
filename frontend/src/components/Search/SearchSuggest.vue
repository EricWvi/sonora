<!-- 搜索建议。照搬 SPlayer components/Search/SearchSuggest.vue，建议来自曲库搜索：
歌曲、艺术家、专辑与自己的歌单各取前几项（阶段 2 的“设置 → 常规”），
分类名称与侧栏一致。不迁入分享链接识别。 -->
<template>
  <Transition
    name="fadeDown"
    mode="out-in"
    @after-enter="calcSearchSuggestHeights"
    @after-leave="calcSearchSuggestHeights"
  >
    <n-card
      v-if="statusStore.searchFocus && statusStore.searchInputValue"
      class="search-suggest"
      content-style="padding: 0"
      :style="{
        height: `${searchSuggestHeights}px`,
        border: searchSuggestHeights === 0 ? 'none' : null,
      }"
    >
      <n-scrollbar class="scrollbar">
        <!-- 直接搜索 -->
        <div
          ref="directSearchRef"
          class="direct"
          @click="emit('toSearch', statusStore.searchInputValue, 'keyword')"
        >
          <SvgIcon name="Search" :depth="3" />
          <n-text class="text text-hidden"
            >直接搜索：{{ statusStore.searchInputValue }}</n-text
          >
        </div>
        <!-- 搜索建议 -->
        <Transition
          name="fade"
          mode="out-in"
          @after-leave="calcSearchSuggestHeights"
        >
          <div
            v-if="searchSuggestData.length && settingStore.enableSearchKeyword"
            ref="searchSuggestRef"
            class="all-suggest"
          >
            <div
              v-for="group in searchSuggestData"
              :key="group.type"
              class="suggest"
            >
              <div class="suggest-type">
                <SvgIcon :name="searchSuggestionsType[group.type].icon" />
                <n-text>{{ searchSuggestionsType[group.type].name }}</n-text>
              </div>
              <div
                v-for="(suggestItem, suggestIndex) in group.items"
                :key="suggestIndex"
                class="suggest-item"
                @click="emit('toSearch', suggestItem.target, group.type)"
              >
                <n-text class="name">{{ suggestItem.name }}</n-text>
                <n-text v-if="suggestItem.artist" class="artist" depth="3">
                  {{ suggestItem.artist }}
                </n-text>
              </div>
            </div>
          </div>
        </Transition>
      </n-scrollbar>
    </n-card>
  </Transition>
</template>

<script setup lang="ts">
import type { Album, Artist, Playlist, Track } from "@/api/types";
import { source } from "@/api/source";
import { useStatusStore, useSettingStore } from "@/stores";

/** 建议的分类，顺序与 SPlayer 相同；名称与 Sonora 的侧栏一致（歌曲、艺术家） */
type SuggestType = "songs" | "artists" | "albums" | "playlists";

/** 一条建议：显示的名称与艺术家，点击时交给搜索框处理的对象 */
interface SuggestItem {
  name: string;
  artist?: string;
  target: Track | Artist | Album | Playlist;
}

const emit = defineEmits<{
  toSearch: [key: string | SuggestItem["target"], type: string];
}>();

const statusStore = useStatusStore();
const settingStore = useSettingStore();

// 搜索建议数据
const searchSuggestData = ref<{ type: SuggestType; items: SuggestItem[] }[]>(
  [],
);
const searchSuggestHeights = ref<number>(0);

// 搜索建议元素
const directSearchRef = ref<HTMLElement | null>(null);
const searchSuggestRef = ref<HTMLElement | null>(null);

// 搜索建议分类
const searchSuggestionsType = {
  songs: {
    name: "歌曲",
    icon: "Music",
  },
  artists: {
    name: "艺术家",
    icon: "Artist",
  },
  albums: {
    name: "专辑",
    icon: "Album",
  },
  playlists: {
    name: "歌单",
    icon: "MusicList",
  },
};

// 请求序号，丢弃过期的结果
let requestId = 0;

// 获取搜索建议：曲库中各类的前几项
const getSearchSuggest = async (keywords: string) => {
  const id = ++requestId;
  searchSuggestData.value = [];
  try {
    const [songs, artists, albums, playlists] = await Promise.all([
      source.tracks({ q: keywords, limit: 4 }),
      source.artists({ q: keywords, limit: 2 }),
      source.albums({ q: keywords, limit: 2 }),
      source.playlists({ q: keywords }),
    ]);
    if (id !== requestId) return;
    const groups: { type: SuggestType; items: SuggestItem[] }[] = [
      {
        type: "songs",
        items: songs.items.map((song) => ({
          name: song.title,
          artist: song.artists[0]?.name,
          target: song,
        })),
      },
      {
        type: "artists",
        items: artists.items.map((artist) => ({
          name: artist.name,
          target: artist,
        })),
      },
      {
        type: "albums",
        items: albums.items.map((album) => ({
          name: album.title,
          artist: album.artists[0]?.name,
          target: album,
        })),
      },
      {
        type: "playlists",
        items: playlists.items
          .slice(0, 2)
          .map((playlist) => ({ name: playlist.name, target: playlist })),
      },
    ];
    searchSuggestData.value = groups.filter((group) => group.items.length);
  } catch (error) {
    console.error("搜索建议获取失败：", error);
  }
  // 计算高度
  nextTick(calcSearchSuggestHeights);
};

// 计算高度
const calcSearchSuggestHeights = () => {
  const directSearchHeight = directSearchRef.value?.offsetHeight;
  const searchSuggestionsHeight = searchSuggestRef.value?.offsetHeight;
  if (directSearchHeight || searchSuggestionsHeight) {
    const totalHeight =
      (directSearchHeight || 0) +
      (searchSuggestionsHeight || 0) +
      (searchSuggestionsHeight ? 8 : 0) +
      20;
    searchSuggestHeights.value = totalHeight;
  } else {
    searchSuggestHeights.value = 0;
  }
};

// 搜索框改变
watchDebounced(
  () => statusStore.searchInputValue,
  (val) => {
    if (!val || val === "" || !settingStore.enableSearchKeyword) return;
    getSearchSuggest(val.trim());
  },
  { debounce: 300 },
);
</script>

<style lang="scss" scoped>
.search-suggest {
  position: absolute;
  left: 0;
  top: 50px;
  width: 300px;
  border-radius: 8px;
  overflow: hidden;
  max-height: calc(100vh - 160px);
  z-index: 101;
  transition:
    height 0.3s ease,
    opacity 0.3s ease,
    transform 0.3s ease;
  :deep(.scrollbar) {
    max-height: calc(100vh - 160px);
    .n-scrollbar-content {
      padding: 10px;
    }
  }
  @media (max-width: 768px) {
    width: 100%;
  }
  .direct {
    display: flex;
    align-items: center;
    padding: 6px;
    border-radius: 8px;
    transition: background-color 0.3s;
    cursor: pointer;
    .n-icon {
      font-size: 16px;
      margin-right: 6px;
    }
    &:hover {
      background-color: var(--n-border-color);
    }
  }
  .all-suggest {
    margin-top: 8px;
    .suggest {
      margin-bottom: 8px;
      .suggest-type {
        display: flex;
        align-items: center;
        margin-bottom: 8px;
        color: var(--n-color-target);
        .n-icon {
          font-size: 18px;
          margin-right: 4px;
        }
        .n-text {
          color: var(--n-color-target);
        }
      }
      .suggest-item {
        padding: 14px 18px 14px 22px;
        margin-bottom: 8px;
        border-radius: 8px;
        transition: background-color 0.3s;
        cursor: pointer;
        .name {
          white-space: normal;
        }
        .artist {
          &::before {
            content: " - ";
          }
        }
        &:last-child {
          margin-bottom: 0;
        }
        &:hover {
          background-color: var(--n-border-color);
        }
      }
      &:last-child {
        margin-bottom: 0;
      }
    }
  }
}
</style>
