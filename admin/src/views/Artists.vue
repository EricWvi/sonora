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
  NModal,
  NSpace,
  useMessage,
  type DataTableColumns,
  type DataTableRowKey,
} from "naive-ui";
import { query, request, type ApiError, type Artist, type Page } from "../api";
import CoverEditor from "../components/CoverEditor.vue";
import { idColumn, neteaseIdColumn, newestFirst } from "../columns";
import CoverImage from "../components/CoverImage.vue";
import MergeModal from "../components/MergeModal.vue";
import NeteaseIdInput from "../components/NeteaseIdInput.vue";

const PAGE_SIZE = 50;
const route = useRoute();
const router = useRouter();
const message = useMessage();

const filters = computed(() => ({
  q: (route.query.q as string) ?? "",
  page: Number(route.query.page) || 1,
}));
const search = ref(filters.value.q);
const data = ref<Page<Artist>>({ items: [], total: 0 });
const loading = ref(false);
const error = ref("");
const checked = ref<DataTableRowKey[]>([]);
const selected = reactive(new Map<number, Artist>());

function setQuery(patch: Record<string, string | undefined>) {
  router.replace({ query: { ...route.query, ...patch } });
}

async function load() {
  const f = filters.value;
  loading.value = true;
  error.value = "";
  try {
    data.value = await request<Page<Artist>>(
      "GET",
      "/artists" +
        query({
          q: f.q,
          ...newestFirst(f.q),
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

const showEdit = ref(false);
const editing = ref<Artist | null>(null);
const name = ref("");
const description = ref("");
const neteaseId = ref<number | null>(null);
const saving = ref(false);

function edit(a: Artist) {
  editing.value = a;
  name.value = a.name;
  description.value = a.description;
  neteaseId.value = a.neteaseId;
  showEdit.value = true;
}

async function save() {
  if (!editing.value) return;
  saving.value = true;
  try {
    editing.value = await request<Artist>(
      "PUT",
      `/artists/${editing.value.id}`,
      {
        name: name.value,
        description: description.value,
        neteaseId: neteaseId.value,
      },
    );
    message.success("已保存");
    load();
  } catch (err) {
    message.error((err as ApiError).message);
  } finally {
    saving.value = false;
  }
}

async function setImage(file: File) {
  editing.value = await request<Artist>(
    "PUT",
    `/artists/${editing.value!.id}/image`,
    file,
  );
  load();
}
async function clearImage() {
  editing.value = await request<Artist>(
    "DELETE",
    `/artists/${editing.value!.id}/image`,
  );
  load();
}

const merging = ref(false);
async function merge(target: number, sources: number[]) {
  try {
    await request<Artist>("POST", `/artists/${target}/merge`, {
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

const columns: DataTableColumns<Artist> = [
  { type: "selection" },
  idColumn(),
  {
    title: "",
    key: "image",
    width: 56,
    render: (a) =>
      h(
        "div",
        { class: "thumb" },
        a.image ? [h(CoverImage, { cover: a.image, size: 96 })] : [],
      ),
  },
  {
    title: "艺术家",
    key: "name",
    ellipsis: { tooltip: true },
    render: (a) => h("a", { class: "link", onClick: () => edit(a) }, a.name),
  },
  {
    title: "歌曲",
    key: "trackCount",
    width: 100,
    render: (a) =>
      h(RouterLink, { to: `/tracks?artistId=${a.id}` }, () => a.trackCount),
  },
  {
    title: "专辑",
    key: "albumCount",
    width: 100,
    render: (a) =>
      h(RouterLink, { to: `/albums?artistId=${a.id}` }, () => a.albumCount),
  },
  neteaseIdColumn("artist"),
];

const pagination = computed(() => ({
  page: filters.value.page,
  pageSize: PAGE_SIZE,
  itemCount: data.value.total,
  prefix: () => `共 ${data.value.total} 位`,
}));
</script>

<template>
  <n-space vertical :size="16">
    <n-space align="center" justify="space-between">
      <n-input
        :value="search"
        placeholder="搜索艺术家"
        clearable
        style="width: 280px"
        @update:value="onSearch"
      />
      <n-space v-if="selected.size" align="center">
        <span>已选 {{ selected.size }} 位</span>
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
      :row-key="(a: Artist) => a.id"
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
    title="编辑艺术家"
    style="width: 480px"
    @update:show="(v: boolean) => (showEdit = v)"
    @after-leave="editing = null"
  >
    <template v-if="editing">
      <n-form @submit.prevent="save">
        <n-form-item label="名称">
          <n-input v-model:value="name" autofocus />
        </n-form-item>
        <n-form-item label="网易云 ID">
          <netease-id-input v-model="neteaseId" />
        </n-form-item>
        <n-form-item label="简介">
          <n-input
            v-model:value="description"
            type="textarea"
            :autosize="{ minRows: 3, maxRows: 10 }"
            :maxlength="10000"
            show-count
          />
        </n-form-item>
        <n-button type="primary" attr-type="submit" :loading="saving"
          >保存</n-button
        >
      </n-form>
      <n-divider title-placement="left">图片</n-divider>
      <cover-editor
        :cover="editing.image"
        empty-text="无图片"
        :upload="setImage"
        :remove="clearImage"
      />
    </template>
  </n-modal>

  <merge-modal
    :show="merging"
    title="合并艺术家"
    :items="[...selected.values()].map((a) => ({ id: a.id, label: a.name }))"
    :merge="merge"
    @close="merging = false"
  />
</template>

<style scoped>
:deep(.thumb) {
  width: 36px;
  height: 36px;
  border-radius: 50%;
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
