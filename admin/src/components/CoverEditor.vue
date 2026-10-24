<script setup lang="ts">
import { ref } from "vue";
import { NButton, NSpace, NText, useMessage } from "naive-ui";
import type { ApiError, Cover } from "../api";
import CoverImage from "./CoverImage.vue";

// Shows a cover (or an artist image) and replaces or removes it. fallback
// is shown, dimmed, when there is no cover of its own (a track falls back
// to its album's).
const props = withDefaults(
  defineProps<{
    cover: Cover | null;
    fallback?: Cover | null;
    fallbackLabel?: string;
    emptyText?: string;
    upload: (file: File) => Promise<unknown>;
    remove: () => Promise<unknown>;
  }>(),
  { fallback: null, fallbackLabel: "", emptyText: "无封面" },
);

const message = useMessage();
const input = ref<HTMLInputElement | null>(null);
const busy = ref(false);

async function run(fn: () => Promise<unknown>) {
  busy.value = true;
  try {
    await fn();
  } catch (err) {
    message.error((err as ApiError).message);
  } finally {
    busy.value = false;
  }
}

function picked(e: Event) {
  const file = (e.target as HTMLInputElement).files?.[0];
  (e.target as HTMLInputElement).value = "";
  if (file) run(() => props.upload(file));
}
</script>

<template>
  <n-space align="center" :size="16">
    <div class="cover">
      <cover-image v-if="cover" :cover="cover" :size="300" />
      <cover-image
        v-else-if="fallback"
        :cover="fallback"
        :size="300"
        class="fallback"
      />
      <n-text v-else depth="3">{{ emptyText }}</n-text>
    </div>
    <n-space vertical :size="8">
      <n-text v-if="!cover && fallback && fallbackLabel" depth="3">{{
        fallbackLabel
      }}</n-text>
      <n-space :size="8">
        <n-button size="small" :loading="busy" @click="input?.click()">
          {{ cover ? "替换" : "上传" }}
        </n-button>
        <n-button
          v-if="cover"
          size="small"
          :disabled="busy"
          @click="run(remove)"
          >删除</n-button
        >
      </n-space>
    </n-space>
    <input
      ref="input"
      type="file"
      accept="image/jpeg,image/png,image/webp"
      hidden
      @change="picked"
    />
  </n-space>
</template>

<style scoped>
.cover {
  width: 120px;
  height: 120px;
  border-radius: 4px;
  overflow: hidden;
  display: grid;
  place-items: center;
  background: rgba(128, 128, 128, 0.12);
}
.cover .fallback {
  opacity: 0.6;
}
</style>
