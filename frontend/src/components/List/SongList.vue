<!-- 歌曲列表 - 虚拟列表。照搬 SPlayer components/List/SongList.vue，歌曲是播放端 API 的 Track。
排序由服务端完成：排序菜单只修改 status store 中的排序字段与顺序，页面据此重新请求。
不迁入电台、大小列与移动端菜单。
与 SPlayer 不同（Eric，2026-11-08）：未传 height 时内联高度改为命令式写入。SPlayer 用
useElementSize 测 .song-list 高度并绑定进模板，详情页折叠动画期间测量值每帧变化，导致整棵
SongList 每帧重渲染（一次折叠约 256 次 SongCard 更新）而掉帧；内联高度本身是承重的（内容多于
视口的页面靠它约束 .song-list，纯 CSS height:100% 会跟着内容增长撑破布局），不能去掉，改为在
ResizeObserver 回调里直接写 wrapper 的 style，测量不再经过响应式。模板的样式对象里不出现
height 这个 key：patchStyle 每次重渲染都会写入样式对象的每个 key（undefined 写成空串），
'0px' 会把高度卡死在 0，undefined 会清空退回 height:100%，只有不出现才不被覆盖。
传了数值 height 的页面仍由模板绑定（折叠时靠 SCSS 的 height 过渡平滑），height="auto" 不变。 -->
<template>
  <Transition name="fade">
    <div v-if="!isEmpty(listData)" ref="songListRef" class="song-list">
      <Transition name="fade">
        <div
          ref="virtualWrapperRef"
          :key="listKey"
          :style="wrapperStyle"
          class="virtual-list-wrapper"
        >
          <!-- 悬浮顶栏 -->
          <div class="list-header song-card sticky-header">
            <n-text class="num">#</n-text>
            <n-popover
              v-if="!disabledSort"
              trigger="click"
              placement="bottom-start"
              :show-arrow="false"
              style="padding: 0"
            >
              <template #trigger>
                <div class="title has-sort">
                  <n-text>标题</n-text>
                  <n-text
                    v-if="statusStore.listSortField !== 'default'"
                    class="sort"
                    depth="3"
                  >
                    {{ sortFieldOptions[statusStore.listSortField].name }}
                  </n-text>
                </div>
              </template>
              <div class="sort-menu">
                <div class="group">
                  <div class="label">排序字段</div>
                  <n-radio-group
                    v-model:value="statusStore.listSortField"
                    name="sortField"
                    @update:value="handleSortFieldChange"
                  >
                    <n-flex :vertical="true" size="small">
                      <n-radio
                        v-for="(option, key) in sortFieldOptions"
                        :key="key"
                        :value="key"
                      >
                        {{ option.name }}
                      </n-radio>
                    </n-flex>
                  </n-radio-group>
                </div>
                <n-divider vertical style="height: auto; margin: 0 12px" />
                <div class="group">
                  <div class="label">排序方式</div>
                  <n-radio-group
                    v-model:value="statusStore.listSortOrder"
                    name="sortOrder"
                    @update:value="handleSortOrderChange"
                  >
                    <n-flex :vertical="true" size="small">
                      <n-radio
                        v-for="(option, key) in sortOrderOptions"
                        :key="key"
                        :value="key"
                      >
                        {{ option.name }}
                      </n-radio>
                    </n-flex>
                  </n-radio-group>
                </div>
              </div>
            </n-popover>
            <n-text v-else class="title">标题</n-text>
            <n-text
              v-if="
                !hiddenAlbum && !isSmallScreen && settingStore.showSongAlbum
              "
              class="album"
            >
              专辑
            </n-text>
            <n-text v-if="settingStore.showSongOperations" class="actions">
              操作
            </n-text>
            <n-text
              v-if="!isSmallScreen && settingStore.showSongDuration"
              class="meta"
            >
              时长
            </n-text>
          </div>
          <!-- 虚拟列表 -->
          <VirtualScroll
            ref="listRef"
            :item-height="90"
            :item-fixed="!discGrouped"
            :items="virtualListItems"
            :height="`calc(100% - 40px)`"
            :padding-bottom="80"
            :class="{ 'is-dragging-global': isDragging && draggable }"
            @scroll="onScroll"
          >
            <template #default="{ item, index }">
              <div v-if="item.type === 'song'" class="song-node">
                <!-- 拖拽放置指示线 -->
                <div
                  v-if="
                    isDragging &&
                    draggable &&
                    dropIndicator.index === index &&
                    dropIndicator.position === 'top'
                  "
                  class="drop-line line-top"
                />
                <div
                  v-if="
                    isDragging &&
                    draggable &&
                    dropIndicator.index === index &&
                    dropIndicator.position === 'bottom'
                  "
                  class="drop-line line-bottom"
                />
                <SongCard
                  :class="{
                    'is-dragging':
                      isDragging && draggable && draggedIndex === index,
                  }"
                  :song="item.data"
                  :index="item.index"
                  :hiddenCover="hiddenCover || settingStore.hiddenCovers.list"
                  :hiddenAlbum="hiddenAlbum"
                  @mousedown="
                    draggable
                      ? handlePointerDown(
                          $event,
                          index,
                          item.data.title || '未知曲目',
                        )
                      : undefined
                  "
                  @touchstart="
                    draggable
                      ? handlePointerDown(
                          $event,
                          index,
                          item.data.title || '未知曲目',
                        )
                      : undefined
                  "
                  @dblclick.stop="handleSongPlay(item.data)"
                  @contextmenu.stop="handleShowMenu($event, item.data, index)"
                />
              </div>
              <!-- 加载更多 -->
              <!-- 碟号分组标题（Sonora）：多碟专辑在默认顺序下按碟号分组，样式沿用表头 -->
              <div v-else-if="item.type === 'disc'" class="disc-header">
                <div class="num"><SvgIcon :size="20" name="Album" /></div>
                <n-text>第 {{ item.disc }} 碟</n-text>
              </div>
              <div v-else-if="item.type === 'footer'" class="load-more">
                <n-flex v-if="loadMore && loading">
                  <n-spin size="small" />
                  <n-text>{{ loadingText || "努力加载中" }}</n-text>
                </n-flex>
                <n-divider v-else dashed> 没有更多啦 ~ </n-divider>
              </div>
            </template>
          </VirtualScroll>
        </div>
      </Transition>
      <!-- 右键菜单 -->
      <SongListMenu
        ref="songListMenuRef"
        :hiddenCover="hiddenCover || settingStore.hiddenCovers.list"
        @removeSong="removeSong"
      />
      <!-- 列表操作 -->
      <Teleport to="body">
        <Transition name="fade" mode="out-in">
          <div v-if="floatToolShow" class="list-menu">
            <n-float-button-group position="relative">
              <n-float-button
                v-if="hasPlaySong >= 0"
                width="42"
                @click="scrollToCurrentSong"
              >
                <SvgIcon :size="22" name="Location" />
              </n-float-button>
              <n-float-button
                :class="{ hidden: scrollTop <= 100 }"
                width="42"
                @click="scrollToTop"
              >
                <SvgIcon :size="22" name="Up" />
              </n-float-button>
            </n-float-button-group>
          </div>
        </Transition>
      </Teleport>
      <!-- 拖拽浮动标签 -->
      <Teleport to="body">
        <Transition name="fade">
          <div
            v-if="isDragging && draggable && dragLabelData"
            class="drag-label"
            :style="{
              top: `${dragLabelPosition.top}px`,
              left: `${dragLabelPosition.left}px`,
            }"
          >
            <n-text class="drag-label-name">{{ dragLabelData.name }}</n-text>
          </div>
        </Transition>
      </Teleport>
    </div>
    <!-- 列表加载 - 骨架屏 -->
    <div v-else-if="loading" class="song-list loading">
      <n-skeleton :repeat="10" text />
    </div>
    <!-- 空列表 -->
    <n-empty
      v-else
      description="列表光秃秃的，啥都没有哦"
      size="large"
      class="song-list empty"
    />
  </Transition>
</template>

<script setup lang="ts">
import { sameTrack, trackKey } from "@/utils/track";
import type { Track } from "@/api/types";
import { useMusicStore, useStatusStore, useSettingStore } from "@/stores";
import { isEmpty } from "lodash-es";
import {
  sortFieldOptions,
  sortOrderOptions,
  type ListSortField,
} from "@/utils/meta";
import { usePlayerController } from "@/core/player/PlayerController";
import { useDragSort } from "@/composables/List/useDragSort";
import SongListMenu from "@/components/Menu/SongListMenu.vue";
import VirtualScroll from "@/components/UI/VirtualScroll.vue";

const props = withDefaults(
  defineProps<{
    /** 列表数据 */
    data: Track[];
    /** 列表高度 */
    height?: number | "auto";
    /** 是否加载 */
    loading?: boolean;
    /** 加载更多 */
    loadMore?: boolean;
    /** 加载文本 */
    loadingText?: string;
    /** 隐藏专辑 */
    hiddenAlbum?: boolean;
    /** 隐藏封面 */
    hiddenCover?: boolean;
    /** 隐藏滚动条 */
    hiddenScrollbar?: boolean;
    /** 禁用排序 */
    disabledSort?: boolean;
    /** 播放列表来源，如 `album-3`（见 music store 的 playPlaylistId） */
    playListId?: string;
    /** 所在的歌单；右键菜单据此显示“从歌单中删除” */
    userPlaylistId?: number;
    /** 双击播放操作 */
    doubleClickAction?: "all" | "add";
    /** 列表版本 */
    listVersion?: string | number;
    /** 禁用高度过渡动画 */
    disableHeightTransition?: boolean;
    /** 是否可拖拽排序 */
    draggable?: boolean;
    /**
     * 专辑详情：多于一张碟、默认排序时显示碟号标题，序号按碟重新开始；默认排序且每首歌都有
     * 曲目序号时，序号显示曲目序号
     */
    groupByDisc?: boolean;
  }>(),
  {
    loadingText: "努力加载中...",
    playListId: "",
    userPlaylistId: 0,
    listVersion: 0,
  },
);

const emit = defineEmits<{
  // 触底
  reachBottom: [e: Event];
  // 滚动
  scroll: [e: Event];
  // 删除歌曲
  removeSong: [id: number[]];
  // 拖拽重排序
  reorder: [fromIndex: number, toIndex: number];
}>();

const musicStore = useMusicStore();
const statusStore = useStatusStore();
const settingStore = useSettingStore();
const player = usePlayerController();
// SPlayer 的 useMobile：小屏（< 768px）隐藏专辑与时长；Sonora 最小宽度 990px，不会出现
const isSmallScreen = useMediaQuery("(max-width: 767.98px)");

// 列表元素
const listRef = ref<InstanceType<typeof VirtualScroll> | null>(null);
const songListRef = ref<HTMLElement | null>(null);

// 拖拽排序
const {
  isDragging,
  draggedIndex,
  dropIndicator,
  dragLabelData,
  dragLabelPosition,
  handlePointerDown,
} = useDragSort({
  virtualScrollRef: listRef,
  itemCount: computed(() => props.data.length),
  onReorder: (from, to) => emit("reorder", from, to),
  paddingTop: 0,
  triggerMode: "longpress",
  longPressDelay: 300,
});

// 处理歌曲播放
const handleSongPlay = (song: Track) => {
  if (props.doubleClickAction === "add") {
    player.addNextSong(song, true);
  } else {
    player.updatePlayList(listData.value, song, props.playListId);
  }
};

// 列表状态
const scrollTop = ref<number>(0);
const scrollIndex = ref<number>(0);

// 悬浮工具
const floatToolShow = ref<boolean>(true);

// 右键菜单
const songListMenuRef = ref<InstanceType<typeof SongListMenu> | null>(null);

const handleShowMenu = (e: MouseEvent, song: Track, index: number) => {
  songListMenuRef.value?.openDropdown(
    e,
    listData.value,
    song,
    index,
    props.userPlaylistId,
  );
};

// 列表数据：排序由服务端完成，按页面给出的顺序显示
const listData = computed<Track[]>(() => props.data);

// 虚拟列表项类型
type VirtualListItem =
  | { type: "song"; data: Track; key: number | string; index: number }
  | { type: "disc"; data: null; key: string; disc: number }
  | { type: "footer"; data: null; key: string };

// 是否按碟号分组：专辑默认顺序下歌曲已按碟号与曲目序号排列
const discGrouped = computed(
  () =>
    !!props.groupByDisc &&
    statusStore.listSortField === "default" &&
    new Set(listData.value.map((song) => song.discNo)).size > 1,
);

// 序号显示曲目序号（Sonora，与 SPlayer 不同）：专辑中的歌可能不全，按位置编号会与曲目序号
// 对不上；有歌没有曲目序号时仍按位置编号，免得号码重复或缺失
const trackNumbered = computed(
  () =>
    !!props.groupByDisc &&
    statusStore.listSortField === "default" &&
    listData.value.every((song) => song.trackNo > 0),
);

// 虚拟列表数据（包含歌曲、碟号标题和 footer）
const virtualListItems = computed<VirtualListItem[]>(() => {
  const items: VirtualListItem[] = [];
  let disc = -1;
  let discStart = 0;
  listData.value.forEach((song, i) => {
    if (discGrouped.value && song.discNo !== disc) {
      disc = song.discNo;
      discStart = i;
      items.push({ type: "disc", data: null, key: `disc-${disc}`, disc });
    }
    items.push({
      type: "song" as const,
      data: song,
      key: trackKey(song),
      // SongCard 显示 index + 1
      index: trackNumbered.value ? song.trackNo - 1 : i - discStart,
    });
  });
  // 如果有加载更多，添加 footer
  if (props.loadMore !== undefined) {
    items.push({
      type: "footer" as const,
      data: null,
      key: "footer",
    });
  }
  return items;
});

// 虚拟列表 key
const listKey = computed(() => {
  // 使用 playListId 作为主要 key
  if (props.playListId) {
    return `playlist-${props.playListId}-${statusStore.listSortField}-${statusStore.listSortOrder}`;
  }
  return `list-${props.listVersion}-${statusStore.listSortField}-${statusStore.listSortOrder}`;
});

// 列表是否具有播放歌曲（虚拟列表中的位置，分组时包含碟号标题）
const hasPlaySong = computed(() => {
  return virtualListItems.value.findIndex(
    (item) => item.type === "song" && sameTrack(item.data, musicStore.playSong),
  );
});

// 虚拟列表容器。未传 height 时由下面的 ResizeObserver 命令式写高度
const virtualWrapperRef = ref<HTMLElement | null>(null);

// wrapper 的内联样式。未传 height 时样式对象里不能出现 height 这个 key：
// Vue 的 patchStyle 每次重渲染都会无条件写入样式对象的每个 key，值为
// undefined/null 时写成空串（prev 里有、next 里为 null 的 key 同样清空），
// 所以 height: '0px' 会把命令式高度冲回 0，height: undefined 会把它清空、
// 退回 SCSS 的 height:100%（内容多于视口时被撑破）。只有完全不出现，
// patchStyle 才不碰它。
const wrapperStyle = computed(() => ({
  ...(props.height !== undefined && {
    height: props.height === "auto" ? "auto" : `${props.height}px`,
  }),
  transition: props.disableHeightTransition
    ? "transform 0.3s, opacity 0.3s"
    : undefined,
}));

// 列表元素高度测量。未传 height 时把测量值直接写进 wrapper 的内联高度，
// 不经过响应式：绑定进模板会让折叠动画期间每帧的测量触发整棵 SongList
// 重渲染（见文件头注释）。内联高度由这里独占（见 wrapperStyle）：
// virtualWrapperRef 绑定（挂载或换 key 重挂）时写入当前测量值，测量值为
// 0 时写 0px，打断“wrapper 高度跟着 .song-list 内容增长”的循环。height
// 从数值切回 undefined 时 patchStyle 先清空，post watch 随后补写。
// height="auto" 时停止测量，交给页面滚动。
const { height: songListHeight, stop: stopCalcHeight } =
  useElementSize(songListRef);
watch(
  [songListHeight, virtualWrapperRef, () => props.height],
  ([h, wrapper, height]) => {
    if (height !== undefined || !wrapper) return;
    wrapper.style.height = h ? `${h}px` : "0px";
  },
  { flush: "post" },
);

// 列表滚动
const onScroll = (e: Event) => {
  emit("scroll", e);
  const target = e.target as HTMLElement;
  const top = target.scrollTop;
  scrollTop.value = top;
  scrollIndex.value = Math.floor(top / 90);

  // 触底检测
  const scrollHeight = target.scrollHeight;
  const clientHeight = target.clientHeight;
  if (
    scrollHeight - top - clientHeight < 100 &&
    !props.loading &&
    props.loadMore
  ) {
    emit("reachBottom", e);
  }
};

// 滚动到顶部
const scrollToTop = () => {
  listRef.value?.scrollToIndex(0);
};

// 滚动到当前播放歌曲
const scrollToCurrentSong = () => {
  if (hasPlaySong.value < 0) return;
  // 内部滚动
  listRef.value?.scrollToIndex(hasPlaySong.value);
  // 自动高度时回退
  nextTick(() => {
    const el = listRef.value?.contentRef?.querySelector<HTMLElement>(
      `[data-index="${hasPlaySong.value}"]`,
    );
    if (el) el.scrollIntoView({ block: "center" });
  });
};

// 排序更改后回到顶部；页面按新的排序重新请求
const scrollAfterSort = () => {
  nextTick(() => listRef.value?.scrollToIndex(0));
};

// 排序更改
const handleSortFieldChange = (val: ListSortField) => {
  // 单击已选择的字段 -> 切换升降序
  if (statusStore.listSortField === val && val !== "default") {
    statusStore.listSortOrder =
      statusStore.listSortOrder === "asc" ? "desc" : "asc";
  } else {
    statusStore.listSortField = val;
    // 如果切换到具体的字段且当前顺序是默认，自动切换到降序(对于时间相关)或升序(对于文本)
    if (statusStore.listSortOrder === "default") {
      if (val === "createTime" || val === "duration") {
        statusStore.listSortOrder = "desc";
      } else {
        statusStore.listSortOrder = "asc";
      }
    }
  }
  scrollAfterSort();
};

const handleSortOrderChange = (val: "default" | "asc" | "desc") => {
  statusStore.listSortOrder = val;
  scrollAfterSort();
};

// 删除指定索引
const removeSong = (id: number[]) => emit("removeSong", id);

// keep-alive 处理
onBeforeRouteLeave(() => {
  floatToolShow.value = false;
});

onActivated(() => {
  floatToolShow.value = true;
  if (props.height === "auto") stopCalcHeight();
  if (scrollIndex.value > 0) {
    nextTick(() => {
      listRef.value?.scrollToIndex(scrollIndex.value);
    });
  }
});

onBeforeUnmount(() => {
  stopCalcHeight();
  floatToolShow.value = false;
});
</script>

<style lang="scss" scoped>
.song-list {
  height: 100%;
  border-radius: 12px 0 0 12px;
  overflow: hidden;
  // 离场时脱离文档流，避免新旧内容同时存在时布局跳动
  &.fade-leave-active {
    position: absolute;
    left: 0;
    right: 0;
    pointer-events: none;
  }
  .song-card {
    padding-bottom: 12px;
    // padding-right: 4px;
  }
  // 悬浮顶栏
  .list-header {
    width: 100%;
    height: 40px;
    display: flex;
    flex-direction: row;
    align-items: center;
    justify-content: space-between;
    padding: 8px 12px;
    // margin-right: 4px;
    border: 1px solid transparent;
    // background-color: var(--background-hex);
    .n-text {
      opacity: 0.6;
    }
    .num {
      position: relative;
      display: flex;
      justify-content: center;
      align-items: center;
      width: 40px;
      min-width: 40px;
      font-weight: bold;
      margin-right: 12px;
    }
    .title {
      position: relative;
      flex: 1;
      display: flex;
      align-items: center;
      padding: 4px 20px 4px 0;
      cursor: pointer;
      .sort {
        margin-left: 6px;
        &::after {
          content: " )";
        }
        &::before {
          content: "( ";
        }
      }
      &.has-sort {
        &::after {
          content: "";
          position: absolute;
          opacity: 0;
          top: 0;
          left: -8px;
          width: 100%;
          height: 100%;
          border-radius: 8px;
          background-color: rgba(var(--primary), 0.08);
          transition: opacity 0.3s;
        }
        &:hover {
          &::after {
            opacity: 1;
          }
        }
      }
    }
    .album {
      flex: 1;
      padding-right: 20px;
    }
    .actions {
      display: flex;
      align-items: center;
      justify-content: center;
      width: 40px;
    }
    .meta {
      width: 50px;
      text-align: center;
      &.size {
        width: 60px;
      }
      &.date {
        width: 80px;
      }
    }
  }
  .virtual-list-wrapper {
    height: 100%;
    position: relative;
    // 传了数值 height 时内联高度由页面驱动，折叠时靠这条 height 过渡平滑
    // （ListDetail 系页面）；disableHeightTransition 的内联 transition 会整体
    // 覆盖这里，未传 height 时内联由测量写入，height 过渡由命令式更新承担
    transition:
      height 0.3s,
      transform 0.3s,
      opacity 0.3s;
    // 离场时脱离文档流，避免新旧列表同时存在时布局跳动
    &.fade-leave-active {
      position: absolute;
      left: 0;
      right: 0;
      pointer-events: none;
    }
    .sticky-header {
      position: sticky;
      top: 0;
      z-index: 10;
    }
  }
  // 加载更多
  .load-more {
    display: flex;
    flex-direction: row;
    justify-content: center;
    margin: 20px 0 40px;
    .n-spin-body {
      --n-size: 20px;
    }
    .n-divider {
      margin: 0;
      font-size: 14px;
      opacity: 0.6;
    }
  }
  // 加载
  &.loading {
    margin-top: 20px;
    :deep(.n-skeleton) {
      height: 72px;
      margin-bottom: 12px;
      border-radius: 12px;
    }
  }
  // 空列表
  &.empty {
    margin-top: 60px;
  }
}
.list-menu {
  position: fixed;
  right: 40px;
  bottom: 120px;
  z-index: 10;
  pointer-events: none;
  .n-float-button {
    height: 42px;
    border: 1px solid rgba(var(--primary), 0.28);
    pointer-events: auto;
    transition: opacity 0.3s;
    &.hidden {
      opacity: 0;
      pointer-events: none;
    }
  }
}

// 碟号分组标题（Sonora）：沿用表头的高度与透明度，图标与文字对齐序号列与标题列
.disc-header {
  display: flex;
  align-items: center;
  height: 40px;
  padding: 8px 12px;
  margin-bottom: 12px;
  border: 2px solid transparent;
  font-weight: bold;
  opacity: 0.6;
  .num {
    display: flex;
    justify-content: center;
    align-items: center;
    width: 40px;
    min-width: 40px;
    margin-right: 12px;
  }
}

// 拖拽排序
.song-node {
  position: relative;

  .drop-line {
    position: absolute;
    left: 0;
    right: 0;
    height: 2px;
    background-color: var(--primary-hex);
    border-radius: 2px;
    z-index: 10;
    pointer-events: none;

    &.line-top {
      top: 0;
    }
    &.line-bottom {
      bottom: 0;
    }
  }

  :deep(.song-card) {
    transition:
      opacity 0.2s,
      transform 0.3s;

    &.is-dragging {
      opacity: 0.3;
      transform: scale(0.98);
      .song-content {
        border-color: rgba(var(--primary), 0.5);
      }
    }
  }
}

:deep(.is-dragging-global) {
  cursor: grabbing;

  * {
    cursor: grabbing;
  }

  .song-card {
    pointer-events: none;
  }
}

.drag-label {
  position: fixed;
  z-index: 9999;
  padding: 8px 16px;
  border-radius: 20px;
  background-color: rgba(var(--primary), 0.15);
  backdrop-filter: blur(8px);
  pointer-events: none;
  transform: translate(12px, 12px);

  max-width: 260px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  font-size: 14px;
  font-weight: 500;
  color: rgba(var(--text-color), 0.3);
}

.sort-menu {
  display: flex;
  padding: 12px;
  .group {
    display: flex;
    flex-direction: column;
    .label {
      font-size: 12px;
      opacity: 0.6;
      margin-bottom: 8px;
      padding-left: 4px;
    }
    .n-radio-group {
      width: 120px;
      .n-radio {
        --n-font-size: 13px;
        padding: 4px 8px;
        border-radius: 4px;
        transition: background-color 0.3s;
        &:hover {
          background-color: var(--n-color-target);
        }
      }
    }
  }
}
</style>
