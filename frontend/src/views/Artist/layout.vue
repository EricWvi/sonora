<!-- 艺术家详情。照搬 SPlayer views/Artist/layout.vue：图片、名称、歌曲数与专辑数，
“歌曲 / 专辑”标签页。图片是管理后台上传的艺术家图片，没有时为 SPlayer 的默认图片。
与 SPlayer 的圆形图片不同（Eric 要求，2026-11-06）：图片样式与在线艺术家页的 ListDetail 一致，
圆角 8px，头部底部内边距 24px，所以边长是 204px；非正方形的图片居中裁切。
列表里的艺术家卡片仍是圆形。
不迁入别名、身份、视频、关注、分享与源页面。“更多 → 批量操作”（阶段 7，Eric 确认 2026-11-04）
照搬 SPlayer 只在“歌曲”标签下显示；与 SPlayer 不同，点开时按当前排序取这位艺术家的全部歌曲，
而不是只用已载入的部分。
简介（2026-11-04 起）照搬 SPlayer，内容由管理后台填写。
与 SPlayer 不同（Eric，2026-11-09）：列表滚动改用歌单/专辑页同款 useListScroll。SPlayer 这里的
listScroll 只看 scrollTop > 10，歌曲刚好比可视区高一点（溢出不足收缩让出的 120px）时，收缩后
内容不再溢出，scrollTop 被夹回 0 又展开，头部原地反复抽动；useListScroll 在溢出不足 150px 时不进入
收缩状态。 -->
<template>
  <div :key="artistId" :class="['artist', { small: listScrolling }]">
    <Transition name="fade" mode="out-in">
      <div v-if="artistDetailData" class="detail">
        <div v-if="!settingStore.hiddenCovers.artistDetail" class="cover">
          <n-image
            :src="coverSrc"
            object-fit="cover"
            :previewed-img-props="{ style: { borderRadius: '8px' } }"
            :preview-src="previewSrc"
            :preview-disabled="!artistDetailData.image"
            :renderToolbar="artistImageToolbar"
            show-toolbar-tooltip
            class="cover-img"
            @load="coverLoaded"
          >
            <template #placeholder>
              <div class="cover-loading">
                <img
                  src="/images/artist.jpg?asset"
                  class="loading-img"
                  alt="loading-img"
                />
              </div>
            </template>
          </n-image>
          <!-- 封面背板 -->
          <n-image
            class="cover-shadow"
            preview-disabled
            object-fit="cover"
            :src="coverSrc"
          />
        </div>
        <div class="data">
          <div class="name text-hidden">
            <n-text class="name-text">{{
              settingStore.hideBracketedContent
                ? removeBrackets(artistDetailData.name)
                : artistDetailData.name || "未知艺术家"
            }}</n-text>
          </div>
          <n-collapse-transition :show="!listScrolling" class="collapse">
            <!-- 信息 -->
            <n-flex class="meta">
              <div
                class="item"
                @click="
                  router.push({ name: 'artist-songs', query: { id: artistId } })
                "
              >
                <SvgIcon name="Music" :depth="3" />
                <n-text>{{ artistDetailData.trackCount || 0 }}</n-text>
              </div>
              <div
                class="item"
                @click="
                  router.push({
                    name: 'artist-albums',
                    query: { id: artistId },
                  })
                "
              >
                <SvgIcon name="Album" :depth="3" />
                <n-text>{{ artistDetailData.albumCount || 0 }}</n-text>
              </div>
            </n-flex>
            <!-- 简介 -->
            <n-text
              v-if="artistDetailData.description"
              class="description text-hidden"
              @click="openDescModal(artistDetailData.description, '歌手简介')"
            >
              {{ artistDetailData.description }}
            </n-text>
          </n-collapse-transition>
          <n-flex class="menu" justify="space-between">
            <n-flex class="left" align="flex-end">
              <n-button
                :focusable="false"
                type="primary"
                strong
                secondary
                round
                @click="playAllSongs"
              >
                <template #icon>
                  <SvgIcon name="Play" />
                </template>
                播放
              </n-button>
              <!-- 更多操作 -->
              <n-dropdown
                v-if="moreOptions.some((option) => option.show !== false)"
                :options="moreOptions"
                trigger="click"
                placement="bottom-start"
              >
                <n-button
                  :focusable="false"
                  :loading="batchLoading"
                  class="more"
                  circle
                  strong
                  secondary
                >
                  <template #icon>
                    <SvgIcon name="List" />
                  </template>
                </n-button>
              </n-dropdown>
            </n-flex>
          </n-flex>
        </div>
      </div>
      <ListError
        v-else-if="error"
        :error="error"
        :missing="missing"
        @retry="getArtistDetail(artistId)"
      />
      <div v-else class="detail">
        <n-skeleton
          v-if="!settingStore.hiddenCovers.artistDetail"
          class="cover"
        />
        <div class="data">
          <n-skeleton :repeat="4" text />
        </div>
      </div>
    </Transition>
    <!-- 标签页 -->
    <n-tabs
      v-if="!error"
      v-model:value="artistType"
      class="tabs"
      type="segment"
      @update:value="tabChange"
    >
      <n-tab name="artist-songs"> 歌曲 </n-tab>
      <n-tab name="artist-albums"> 专辑 </n-tab>
    </n-tabs>
    <!-- 路由 -->
    <RouterView v-if="!error" v-slot="{ Component }">
      <Transition :name="`router-${settingStore.routeAnimation}`" mode="out-in">
        <KeepAlive v-if="settingStore.useKeepAlive">
          <component
            ref="componentRef"
            :is="Component"
            :id="artistId"
            class="router-view"
            @scroll="listScroll"
          />
        </KeepAlive>
        <component
          v-else
          ref="componentRef"
          :is="Component"
          :id="artistId"
          class="router-view"
          @scroll="listScroll"
        />
      </Transition>
    </RouterView>
  </div>
</template>

<script setup lang="ts">
import type { DropdownOption } from "naive-ui";
import type { Artist } from "@/api/types";
import { fetchAll } from "@/api/pages";
import { listSortQuery } from "@/utils/meta";
import { source } from "@/api/source";
import { useLoadError } from "@/composables/List/useLoadError";
import { useListScroll } from "@/composables/List/useListScroll";
import { coverLoaded, downloadImage, renderIcon } from "@/utils/helper";
import { fillArtist } from "@/utils/onlineFill";
import { onlinePageRoute } from "@/utils/track";
import { renderDownloadToolbar } from "@/utils/meta";
import { removeBrackets } from "@/utils/format";
import { coverUrl, defaultCovers } from "@/utils/cover";
import { useSettingStore, useStatusStore, useUserStore } from "@/stores";
import { openBatchList, openDescModal } from "@/utils/modal";
import ListError from "@/components/List/ListError.vue";
import ArtistSongs from "./songs.vue";

const route = useRoute();
const router = useRouter();
const settingStore = useSettingStore();
const statusStore = useStatusStore();
const userStore = useUserStore();

// 路由元素
const componentRef = ref<InstanceType<typeof ArtistSongs> | null>(null);

// 歌手 ID
const artistId = computed<number>(() => Number(route.query.id));

// 歌手分类
const artistType = ref<string>((route.name as string) || "artist-songs");

// 歌手数据
const artistDetailData = ref<Artist | null>(null);
const { error, missing, setError, clearError } = useLoadError(
  "艺术家不存在或已被删除",
  "获取歌手详情失败",
);

// 图片：SPlayer 的 coverSize.m，预览原图
const coverSrc = computed(() =>
  coverUrl(artistDetailData.value?.image, "m", defaultCovers.artist),
);
const previewSrc = computed(() =>
  coverUrl(artistDetailData.value?.image, "cover", defaultCovers.artist),
);
const artistImageToolbar = renderDownloadToolbar(() =>
  downloadImage(previewSrc.value, artistDetailData.value?.name || "artist"),
);

// 列表是否滚动。溢出不足以支撑收缩时不收缩，避免收缩 -> 不溢出 -> 展开的反复抽动
const { listScrolling, handleListScroll: listScroll } = useListScroll();

// 获取歌手详情
const getArtistDetail = async (id: number) => {
  try {
    if (!id) return;
    listScrolling.value = false;
    artistDetailData.value = null;
    clearError();
    artistDetailData.value = await source.artist(id);
  } catch (e) {
    console.error("Erorr getting artist detail:", e);
    setError(e);
  }
};

// Tabs 改变
const tabChange = (value: string) => {
  router.push({
    name: value,
    query: { id: artistId.value },
  });
};

// 播放全部歌曲
const playAllSongs = async () => {
  await router.push({ name: "artist-songs", query: { id: artistId.value } });
  if (componentRef.value) componentRef.value.playAllSongs();
};

// 批量操作：按当前排序取这位艺术家的全部歌曲后打开弹窗
const batchLoading = ref<boolean>(false);
const openBatch = async () => {
  const id = artistId.value;
  if (!id || batchLoading.value) return;
  batchLoading.value = true;
  try {
    const songs = await fetchAll((offset, limit) =>
      source.tracks({
        artistId: id,
        offset,
        limit,
        ...listSortQuery(statusStore.listSortField, statusStore.listSortOrder),
      }),
    );
    if (artistId.value === id) await openBatchList(songs);
  } catch (e) {
    console.error("获取艺术家歌曲失败：", e);
    window.$message.error("获取歌曲失败，请重试");
  } finally {
    batchLoading.value = false;
  }
};

// 补全图片与简介
const fillProfile = async () => {
  const artist = artistDetailData.value;
  if (!artist) return;
  const updated = await fillArtist(artist);
  if (updated && artistDetailData.value?.id === updated.id)
    artistDetailData.value = updated;
};

// 更多操作
// 照搬 SPlayer，批量操作只在歌曲标签下。与 SPlayer 不同：关联了在线艺术家时多“查看在线艺术家”，
// 两个标签下都有，所以有它时专辑标签下也显示“更多”；有导入权限且图片或简介为空时多
// “从网易云补全资料”（Eric，2026-11-07）
const moreOptions = computed<DropdownOption[]>(() => {
  const artist = artistDetailData.value;
  const onlineId = artist?.onlineId;
  return [
    {
      label: "查看在线艺术家",
      key: "open-online",
      show: !!onlineId,
      props: {
        onClick: () =>
          onlineId && router.push(onlinePageRoute("artist", onlineId)),
      },
      icon: renderIcon("Cloud"),
    },
    {
      label: "从网易云补全资料",
      key: "fill-profile",
      show:
        !!userStore.user?.canImport &&
        !!onlineId &&
        !!artist &&
        (!artist.image || !artist.description),
      props: { onClick: fillProfile },
      icon: renderIcon("EditNote"),
    },
    {
      label: "批量操作",
      key: "batch",
      show: artistType.value === "artist-songs",
      props: { onClick: openBatch },
      icon: renderIcon("Batch"),
    },
  ];
});

// 监听路由更新
onBeforeRouteUpdate((to) => {
  listScrolling.value = false;
  // 检查是否仍在 artist 路由下
  const isArtistRoute = to.matched.some((m) => m.name === "artist");
  if (!isArtistRoute) return;
  artistType.value = to.name as string;
});

// 监听 ID 变化
watch(
  () => artistId.value,
  (val) => {
    if (val) getArtistDetail(val);
  },
  { immediate: true },
);
</script>

<style lang="scss" scoped>
.artist {
  display: flex;
  flex-direction: column;
  height: 100%;
  .detail {
    display: flex;
    height: 240px;
    width: 100%;
    padding: 12px 0 24px 0;
    will-change: height, opacity;
    z-index: 1;
    transition:
      height 0.3s,
      opacity 0.3s;
    .cover {
      position: relative;
      display: flex;
      width: auto;
      height: 100%;
      aspect-ratio: 1 / 1;
      margin-right: 20px;
      border-radius: 8px;
      transition:
        opacity 0.3s,
        margin 0.3s,
        transform 0.3s;
      :deep(img) {
        width: 100%;
        height: 100%;
        opacity: 0;
        transition: opacity 0.35s ease-in-out;
      }
      .cover-img {
        border-radius: 8px;
        overflow: hidden;
        z-index: 1;
        transition:
          opacity 0.3s,
          filter 0.3s,
          transform 0.3s;
      }
      .cover-shadow {
        position: absolute;
        top: 6px;
        height: 100%;
        width: 100%;
        filter: blur(12px) opacity(0.6);
        transform: scale(0.92, 0.96);
        z-index: 0;
        background-size: cover;
        aspect-ratio: 1/1;
        :deep(img) {
          opacity: 1;
        }
      }
      &:active {
        transform: scale(0.98);
      }
    }
    .data {
      position: relative;
      display: flex;
      flex-direction: column;
      flex: 1;
      padding-right: 60px;
      :deep(.n-skeleton) {
        height: 30px;
        margin-top: 12px;
        border-radius: 8px;
        &:first-child {
          width: 60%;
          margin-top: 0;
          height: 40px;
        }
      }
      .description {
        margin-bottom: 8px;
        padding-left: 4px;
        cursor: pointer;
      }
      .name {
        font-size: 30px;
        font-weight: bold;
        height: 48px;
        transition:
          font-size 0.3s var(--n-bezier),
          color 0.3s var(--n-bezier);
        .name-alias {
          &::before {
            content: "（";
            margin-right: 6px;
          }
          &::after {
            content: "）";
            margin-left: 6px;
          }
        }
      }
      .identify {
        font-size: 16px;
        margin-bottom: 8px;
        padding-left: 4px;
      }
      .collapse {
        position: absolute;
        top: 48px;
        margin: 8px 0;
      }
      .meta {
        margin-bottom: 8px;
        .item {
          display: flex;
          align-items: center;
          cursor: pointer;
          .n-icon {
            font-size: 20px;
            margin-right: 4px;
          }
        }
      }
      .menu {
        position: absolute;
        left: 0;
        bottom: 0;
        width: 100%;
        .n-button {
          height: 40px;
          transition: all 0.3s var(--n-bezier);
        }
        .more {
          width: 40px;
        }
      }
    }
  }
  .tabs {
    height: 40px;
    z-index: 1;
  }
  .router-view {
    flex: 1;
    overflow: hidden;
    &.artist-songs {
      position: absolute;
      width: 100%;
      height: 100%;
      padding-top: 280px;
      transition:
        padding 0.3s,
        transform 0.3s,
        opacity 0.3s;
    }
  }
  &.small {
    .detail {
      height: 120px;
      .cover {
        margin-right: 12px;
        .cover-mask,
        .play-count {
          opacity: 0;
        }
      }
      .data {
        .name {
          font-size: 22px;
        }
        .menu {
          .n-button,
          .search {
            height: 32px;
            --n-font-size: 13px;
            --n-padding: 0 14px;
            --n-icon-size: 16px;
          }
        }
      }
    }
    .router-view {
      &.artist-songs {
        padding-top: 160px;
      }
    }
  }
}
</style>
