<script setup lang="ts">
import { onMounted, ref } from "vue";
import { RouterLink } from "vue-router";
import {
  NAlert,
  NButton,
  NCard,
  NGi,
  NGrid,
  NSkeleton,
  NStatistic,
} from "naive-ui";
import {
  formatBytes,
  formatDuration,
  request,
  type ApiError,
  type Overview,
} from "../api";

const data = ref<Overview | null>(null);
const error = ref("");

async function load() {
  error.value = "";
  try {
    data.value = await request<Overview>("GET", "/overview");
  } catch (err) {
    error.value = (err as ApiError).message;
  }
}
onMounted(load);
</script>

<template>
  <n-alert v-if="error" type="error" :show-icon="false">
    {{ error }}
    <n-button text type="primary" style="margin-left: 8px" @click="load"
      >重试</n-button
    >
  </n-alert>
  <n-grid v-else :cols="3" :x-gap="16" :y-gap="16">
    <template v-if="data">
      <n-gi>
        <n-card><n-statistic label="歌曲" :value="data.tracks" /></n-card>
      </n-gi>
      <n-gi>
        <n-card><n-statistic label="专辑" :value="data.albums" /></n-card>
      </n-gi>
      <n-gi>
        <n-card><n-statistic label="艺术家" :value="data.artists" /></n-card>
      </n-gi>
      <n-gi>
        <n-card
          ><n-statistic label="总时长" :value="formatDuration(data.durationMs)"
        /></n-card>
      </n-gi>
      <n-gi>
        <n-card>
          <n-statistic
            label="存储占用"
            :value="formatBytes(data.mediaBytes + data.coverBytes)"
          />
        </n-card>
      </n-gi>
      <n-gi>
        <n-card>
          <n-statistic label="缺少标签">
            <router-link v-if="data.missingTags" to="/tracks?missing=true">
              {{ data.missingTags }}
            </router-link>
            <template v-else>0</template>
          </n-statistic>
        </n-card>
      </n-gi>
    </template>
    <n-gi v-for="i in 6" v-else :key="i">
      <n-card><n-skeleton text :repeat="2" /></n-card>
    </n-gi>
  </n-grid>
</template>
