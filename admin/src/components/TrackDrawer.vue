<script setup lang="ts">
import { reactive, ref, watch } from "vue";
import {
  NButton,
  NDescriptions,
  NDescriptionsItem,
  NDivider,
  NDrawer,
  NDrawerContent,
  NForm,
  NFormItem,
  NGi,
  NGrid,
  NInput,
  NInputNumber,
  NSpace,
  NSpin,
  useDialog,
  useMessage,
} from "naive-ui";
import {
  audioURL,
  formatBytes,
  formatDuration,
  request,
  type ApiError,
  type Ref,
  type TrackDetail,
  type TrackInput,
} from "../api";
import ArtistsInput from "./ArtistsInput.vue";
import NeteaseIdInput from "./NeteaseIdInput.vue";
import CoverEditor from "./CoverEditor.vue";

const props = defineProps<{ id: number | null }>();
const emit = defineEmits<{ close: []; changed: [] }>();

const message = useMessage();
const dialog = useDialog();
const track = ref<TrackDetail | null>(null);
const loadError = ref("");
const form = reactive<TrackInput>({
  title: "",
  artists: [],
  album: "",
  albumArtists: [],
  trackNo: 0,
  discNo: 0,
  year: 0,
  neteaseId: null,
});
const lyrics = ref("");
const saving = ref(false);
const savingLyrics = ref(false);
const lyricsInput = ref<HTMLInputElement | null>(null);

// Album identity ignores artist order, so this does too.
const sameArtists = (a: Ref[], b: Ref[]) =>
  a.length === b.length && a.every((x) => b.some((y) => y.id === x.id));

function fill(t: TrackDetail) {
  track.value = t;
  Object.assign(form, {
    title: t.title,
    artists: t.artists.map((a) => a.name),
    album: t.album?.title ?? "",
    // Left empty while the album artists are the track's own.
    albumArtists:
      t.album && !sameArtists(t.album.artists, t.artists)
        ? t.album.artists.map((a) => a.name)
        : [],
    trackNo: t.trackNo,
    discNo: t.discNo,
    year: t.year,
    neteaseId: t.neteaseId,
  });
  lyrics.value = t.lyrics;
}

watch(
  () => props.id,
  async (id) => {
    // Keep the last track on close so the drawer doesn't blank while sliding out.
    if (id == null) return;
    track.value = null;
    loadError.value = "";
    try {
      fill(await request<TrackDetail>("GET", `/tracks/${id}`));
    } catch (err) {
      loadError.value = (err as ApiError).message;
    }
  },
  { immediate: true },
);

async function save() {
  if (!track.value) return;
  saving.value = true;
  try {
    fill(await request<TrackDetail>("PUT", `/tracks/${track.value.id}`, form));
    message.success("已保存");
    emit("changed");
  } catch (err) {
    message.error((err as ApiError).message);
  } finally {
    saving.value = false;
  }
}

async function replaceLyrics(body: Blob | string, method = "PUT") {
  if (!track.value) return;
  savingLyrics.value = true;
  try {
    fill(
      await request<TrackDetail>(
        method,
        `/tracks/${track.value.id}/lyrics`,
        body,
      ),
    );
    message.success(method === "DELETE" ? "已删除歌词" : "已保存歌词");
    emit("changed");
  } catch (err) {
    message.error((err as ApiError).message);
  } finally {
    savingLyrics.value = false;
  }
}

function lyricsPicked(e: Event) {
  const file = (e.target as HTMLInputElement).files?.[0];
  (e.target as HTMLInputElement).value = "";
  // Sent as raw bytes so the service can detect GB18030 and UTF-16.
  if (file) replaceLyrics(file);
}

async function setCover(file: File) {
  fill(
    await request<TrackDetail>("PUT", `/tracks/${track.value!.id}/cover`, file),
  );
  emit("changed");
}
async function clearCover() {
  fill(
    await request<TrackDetail>("DELETE", `/tracks/${track.value!.id}/cover`),
  );
  emit("changed");
}

function confirmDelete() {
  const t = track.value;
  if (!t) return;
  dialog.warning({
    title: "删除歌曲",
    content: `“${t.title}”及其音频文件将被永久删除。`,
    positiveText: "删除",
    negativeText: "取消",
    onPositiveClick: async () => {
      try {
        await request("DELETE", `/tracks/${t.id}`);
        message.success("已删除");
        emit("changed");
        emit("close");
      } catch (err) {
        message.error((err as ApiError).message);
        return false;
      }
    },
  });
}
</script>

<template>
  <n-drawer
    :show="id != null"
    :width="560"
    @update:show="(v: boolean) => !v && emit('close')"
  >
    <n-drawer-content :title="track?.title ?? '歌曲'" closable>
      <div v-if="loadError">{{ loadError }}</div>
      <div v-else-if="!track" style="text-align: center; padding: 48px">
        <n-spin />
      </div>
      <template v-else>
        <audio
          :src="audioURL(track.id)"
          controls
          preload="none"
          style="width: 100%"
        />
        <n-divider title-placement="left">元数据</n-divider>
        <n-form :model="form" label-placement="top" @submit.prevent="save">
          <n-form-item label="标题">
            <n-input v-model:value="form.title" />
          </n-form-item>
          <n-form-item label="艺术家">
            <artists-input v-model="form.artists" />
          </n-form-item>
          <n-form-item label="专辑艺术家">
            <artists-input v-model="form.albumArtists" empty-hint="同艺术家" />
          </n-form-item>
          <n-form-item label="专辑">
            <n-input v-model:value="form.album" clearable />
          </n-form-item>
          <n-grid :cols="3" :x-gap="12">
            <n-gi>
              <n-form-item label="音轨号">
                <n-input-number
                  v-model:value="form.trackNo"
                  :min="0"
                  :max="9999"
                />
              </n-form-item>
            </n-gi>
            <n-gi>
              <n-form-item label="碟号">
                <n-input-number
                  v-model:value="form.discNo"
                  :min="0"
                  :max="999"
                />
              </n-form-item>
            </n-gi>
            <n-gi>
              <n-form-item label="年份">
                <n-input-number
                  v-model:value="form.year"
                  :min="0"
                  :max="9999"
                  :show-button="false"
                />
              </n-form-item>
            </n-gi>
          </n-grid>
          <n-form-item
            label="网易云 ID"
            feedback="导入网易云歌曲时据此判断它是否已在曲库中"
          >
            <netease-id-input v-model="form.neteaseId" />
          </n-form-item>
          <n-button type="primary" attr-type="submit" :loading="saving"
            >保存</n-button
          >
        </n-form>

        <n-divider title-placement="left">封面</n-divider>
        <cover-editor
          :cover="track.cover"
          :fallback="track.album?.cover"
          fallback-label="使用专辑封面"
          :upload="setCover"
          :remove="clearCover"
        />

        <n-divider title-placement="left">
          歌词<template v-if="track.lyricsFormat">
            · {{ track.lyricsFormat.toUpperCase() }}</template
          >
        </n-divider>
        <n-input
          v-model:value="lyrics"
          type="textarea"
          :autosize="{ minRows: 4, maxRows: 14 }"
          placeholder="无歌词"
        />
        <n-space :size="8" style="margin-top: 12px">
          <n-button
            size="small"
            :loading="savingLyrics"
            :disabled="lyrics === track.lyrics"
            @click="replaceLyrics(lyrics)"
          >
            保存歌词
          </n-button>
          <n-button
            size="small"
            :disabled="savingLyrics"
            @click="lyricsInput?.click()"
          >
            上传 .lrc / .ttml
          </n-button>
          <n-button
            v-if="track.hasLyrics"
            size="small"
            :disabled="savingLyrics"
            @click="replaceLyrics('', 'DELETE')"
          >
            删除歌词
          </n-button>
        </n-space>
        <input
          ref="lyricsInput"
          type="file"
          accept=".lrc,.ttml,.txt"
          hidden
          @change="lyricsPicked"
        />

        <n-divider title-placement="left">文件</n-divider>
        <n-descriptions :column="2" size="small" label-placement="left">
          <n-descriptions-item label="格式">{{
            track.format.toUpperCase()
          }}</n-descriptions-item>
          <n-descriptions-item label="时长">{{
            formatDuration(track.durationMs)
          }}</n-descriptions-item>
          <n-descriptions-item label="大小">{{
            formatBytes(track.size)
          }}</n-descriptions-item>
          <n-descriptions-item label="入库">{{
            new Date(track.createdAt).toLocaleString()
          }}</n-descriptions-item>
          <n-descriptions-item label="上传文件名" :span="2">{{
            track.originalName
          }}</n-descriptions-item>
        </n-descriptions>
      </template>
      <template v-if="track" #footer>
        <n-button type="error" ghost @click="confirmDelete">删除歌曲</n-button>
      </template>
    </n-drawer-content>
  </n-drawer>
</template>
