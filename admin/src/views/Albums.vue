<script setup lang="ts">
import { computed, h, reactive, ref, watch } from "vue";
import { RouterLink, useRoute, useRouter } from "vue-router";
import {
  NAlert,
  NButton,
  NDataTable,
  NDivider,
  NForm,
  NFormItem,
  NInput,
  NInputNumber,
  NModal,
  NSpace,
  NTag,
  useDialog,
  useMessage,
  type DataTableColumns,
  type DataTableRowKey,
} from "naive-ui";
import {
  artistNames,
  formatDuration,
  query,
  request,
  type Album,
  type ApiError,
  type Page,
} from "../api";
import ArtistLinks from "../components/ArtistLinks.vue";
import CoverImage from "../components/CoverImage.vue";
import ArtistsInput from "../components/ArtistsInput.vue";
import CoverEditor from "../components/CoverEditor.vue";
import MergeModal from "../components/MergeModal.vue";

const PAGE_SIZE = 50;
const route = useRoute();
const router = useRouter();
const message = useMessage();
const dialog = useDialog();

const filters = computed(() => ({
  q: (route.query.q as string) ?? "",
  artistId: Number(route.query.artistId) || undefined,
  page: Number(route.query.page) || 1,
}));
const search = ref(filters.value.q);
const data = ref<Page<Album>>({ items: [], total: 0 });
const loading = ref(false);
const error = ref("");
const checked = ref<DataTableRowKey[]>([]);
// Selected albums are remembered across pages, so the labels are kept here.
const selected = reactive(new Map<number, Album>());

function setQuery(patch: Record<string, string | undefined>) {
  router.replace({ query: { ...route.query, ...patch } });
}

async function load() {
  const f = filters.value;
  loading.value = true;
  error.value = "";
  try {
    data.value = await request<Page<Album>>(
      "GET",
      "/albums" +
        query({
          q: f.q,
          artistId: f.artistId,
          offset: (f.page - 1) * PAGE_SIZE,
          limit: PAGE_SIZE,
        }),
    );
  } catch (err) {
    error.value = (err as ApiError).message;
  } finally {
    loading.value = false;
  }
}
watch(filters, load, { immediate: true });

let searchTimer = 0;
function onSearch(v: string) {
  search.value = v;
  clearTimeout(searchTimer);
  searchTimer = window.setTimeout(
    () => setQuery({ q: v || undefined, page: undefined }),
    300,
  );
}

function onCheck(keys: DataTableRowKey[]) {
  checked.value = keys;
  for (const a of data.value.items) {
    if (keys.includes(a.id)) selected.set(a.id, a);
    else selected.delete(a.id);
  }
  for (const id of [...selected.keys()])
    if (!keys.includes(id)) selected.delete(id);
}
function clearSelection() {
  checked.value = [];
  selected.clear();
}

const label = (a: Album) =>
  a.artists.length ? `${a.title} — ${artistNames(a.artists)}` : a.title;

// --- editing ----------------------------------------------------------------

const editing = ref<Album | null>(null);
// Kept separate from `editing` so the content survives the leave transition.
const showEdit = ref(false);
const form = reactive({ title: "", artists: [] as string[], year: 0 });
const saving = ref(false);

function edit(a: Album) {
  editing.value = a;
  showEdit.value = true;
  Object.assign(form, {
    title: a.title,
    artists: a.artists.map((r) => r.name),
    year: a.year,
  });
}

async function save() {
  if (!editing.value) return;
  saving.value = true;
  try {
    editing.value = await request<Album>(
      "PUT",
      `/albums/${editing.value.id}`,
      form,
    );
    message.success("已保存");
    load();
  } catch (err) {
    message.error((err as ApiError).message);
  } finally {
    saving.value = false;
  }
}

async function setCover(file: File) {
  editing.value = await request<Album>(
    "PUT",
    `/albums/${editing.value!.id}/cover`,
    file,
  );
  load();
}
async function clearCover() {
  editing.value = await request<Album>(
    "DELETE",
    `/albums/${editing.value!.id}/cover`,
  );
  load();
}

function confirmDelete(a: Album) {
  dialog.warning({
    title: "删除专辑",
    content: `“${a.title}”的 ${a.trackCount} 首歌曲及音频文件将被永久删除。`,
    positiveText: "删除",
    negativeText: "取消",
    onPositiveClick: async () => {
      try {
        await request("DELETE", `/albums/${a.id}`);
        message.success("已删除");
        showEdit.value = false;
        selected.delete(a.id);
        checked.value = checked.value.filter((k) => k !== a.id);
        load();
      } catch (err) {
        message.error((err as ApiError).message);
        return false;
      }
    },
  });
}

// --- merging ----------------------------------------------------------------

const merging = ref(false);
async function merge(target: number, sources: number[]) {
  try {
    await request<Album>("POST", `/albums/${target}/merge`, {
      sourceIds: sources,
    });
    message.success("已合并");
    clearSelection();
    load();
  } catch (err) {
    message.error((err as ApiError).message);
    throw err;
  }
}

const columns: DataTableColumns<Album> = [
  { type: "selection" },
  {
    title: "",
    key: "cover",
    width: 56,
    render: (a) =>
      h(
        "div",
        { class: "thumb" },
        a.cover ? [h(CoverImage, { cover: a.cover, size: 96 })] : [],
      ),
  },
  {
    title: "专辑",
    key: "title",
    ellipsis: { tooltip: true },
    render: (a) => h("a", { class: "link", onClick: () => edit(a) }, a.title),
  },
  {
    title: "专辑艺术家",
    key: "artist",
    ellipsis: { tooltip: true },
    render: (a) =>
      a.artists.length
        ? h(ArtistLinks, {
            artists: a.artists,
            to: (id: number) => `/albums?artistId=${id}`,
          })
        : h(NTag, { size: "small", bordered: false }, () => "未知"),
  },
  { title: "年份", key: "year", width: 80, render: (a) => a.year || "" },
  {
    title: "歌曲",
    key: "trackCount",
    width: 80,
    render: (a) =>
      h(RouterLink, { to: `/tracks?albumId=${a.id}` }, () => a.trackCount),
  },
  {
    title: "时长",
    key: "durationMs",
    width: 90,
    render: (a) => formatDuration(a.durationMs),
  },
];

const pagination = computed(() => ({
  page: filters.value.page,
  pageSize: PAGE_SIZE,
  itemCount: data.value.total,
  prefix: () => `共 ${data.value.total} 张`,
}));
</script>

<template>
  <n-space vertical :size="16">
    <n-space align="center" justify="space-between">
      <n-space align="center">
        <n-input
          :value="search"
          placeholder="搜索专辑或艺术家"
          clearable
          style="width: 280px"
          @update:value="onSearch"
        />
        <n-tag
          v-if="filters.artistId"
          closable
          @close="setQuery({ artistId: undefined })"
        >
          按艺术家筛选
        </n-tag>
      </n-space>
      <n-space v-if="selected.size" align="center">
        <span>已选 {{ selected.size }} 张</span>
        <n-button size="small" @click="clearSelection">取消选择</n-button>
        <n-button
          size="small"
          type="primary"
          :disabled="selected.size < 2"
          @click="merging = true"
        >
          合并
        </n-button>
      </n-space>
    </n-space>
    <n-alert v-if="error" type="error" :show-icon="false">
      {{ error }}
      <n-button text type="primary" style="margin-left: 8px" @click="load"
        >重试</n-button
      >
    </n-alert>
    <n-data-table
      remote
      :columns="columns"
      :data="data.items"
      :loading="loading"
      :pagination="pagination"
      :row-key="(a: Album) => a.id"
      :checked-row-keys="checked"
      size="small"
      @update:checked-row-keys="onCheck"
      @update:page="
        (p: number) => setQuery({ page: p > 1 ? String(p) : undefined })
      "
    />
  </n-space>

  <n-modal
    :show="showEdit"
    preset="card"
    title="编辑专辑"
    style="width: 480px"
    @update:show="(v: boolean) => (showEdit = v)"
    @after-leave="editing = null"
  >
    <template v-if="editing">
      <n-form :model="form" @submit.prevent="save">
        <n-form-item label="专辑名">
          <n-input v-model:value="form.title" />
        </n-form-item>
        <n-form-item label="专辑艺术家">
          <artists-input v-model="form.artists" />
        </n-form-item>
        <n-form-item label="年份">
          <n-input-number
            v-model:value="form.year"
            :min="0"
            :max="9999"
            :show-button="false"
          />
        </n-form-item>
        <n-button type="primary" attr-type="submit" :loading="saving"
          >保存</n-button
        >
      </n-form>
      <n-divider title-placement="left">封面</n-divider>
      <cover-editor
        :cover="editing.cover"
        :upload="setCover"
        :remove="clearCover"
      />
    </template>
    <template v-if="editing" #footer>
      <n-button type="error" ghost @click="confirmDelete(editing)"
        >删除专辑</n-button
      >
    </template>
  </n-modal>

  <merge-modal
    :show="merging"
    title="合并专辑"
    :items="[...selected.values()].map((a) => ({ id: a.id, label: label(a) }))"
    :merge="merge"
    @close="merging = false"
  />
</template>

<style scoped>
:deep(.thumb) {
  width: 36px;
  height: 36px;
  border-radius: 3px;
  overflow: hidden;
  background: rgba(128, 128, 128, 0.12);
}
:deep(.link) {
  color: inherit !important;
}
:deep(.link:hover) {
  text-decoration: underline;
}
</style>
