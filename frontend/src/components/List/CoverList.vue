<!-- 专辑、歌单封面网格。照搬 SPlayer components/List/CoverList.vue，数据是播放端 API 的
Album 与 Playlist；不迁入视频、电台与播放量。专辑、歌单卡片的右键菜单为 CoverMenu（阶段 7）。专辑卡片在名称下显示艺术家
（沿用 SPlayer 视频卡片的艺术家样式），再显示“N首 · 年份”。
与 SPlayer 不同（阶段 6）：网格是虚拟的，只渲染视口附近的行；分页列表滚动到底时自动载入
下一页，不用“加载更多”按钮。
与 SPlayer 不同（首页“最近添加的歌曲”）：增加歌曲卡片。歌曲没有详情页，点卡片与播放按钮
都是播放这首，名称下的第三行换成专辑名，点击进入专辑；右键菜单用歌曲列表的 SongListMenu
（阶段 7），补上收藏、加入歌单、下一首播放等操作。 -->
<template>
  <Transition name="fade" mode="out-in">
    <div v-if="data.length > 0" :class="['cover-list', type]">
      <div
        ref="gridRef"
        class="cover-rows"
        :style="{ paddingTop: `${padTop}px`, paddingBottom: `${padBottom}px` }"
      >
        <div
          v-for="row in rows"
          :key="row.index"
          :data-row="row.index"
          :style="{ gridTemplateColumns: `repeat(${columns}, minmax(0, 1fr))` }"
          class="cover-grid"
        >
          <div
            v-for="item in row.items"
            :key="item.id"
            :class="['cover-item', { 'no-cover': hiddenCover }]"
            @click="openItem(item)"
            @contextmenu="openMenu($event, item)"
          >
            <!-- 封面 -->
            <div v-if="!hiddenCover" class="cover">
              <s-image
                :key="coverOf(item)"
                :src="coverOf(item)"
                default-src="/images/album.jpg?asset"
                class="cover-img"
                once
              />
              <!-- 简介 -->
              <div
                v-if="'description' in item && item.description"
                class="description"
              >
                <n-text class="text-hidden"> {{ item.description }}</n-text>
              </div>
              <!-- 播放按钮 -->
              <div class="play-btn" @click.stop>
                <n-button
                  :focusable="false"
                  :loading="loadingId === item.id"
                  secondary
                  circle
                  class="play"
                  @click.stop="playList(item)"
                >
                  <template #icon>
                    <SvgIcon
                      :size="32"
                      :name="isPlaying(item.id) ? 'Pause' : 'Play'"
                    />
                  </template>
                </n-button>
              </div>
            </div>
            <!-- 信息 -->
            <div class="cover-data">
              <n-text class="name text-hidden">{{ nameOf(item) }}</n-text>
              <!-- 专辑信息 -->
              <template v-if="'artists' in item">
                <div class="artists text-hidden">
                  <n-text
                    v-for="(ar, arIndex) in item.artists"
                    :key="arIndex"
                    class="ar"
                    @click.stop="openJumpArtist(item.artists, ar.id)"
                  >
                    {{
                      settingStore.hideBracketedContent
                        ? removeBrackets(ar.name)
                        : ar.name
                    }}
                  </n-text>
                </div>
                <!-- 歌曲的专辑 -->
                <n-text
                  v-if="'album' in item"
                  class="album-name text-hidden"
                  depth="3"
                  @click.stop="goAlbum(item)"
                >
                  {{ item.album?.title || "未知专辑" }}
                </n-text>
                <div v-else class="meta">
                  <n-text class="count" depth="3"
                    >{{ item.trackCount || 0 }}首</n-text
                  >
                  <n-text class="date" depth="3">{{
                    item.year || "未知年份"
                  }}</n-text>
                </div>
              </template>
              <!-- 歌单信息 -->
              <n-text v-else class="tip" depth="3"
                >{{ item.trackCount || 0 }}首</n-text
              >
            </div>
          </div>
        </div>
      </div>
      <!-- 载入下一页（沿用 SongList 的样式） -->
      <n-flex v-if="loadMore && loading" class="load-more" justify="center">
        <n-spin size="small" />
        <n-text>{{ loadingText || "努力加载中" }}</n-text>
      </n-flex>
    </div>
    <div v-else-if="loading" :class="['cover-list', 'loading', type]">
      <div class="cover-grid">
        <div
          v-for="item in loadingNum || 50"
          :key="item"
          :class="['cover-item', { 'no-cover': hiddenCover }]"
        >
          <div v-if="!hiddenCover" class="cover">
            <n-skeleton class="cover-img" />
          </div>
          <div
            class="cover-data"
            :style="hiddenCover ? { width: '100%', padding: '0 12px' } : {}"
          >
            <n-skeleton text round :repeat="2" />
          </div>
        </div>
      </div>
    </div>
    <!-- 空列表 -->
    <n-empty
      v-else
      :description="emptyDescription || '空空如也，怎么什么都没有啊'"
      size="large"
    />
  </Transition>
  <!-- 右键菜单 -->
  <CoverMenu v-if="type !== 'song'" ref="coverMenuRef" @to-play="playList" />
  <SongListMenu v-else ref="songMenuRef" />
</template>

<script setup lang="ts">
import type { Album, Playlist, Track } from "@/api/types";
import { source } from "@/api/source";
import { fetchAll } from "@/api/pages";
import { useMusicStore, useStatusStore, useSettingStore } from "@/stores";
import { debounce } from "lodash-es";
import { removeBrackets } from "@/utils/format";
import { coverUrl, defaultCovers } from "@/utils/cover";
import { openJumpArtist } from "@/utils/modal";
import { usePlayerController } from "@/core/player/PlayerController";
import { useVirtualGrid } from "@/composables/List/useVirtualGrid";
import CoverMenu from "@/components/Menu/CoverMenu.vue";
import SongListMenu from "@/components/Menu/SongListMenu.vue";

const props = defineProps<{
  data: (Album | Playlist | Track)[];
  type: "playlist" | "album" | "song";
  loadMore?: boolean;
  loading?: boolean;
  loadingNum?: number;
  loadingText?: string;
  emptyDescription?: string;
  hiddenCover?: boolean;
}>();

const emit = defineEmits<{
  // 载入下一页
  loadMore: [];
}>();

// 虚拟网格：列宽与间距同下方 .cover-grid 的 minmax(160px, 1fr) 与 gap
const {
  container: gridRef,
  columns,
  rows,
  padTop,
  padBottom,
  reset: resetGrid,
} = useVirtualGrid({
  items: () => props.data,
  minWidth: 160,
  gap: 20,
  onReachEnd: () => {
    if (props.loadMore && !props.loading) emit("loadMore");
  },
});
watch(() => props.hiddenCover, resetGrid);

const router = useRouter();
const musicStore = useMusicStore();
const statusStore = useStatusStore();
const settingStore = useSettingStore();
const player = usePlayerController();

// 正在载入播放列表的项（SPlayer 写在 item.loading 上）
const loadingId = ref<number>(0);

// 封面：SPlayer 的 coverSize.m
const coverOf = (item: Album | Playlist | Track) =>
  coverUrl(item.cover, "m", defaultCovers.album);

// 名称
const nameOf = (item: Album | Playlist | Track) =>
  "title" in item ? item.title : item.name;

// 播放列表来源（见 music store 的 playPlaylistId）
const listId = (id: number) => `${props.type}-${id}`;

// 是否处于当前播放列表
const isPlaying = (id: number) =>
  (props.type === "song"
    ? musicStore.playSong.id === id
    : musicStore.playPlaylistId === listId(id)) && statusStore.playStatus;

const isTrack = (item: Album | Playlist | Track): item is Track =>
  "album" in item;

// 点击卡片：歌曲直接播放，其他查看详情
const openItem = (item: Album | Playlist | Track) => {
  if (isTrack(item)) playList(item);
  else router.push({ name: props.type, query: { id: item.id } });
};

// 右键菜单
const coverMenuRef = ref<InstanceType<typeof CoverMenu> | null>(null);
const songMenuRef = ref<InstanceType<typeof SongListMenu> | null>(null);
const openMenu = (e: MouseEvent, item: Album | Playlist | Track) => {
  if (isTrack(item)) {
    const list = props.data as Track[];
    songMenuRef.value?.openDropdown(e, list, item, list.indexOf(item));
  } else if (props.type !== "song") {
    coverMenuRef.value?.openDropdown(e, item, props.type);
  }
};

// 歌曲所属的专辑
const goAlbum = ({ album }: Track) => {
  if (album) router.push({ name: "album", query: { id: album.id } });
};

// 播放歌单
const playList = debounce(
  async (item: Album | Playlist | Track) => {
    try {
      // 歌曲：播放这首，这几张卡片作为播放列表（同 SongList）
      if (isTrack(item)) {
        if (musicStore.playSong.id === item.id) return player.playOrPause();
        return player.updatePlayList(props.data as Track[], item);
      }
      // 是否为当前列表
      if (musicStore.playPlaylistId === listId(item.id))
        return player.playOrPause();
      // 开始加载
      loadingId.value = item.id;
      // 获取播放列表
      const list = await getListData(item.id);
      player.updatePlayList(list, undefined, listId(item.id));
    } catch (error) {
      console.log("Error to play: ", error);
    } finally {
      loadingId.value = 0;
    }
  },
  300,
  { leading: true, trailing: false },
);

// 获取列表数据：整个专辑或歌单
const getListData = (id: number): Promise<Track[]> =>
  props.type === "album"
    ? fetchAll((offset, limit) => source.tracks({ albumId: id, offset, limit }))
    : fetchAll((offset, limit) => source.playlistTracks(id, { offset, limit }));
</script>

<style lang="scss" scoped>
.cover-list {
  width: 100%;
  padding: 20px 4px;
  .cover-rows {
    display: flex;
    flex-direction: column;
    gap: 20px;
  }
  .cover-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(160px, 1fr));
    gap: 20px;
    @media (max-width: 600px) {
      grid-template-columns: repeat(3, 1fr);
      gap: 12px;
    }
  }
  .cover-item {
    position: relative;
    height: auto;
    border-radius: 16px;
    z-index: 0;
    transition:
      background-color 0.3s,
      transform 0.3s;
    cursor: pointer;
    .cover {
      position: relative;
      display: flex;
      align-items: center;
      justify-content: center;
      width: 100%;
      aspect-ratio: 1 / 1;
      border-radius: 16px;
      overflow: hidden;
      box-shadow: 0px 0px 4px 2px rgba(0, 0, 0, 0.1);
      transition:
        border-radius 0.3s,
        box-shadow 0.3s;
      :deep(img) {
        width: 100%;
        height: 100%;
        // opacity: 0;
        transition: opacity 0.35s ease-in-out;
      }
      .cover-img {
        transition:
          filter 0.3s,
          transform 0.3s;
      }
      .cover-mask {
        position: absolute;
        top: 0;
        left: 0;
        height: 30%;
        width: 100%;
        background: linear-gradient(rgba(0, 0, 0, 0.3), rgba(0, 0, 0, 0));
      }
      .play-count {
        position: absolute;
        display: flex;
        align-items: center;
        top: 10px;
        right: 12px;
        color: #fff;
        font-weight: bold;
        z-index: 2;
        .n-icon {
          color: #fff;
          font-size: 16px;
          margin-right: 4px;
        }
      }
      .description {
        position: absolute;
        left: 0;
        bottom: 0;
        width: 100%;
        padding: 40px 60px 12px 12px;
        background: linear-gradient(rgba(0, 0, 0, 0), rgba(0, 0, 0, 0.6));
        transform: translateY(100%);
        transition: transform 0.3s;
        .n-text {
          color: #fff;
          line-clamp: 3;
          -webkit-line-clamp: 3;
        }
      }
      .play {
        position: absolute;
        right: 10px;
        bottom: 10px;
        transform: translateY(8px);
        opacity: 0;
        transition: all 0.3s;
        background-color: #ffffff66;
        backdrop-filter: blur(6px);
        --n-width: 42px;
        --n-height: 42px;
        .n-icon {
          color: #fff;
        }
        :deep(.n-base-loading) {
          color: #fff;
        }
        &:active {
          background-color: #ffffff33;
        }
      }
      .n-skeleton {
        height: 100%;
      }
    }
    .cover-data {
      display: flex;
      flex-direction: column;
      padding: 12px;
      .name {
        font-size: 16px;
        line-clamp: 2;
        -webkit-line-clamp: 2;
      }
      .tip {
        font-size: 13px;
      }
      .meta {
        font-size: 13px;
        .count {
          &::after {
            content: "·";
            margin: 0 2px;
          }
        }
      }
      .album-name {
        font-size: 13px;
        transition: opacity 0.3s;
        &:hover {
          opacity: 0.8;
        }
      }
      .artists {
        margin-top: 2px;
        font-size: 13px;
        .ar {
          display: inline-flex;
          transition: opacity 0.3s;
          opacity: 0.6;
          cursor: pointer;
          &::after {
            content: "/";
            margin: 0 4px;
          }
          &:last-child {
            &::after {
              display: none;
            }
          }
          &:hover {
            opacity: 0.8;
          }
        }
      }
      :deep(.n-skeleton) {
        &:first-child {
          margin-bottom: 12px;
        }
      }
    }
    &:hover {
      background-color: rgba(var(--primary), 0.12);
      .cover {
        .cover-img {
          transform: scale(1.1);
          filter: brightness(0.8);
        }
        .description {
          transform: translateY(0);
        }
        .play {
          transform: translateY(0);
          opacity: 1;
        }
      }
    }
    &.no-cover {
      background-color: var(--surface-container-hex);
      border: 2px solid rgba(var(--primary), 0.12);
      padding: 0;
      overflow: hidden;
      &:hover {
        border-color: rgba(var(--primary), 0.58);
      }
      .cover-data {
        height: 100%;
        justify-content: center;
        .name {
          font-size: 18px;
          font-weight: bold;
        }
      }
    }
  }
  .load-more {
    margin: 20px 0;
    .n-spin-body {
      --n-size: 20px;
    }
  }
  &.video {
    .cover-grid {
      grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
    }
    .cover-item {
      .cover {
        aspect-ratio: 16/9;
      }
    }
  }
  &.loading {
    .cover {
      box-shadow: none;
    }
    .cover-item.no-cover {
      height: 80px;
      .cover-data {
        height: 100%;
        display: flex;
        flex-direction: column;
        justify-content: center;
      }
    }
  }
}
.n-empty {
  margin-top: 60px;
}
// Sonora：专辑与歌曲卡片的艺术家只占一行，过长时省略
.cover-list.album .cover-data .artists,
.cover-list.song .cover-data .artists {
  display: block;
  white-space: nowrap;
  text-overflow: ellipsis;
  .ar {
    display: inline;
  }
}
</style>
