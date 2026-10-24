<script setup lang="ts">
import { computed, h, ref, watch } from "vue";
import { RouterLink, useRoute, useRouter } from "vue-router";
import {
  NAlert,
  NButton,
  NDataTable,
  NInput,
  NSpace,
  NSwitch,
  NTag,
  NText,
  type DataTableColumns,
} from "naive-ui";
import {
  coverURL,
  formatDuration,
  query,
  request,
  type ApiError,
  type Page,
  type Track,
} from "../api";
import TrackDrawer from "../components/TrackDrawer.vue";

const PAGE_SIZE = 50;
const route = useRoute();
const router = useRouter();

// The URL holds the filters so links like "缺少标签" land here filtered.
const filters = computed(() => ({
  q: (route.query.q as string) ?? "",
  missing: route.query.missing === "true",
  albumId: Number(route.query.albumId) || undefined,
  artistId: Number(route.query.artistId) || undefined,
  page: Number(route.query.page) || 1,
}));
const editing = computed(() => Number(route.query.edit) || null);

const search = ref(filters.value.q);
const data = ref<Page<Track>>({ items: [], total: 0 });
const loading = ref(false);
const error = ref("");
const scopeLabel = ref("");

function setQuery(patch: Record<string, string | undefined>) {
  router.replace({ query: { ...route.query, ...patch } });
}

async function load() {
  const f = filters.value;
  loading.value = true;
  error.value = "";
  try {
    data.value = await request<Page<Track>>(
      "GET",
      "/tracks" +
        query({
          q: f.q,
          missing: f.missing,
          albumId: f.albumId,
          artistId: f.artistId,
          offset: (f.page - 1) * PAGE_SIZE,
          limit: PAGE_SIZE,
        }),
    );
    if (f.albumId) {
      const a = data.value.items[0]?.album;
      scopeLabel.value = a ? `专辑：${a.title}` : "专辑";
    } else if (f.artistId) {
      const t = data.value.items[0];
      const name =
        t?.artist?.id === f.artistId
          ? t.artist.name
          : (t?.album?.artist?.name ?? "");
      scopeLabel.value = `艺术家：${name}`;
    }
  } catch (err) {
    error.value = (err as ApiError).message;
  } finally {
    loading.value = false;
  }
}

watch(filters, load, { immediate: true });
watch(
  () => filters.value.q,
  (q) => (search.value = q),
);

function toggleMissing(v: boolean) {
  setQuery({ missing: v ? "true" : undefined, page: undefined });
}

let searchTimer = 0;
function onSearch(v: string) {
  search.value = v;
  clearTimeout(searchTimer);
  searchTimer = window.setTimeout(
    () => setQuery({ q: v || undefined, page: undefined }),
    300,
  );
}

const columns: DataTableColumns<Track> = [
  {
    title: "",
    key: "cover",
    width: 56,
    render: (t) => {
      const c = t.cover ?? t.album?.cover;
      return c
        ? h("img", { src: coverURL(c), class: "thumb", alt: "" })
        : h("div", { class: "thumb" });
    },
  },
  {
    title: "标题",
    key: "title",
    ellipsis: { tooltip: true },
    render: (t) =>
      h(
        "a",
        { class: "link", onClick: () => setQuery({ edit: String(t.id) }) },
        t.title,
      ),
  },
  {
    title: "艺术家",
    key: "artist",
    ellipsis: { tooltip: true },
    render: (t) =>
      t.artist
        ? h(
            RouterLink,
            { to: `/tracks?artistId=${t.artist.id}` },
            () => t.artist!.name,
          )
        : h(
            NTag,
            { size: "small", type: "warning", bordered: false },
            () => "缺失",
          ),
  },
  {
    title: "专辑",
    key: "album",
    ellipsis: { tooltip: true },
    render: (t) =>
      t.album
        ? h(
            RouterLink,
            { to: `/tracks?albumId=${t.album.id}` },
            () => t.album!.title,
          )
        : h(
            NTag,
            { size: "small", type: "warning", bordered: false },
            () => "缺失",
          ),
  },
  {
    title: "音轨",
    key: "trackNo",
    width: 70,
    render: (t) =>
      t.trackNo ? (t.discNo > 1 ? `${t.discNo}-${t.trackNo}` : t.trackNo) : "",
  },
  {
    title: "时长",
    key: "durationMs",
    width: 80,
    render: (t) => formatDuration(t.durationMs),
  },
  {
    title: "格式",
    key: "format",
    width: 70,
    render: (t) => t.format.toUpperCase(),
  },
  {
    title: "歌词",
    key: "hasLyrics",
    width: 60,
    render: (t) => (t.hasLyrics ? "有" : h(NText, { depth: 3 }, () => "—")),
  },
];

const pagination = computed(() => ({
  page: filters.value.page,
  pageSize: PAGE_SIZE,
  itemCount: data.value.total,
  showQuickJumper: data.value.total > PAGE_SIZE * 5,
  prefix: () => `共 ${data.value.total} 首`,
}));
</script>

<template>
  <n-space vertical :size="16">
    <n-space align="center" justify="space-between">
      <n-space align="center">
        <n-input
          :value="search"
          placeholder="搜索标题、艺术家、专辑"
          clearable
          style="width: 280px"
          @update:value="onSearch"
        />
        <span class="toggle">
          <n-switch :value="filters.missing" @update:value="toggleMissing" />
          <n-text @click="toggleMissing(!filters.missing)">只看缺少标签</n-text>
        </span>
        <n-tag
          v-if="filters.albumId || filters.artistId"
          closable
          @close="
            setQuery({
              albumId: undefined,
              artistId: undefined,
              page: undefined,
            })
          "
        >
          {{ scopeLabel }}
        </n-tag>
      </n-space>
      <router-link to="/upload"
        ><n-button type="primary">上传</n-button></router-link
      >
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
      :row-key="(t: Track) => t.id"
      size="small"
      @update:page="
        (p: number) => setQuery({ page: p > 1 ? String(p) : undefined })
      "
    />
  </n-space>
  <track-drawer
    :id="editing"
    @close="setQuery({ edit: undefined })"
    @changed="load"
  />
</template>

<style scoped>
.toggle {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
}
:deep(.thumb) {
  width: 36px;
  height: 36px;
  border-radius: 3px;
  object-fit: cover;
  display: block;
  background: rgba(128, 128, 128, 0.12);
}
:deep(.link) {
  color: inherit !important;
}
:deep(.link:hover) {
  text-decoration: underline;
}
</style>
