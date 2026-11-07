<!-- 私人 FM 的分组歌单（Sonora 新增，SPlayer 的私人 FM 没有这项设置）。一张专辑的歌分属
多个歌单时，归入歌曲数最多的歌单；数量相同时归入排在前面的，所以选择的先后有意义。 -->
<template>
  <n-flex vertical :size="12">
    <n-text depth="3">
      不足 10
      首的专辑按所在歌单分组，同一段里同组的专辑连在一起；不在这些歌单里的不进入私人FM。
      一张专辑分属多个歌单时，归入歌曲最多的那个，数量相同时归入先选的。
    </n-text>
    <n-select
      v-model:value="selected"
      :options="options"
      :loading="loading"
      :disabled="error"
      placeholder="选择歌单"
      multiple
      @update:value="save"
    />
  </n-flex>
</template>

<script setup lang="ts">
import type { SelectOption } from "naive-ui";
import { source } from "@/api/source";
import { fmRevision } from "@/core/player/fm";

const selected = ref<number[]>([]);
const options = shallowRef<SelectOption[]>([]);
const loading = ref<boolean>(true);
const error = ref<boolean>(false);

const save = async (ids: number[]) => {
  try {
    const state = await source.fmSetGroups(ids);
    selected.value = state.groupIds;
    fmRevision.value++;
  } catch (e) {
    console.error("保存分组歌单失败：", e);
    window.$message.error("保存失败，请重试");
  }
};

onMounted(async () => {
  try {
    const [playlists, state] = await Promise.all([
      source.playlists(),
      source.fm(),
    ]);
    options.value = playlists.items.map((p) => ({
      label: p.name,
      value: p.id,
    }));
    selected.value = state.groupIds;
  } catch (e) {
    console.error("读取分组歌单失败：", e);
    error.value = true;
    window.$message.error("读取失败，请重试");
  } finally {
    loading.value = false;
  }
});
</script>
