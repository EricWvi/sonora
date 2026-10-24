<script setup lang="ts">
import { computed, h, reactive, ref, watch } from "vue";
import { RouterLink, useRoute, useRouter } from "vue-router";
import {
  NAlert,
  NButton,
  NDataTable,
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
import MergeModal from "../components/MergeModal.vue";

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
        query({ q: f.q, offset: (f.page - 1) * PAGE_SIZE, limit: PAGE_SIZE }),
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

const renaming = ref<Artist | null>(null);
const name = ref("");
const saving = ref(false);

async function rename() {
  if (!renaming.value) return;
  saving.value = true;
  try {
    await request<Artist>("PUT", `/artists/${renaming.value.id}`, {
      name: name.value,
    });
    message.success("已保存");
    renaming.value = null;
    load();
  } catch (err) {
    message.error((err as ApiError).message);
  } finally {
    saving.value = false;
  }
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
  {
    title: "艺术家",
    key: "name",
    ellipsis: { tooltip: true },
    render: (a) =>
      h(
        "a",
        {
          class: "link",
          onClick: () => {
            renaming.value = a;
            name.value = a.name;
          },
        },
        a.name,
      ),
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
    :show="renaming != null"
    preset="card"
    title="重命名艺术家"
    style="width: 400px"
    @update:show="(v: boolean) => !v && (renaming = null)"
  >
    <n-form @submit.prevent="rename">
      <n-form-item label="名称">
        <n-input v-model:value="name" autofocus />
      </n-form-item>
      <n-button type="primary" attr-type="submit" :loading="saving"
        >保存</n-button
      >
    </n-form>
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
:deep(.link) {
  color: inherit !important;
}
:deep(.link:hover) {
  text-decoration: underline;
}
</style>
