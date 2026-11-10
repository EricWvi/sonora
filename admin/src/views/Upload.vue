<script setup lang="ts">
import { computed, h, reactive } from "vue";
import { RouterLink } from "vue-router";
import {
  NButton,
  NDataTable,
  NP,
  NProgress,
  NSpace,
  NTag,
  NText,
  NUpload,
  NUploadDragger,
  type DataTableColumns,
  type UploadFileInfo,
} from "naive-ui";
import {
  artistNames,
  formatBytes,
  type ApiError,
  type TrackDetail,
} from "../api";
import { isLyrics, lyricsRank, stem, uploadTrack } from "../upload";

type State =
  "queued" | "uploading" | "processing" | "created" | "duplicate" | "failed";

interface Item {
  key: number;
  file: File | null;
  name: string;
  size: number;
  lyrics: File | null;
  state: State;
  progress: number;
  error: string;
  track: TrackDetail | null;
}

const CONCURRENCY = 2;
let nextKey = 1;
const items = reactive<Item[]>([]);
const counts = computed(() => {
  const c = { done: 0, duplicate: 0, failed: 0, pending: 0 };
  for (const it of items) {
    if (it.state === "created") c.done++;
    else if (it.state === "duplicate") c.duplicate++;
    else if (it.state === "failed") c.failed++;
    else c.pending++;
  }
  return c;
});

// Files picked together are paired by name: "song.lrc", "song.yrc",
// "song.qrc" or "song.ttml" goes with "song.flac". A song has one lyrics
// file; TTML wins over YRC and QRC, those over LRC.
// The service decides what counts as audio.
function add(files: File[]) {
  const audio = files.filter((f) => !isLyrics(f.name));
  const lyrics = new Map<string, File>();
  const spare: File[] = [];
  for (const f of files.filter((f) => isLyrics(f.name))) {
    const other = lyrics.get(stem(f.name));
    if (other && lyricsRank(f.name) <= lyricsRank(other.name)) {
      spare.push(f);
      continue;
    }
    if (other) spare.push(other);
    lyrics.set(stem(f.name), f);
  }
  for (const f of audio) {
    const lrc = lyrics.get(stem(f.name)) ?? null;
    if (lrc) lyrics.delete(stem(f.name));
    items.push({
      key: nextKey++,
      file: f,
      name: f.name,
      size: f.size,
      lyrics: lrc,
      state: "queued",
      progress: 0,
      error: "",
      track: null,
    });
  }
  for (const f of spare) {
    items.push({
      key: nextKey++,
      file: null,
      name: f.name,
      size: f.size,
      lyrics: null,
      state: "failed",
      progress: 0,
      error: "已优先使用同名的另一份歌词",
      track: null,
    });
  }
  for (const f of lyrics.values()) {
    items.push({
      key: nextKey++,
      file: null,
      name: f.name,
      size: f.size,
      lyrics: null,
      state: "failed",
      progress: 0,
      error: "没有同时选择同名的音频文件",
      track: null,
    });
  }
  pump();
}

let batch: File[] = [];
let batchTimer = 0;
// NUpload reports files one by one; collect a whole selection before
// pairing lyrics with audio.
function onChange(opts: { file: UploadFileInfo }) {
  if (opts.file.file) batch.push(opts.file.file);
  clearTimeout(batchTimer);
  batchTimer = window.setTimeout(() => {
    const files = batch;
    batch = [];
    add(files);
  }, 50);
}

function pump() {
  let running = items.filter(
    (i) => i.state === "uploading" || i.state === "processing",
  ).length;
  for (const it of items) {
    if (running >= CONCURRENCY) return;
    if (it.state !== "queued") continue;
    running++;
    run(it);
  }
}

async function run(it: Item) {
  it.state = "uploading";
  it.progress = 0;
  try {
    const res = await uploadTrack(it.file!, it.lyrics, (p) => {
      it.progress = p;
      if (p >= 1) it.state = "processing";
    });
    it.state = res.status;
    it.track = res.track;
    it.file = null;
  } catch (err) {
    it.state = "failed";
    it.error = (err as ApiError).message;
  }
  pump();
}

function retry(it: Item) {
  if (!it.file) return;
  it.state = "queued";
  it.error = "";
  pump();
}

function clearFinished() {
  for (let i = items.length - 1; i >= 0; i--) {
    if (["created", "duplicate"].includes(items[i].state)) items.splice(i, 1);
  }
}

const stateTag: Record<
  State,
  { label: string; type: "default" | "info" | "success" | "warning" | "error" }
> = {
  queued: { label: "等待", type: "default" },
  uploading: { label: "上传中", type: "info" },
  processing: { label: "入库中", type: "info" },
  created: { label: "已入库", type: "success" },
  duplicate: { label: "重复", type: "warning" },
  failed: { label: "失败", type: "error" },
};

const columns: DataTableColumns<Item> = [
  {
    title: "文件",
    key: "name",
    ellipsis: { tooltip: true },
    render: (it) =>
      h("span", [
        it.name,
        it.lyrics
          ? h(
              NTag,
              { size: "small", bordered: false, style: "margin-left: 8px" },
              () => "歌词",
            )
          : null,
      ]),
  },
  {
    title: "大小",
    key: "size",
    width: 100,
    render: (it) => formatBytes(it.size),
  },
  {
    title: "状态",
    key: "state",
    width: 220,
    render: (it) => {
      if (it.state === "uploading") {
        return h(NProgress, {
          type: "line",
          percentage: Math.round(it.progress * 100),
          indicatorPlacement: "inside",
        });
      }
      const t = stateTag[it.state];
      return h(
        NTag,
        { type: t.type, size: "small", bordered: false },
        () => t.label,
      );
    },
  },
  {
    title: "结果",
    key: "result",
    ellipsis: { tooltip: true },
    render: (it) => {
      if (it.state === "failed") {
        return h(NSpace, { align: "center", wrap: false }, () => [
          h(NText, { type: "error" }, () => it.error),
          it.file
            ? h(
                NButton,
                { size: "tiny", onClick: () => retry(it) },
                () => "重试",
              )
            : null,
        ]);
      }
      if (!it.track) return null;
      const t = it.track;
      const label = [t.title, artistNames(t.artists), t.album?.title]
        .filter(Boolean)
        .join(" · ");
      const prefix = it.state === "duplicate" ? "曲库中已有：" : "";
      return h(
        RouterLink,
        { to: `/tracks?edit=${t.id}` },
        () => prefix + label,
      );
    },
  },
];
</script>

<template>
  <n-space vertical :size="16">
    <n-upload
      multiple
      directory-dnd
      :default-upload="false"
      :show-file-list="false"
      accept="audio/*,.mp3,.flac,.m4a,.ogg,.opus,.lrc,.ttml,.yrc,.qrc"
      @change="onChange"
    >
      <n-upload-dragger>
        <n-p style="margin: 0; font-size: 16px">点击或拖入音频文件</n-p>
        <n-text depth="3"
          >MP3、FLAC、AAC（M4A）、Ogg Vorbis、Opus；同名 .lrc、.yrc、.qrc 或
          .ttml 作为歌词一并上传</n-text
        >
      </n-upload-dragger>
    </n-upload>
    <template v-if="items.length">
      <n-space align="center" justify="space-between">
        <n-text depth="3">
          已入库 {{ counts.done }} · 重复 {{ counts.duplicate }} · 失败
          {{ counts.failed }} · 进行中
          {{ counts.pending }}
        </n-text>
        <n-button
          size="small"
          :disabled="!counts.done && !counts.duplicate"
          @click="clearFinished"
        >
          清除已完成
        </n-button>
      </n-space>
      <n-data-table
        :columns="columns"
        :data="items"
        :row-key="(it: Item) => it.key"
        size="small"
      />
    </template>
  </n-space>
</template>
