<!-- 批量操作。照搬 SPlayer components/Modal/BatchList.vue：表格勾选、“高级筛选”按序号范围选择。
与 SPlayer 不同（阶段 7，Eric 确认 2026-11-04）：没有批量下载与删除本地歌曲；操作为
“添加到歌单”“下一首播放”（Sonora 新增），歌单页另有“删除选中的歌曲”，
我喜欢的音乐另有“取消喜欢”（Sonora 新增）。在线专辑另有“导入到曲库”，只导入选中的歌曲
（Sonora 新增，可以导入的用户才有，Eric 2026-11-13）。 -->
<template>
  <div class="batch-list">
    <n-data-table
      v-model:checked-row-keys="checkedRowKeys"
      :columns="columnsData"
      :data="tableData"
      :row-key="(row: DataType) => row.key"
      max-height="60vh"
      virtual-scroll
      @update:checked-row-keys="tableCheck"
    />
    <n-flex class="batch-footer" justify="space-between" align="center">
      <n-flex align="center">
        <n-text :depth="3" class="count">已选择 {{ checkCount }} 首</n-text>
        <n-popover trigger="click" placement="right">
          <template #trigger>
            <n-button tertiary> 高级筛选 </n-button>
          </template>
          <n-flex :wrap="false" align="center">
            <n-input-number
              v-model:value="startRange"
              class="range-input"
              placeholder="开始"
              :min="1"
              :max="songs.length"
              size="small"
            />
            <n-text>-</n-text>
            <n-input-number
              v-model:value="endRange"
              class="range-input"
              placeholder="结束"
              :min="1"
              :max="songs.length"
              size="small"
            />
            <n-button size="small" secondary @click="handleRangeSelect">
              选择
            </n-button>
          </n-flex>
        </n-popover>
      </n-flex>
      <n-flex class="menu">
        <!-- 从歌单中删除 -->
        <n-button
          v-if="playlistId"
          :disabled="!checkCount"
          type="error"
          strong
          secondary
          @click="handleDeleteFromPlaylist"
        >
          <template #icon>
            <SvgIcon name="Delete" />
          </template>
          删除选中的歌曲
        </n-button>
        <!-- 取消喜欢 -->
        <n-button
          v-if="liked"
          :disabled="!checkCount"
          :loading="unliking"
          type="error"
          strong
          secondary
          @click="handleUnlike"
        >
          <template #icon>
            <SvgIcon name="FavoriteBorder" />
          </template>
          取消喜欢
        </n-button>
        <!-- 导入到曲库：只导入选中的歌曲，确认与进度沿用导入专辑 -->
        <n-button
          v-if="importAlbumId"
          :disabled="!checkCount"
          type="primary"
          strong
          secondary
          @click="
            importAlbum(
              importAlbumId,
              checkSongData.map((song) => song.id),
            )
          "
        >
          <template #icon>
            <SvgIcon name="Download" />
          </template>
          导入到曲库
        </n-button>
        <!-- 下一首播放 -->
        <n-button
          :disabled="!checkCount"
          type="primary"
          strong
          secondary
          @click="player.addNextSongs(checkSongData)"
        >
          <template #icon>
            <SvgIcon name="PlayNext" />
          </template>
          下一首播放
        </n-button>
        <!-- 添加到歌单 -->
        <n-button
          :disabled="!checkCount"
          type="primary"
          strong
          secondary
          @click="openPlaylistAdd(checkSongData)"
        >
          <template #icon>
            <SvgIcon name="AddList" />
          </template>
          添加到歌单
        </n-button>
      </n-flex>
    </n-flex>
  </div>
</template>

<script setup lang="ts">
import type { DataTableColumns, DataTableRowKey } from "naive-ui";
import type { Track } from "@/api/types";
import { isOnline, trackKey } from "@/utils/track";
import { ApiError, loginUrl, source } from "@/api/source";
import { openPlaylistAdd } from "@/utils/modal";
import { deleteSongs } from "@/utils/auth";
import { importAlbum } from "@/utils/albumImport";
import { useDataStore } from "@/stores";
import { usePlayerController } from "@/core/player/PlayerController";

interface DataType {
  key: number;
  id: string;
  name: string;
  artists: string;
  album: string;
  // 原始数据
  origin: Track;
}

const props = defineProps<{
  data: Track[];
  // 用户歌单 ID：显示“删除选中的歌曲”
  playlistId?: number;
  // playlistId 是自己的在线歌单
  onlinePlaylist?: boolean;
  // 我喜欢的音乐：显示“取消喜欢”
  liked?: boolean;
  // 在线专辑 ID：显示“导入到曲库”
  importAlbumId?: number;
}>();

const emit = defineEmits<{
  // 已从当前列表移除的歌曲
  /** 移除的歌曲，见 trackKey */
  removed: [keys: string[]];
}>();

const player = usePlayerController();
const dataStore = useDataStore();

// 弹窗内的列表：移除歌曲后同步去掉对应行
const songs = shallowRef<Track[]>(props.data);

// 选中数据
const checkCount = ref<number>(0);
const checkSongData = ref<Track[]>([]);
const checkedRowKeys = ref<DataTableRowKey[]>([]);

// 范围选择
const startRange = ref<number | null>(null);
const endRange = ref<number | null>(null);

// 表头数据
const columnsData = computed<DataTableColumns<DataType>>(() => [
  { type: "selection" },
  { title: "#", key: "key", width: 80 },
  { title: "标题", key: "name", ellipsis: { tooltip: true } },
  { title: "歌手", key: "artists", ellipsis: { tooltip: true } },
  { title: "专辑", key: "album", ellipsis: { tooltip: true } },
]);

// 表格数据
const tableData = computed<DataType[]>(() =>
  songs.value.map((song, index) => ({
    key: index + 1,
    id: trackKey(song),
    // 与 SPlayer 不同：在线歌曲标题后加云朵（Eric 要求，2026-11-06）
    name: `${song.title || "未知曲目"}${isOnline(song) ? " ☁️" : ""}`,
    artists: song.artists.length
      ? song.artists.map((ar) => ar.name).join(" / ")
      : "未知歌手",
    album: song.album?.title || "未知专辑",
    origin: song,
  })),
);

// 按勾选的行更新选中歌曲
const applyCheck = (keys: DataTableRowKey[]) => {
  checkedRowKeys.value = keys;
  checkCount.value = keys.length;
  checkSongData.value = tableData.value
    .filter((row) => keys.includes(row.key))
    .map((row) => row.origin);
};

// 表格勾选
const tableCheck = (keys: DataTableRowKey[]) => applyCheck(keys);

// 范围选择处理
const handleRangeSelect = () => {
  if (startRange.value === null || endRange.value === null) {
    window.$message.warning("请输入起始和结束序号");
    return;
  }
  const start = Math.max(1, Math.min(startRange.value, songs.value.length));
  const end = Math.max(1, Math.min(endRange.value, songs.value.length));
  if (start > end) {
    window.$message.warning("起始序号不能大于结束序号");
    return;
  }
  applyCheck(tableData.value.slice(start - 1, end).map((row) => row.key));
};

// 移除歌曲后去掉对应行；序号随之变化，保留其余歌曲的勾选
const afterRemoved = (ids: string[]) => {
  emit("removed", ids);
  const kept = checkSongData.value.filter(
    (song) => !ids.includes(trackKey(song)),
  );
  songs.value = songs.value.filter((song) => !ids.includes(trackKey(song)));
  applyCheck(
    tableData.value
      .filter((row) => kept.includes(row.origin))
      .map((row) => row.key),
  );
};

// 从歌单中删除（确认与提示沿用歌曲菜单的 deleteSongs）
const handleDeleteFromPlaylist = () => {
  if (!props.playlistId) return;
  const selected = checkSongData.value;
  deleteSongs(
    props.playlistId,
    selected.map((song) => song.id),
    {
      songName: selected.length === 1 ? selected[0].title : undefined,
      callback: () => afterRemoved(selected.map(trackKey)),
      online: props.onlinePlaylist,
    },
  );
};

// 取消喜欢：逐首请求，中途失败时只移除已成功的
const unliking = ref<boolean>(false);
const handleUnlike = () => {
  const selected = checkSongData.value;
  window.$dialog.warning({
    title: "取消喜欢",
    content:
      selected.length > 1
        ? `确定取消喜欢选中的 ${selected.length} 首歌曲吗？`
        : `确定取消喜欢歌曲 ${selected[0]?.title || ""} 吗？`,
    positiveText: "取消喜欢",
    negativeText: "取消",
    onPositiveClick: async () => {
      unliking.value = true;
      const done: Track[] = [];
      try {
        for (const song of selected) {
          await (isOnline(song)
            ? source.onlineUnlike(song.id)
            : source.unlike(song.id));
          done.push(song);
        }
      } catch (error) {
        if (error instanceof ApiError && error.status === 401) {
          location.assign(loginUrl);
          return;
        }
        window.$message.error(
          error instanceof ApiError ? error.message : "取消喜欢失败，请重试",
        );
        console.error("❌ 批量取消喜欢失败:", error);
      } finally {
        unliking.value = false;
      }
      if (!done.length) return;
      const doneKeys = done.map(trackKey);
      const { songs, onlineSongs } = dataStore.userLikeData;
      dataStore.userLikeData = {
        songs: songs.filter((id) => !doneKeys.includes(String(id))),
        onlineSongs: onlineSongs.filter(
          (id) => !doneKeys.includes(`online:${id}`),
        ),
      };
      afterRemoved(doneKeys);
      if (done.length === selected.length)
        window.$message.success("已取消喜欢");
    },
  });
};
</script>

<style lang="scss" scoped>
.batch-footer {
  margin-top: 20px;
}
.range-input {
  width: 90px;
}
</style>
