<script setup lang="ts">
import { computed } from "vue";
import { NButton, NDynamicTags, NText } from "naive-ui";

// Edits an ordered list of artist names. Tags often pack several artists
// into one value ("周杰伦/费玉清"); splitting is offered, never automatic,
// since names like "AC/DC" contain separators too.
const model = defineModel<string[]>({ required: true });
defineProps<{ emptyHint?: string }>();

const SEPARATORS = /\s*(?:[/、;；,，&＆]|\s(?:feat\.?|ft\.?)\s)\s*/i;

const splittable = computed(() => model.value.some((n) => SEPARATORS.test(n)));

function split() {
  const out: string[] = [];
  for (const name of model.value)
    for (const part of name.split(SEPARATORS)) {
      const p = part.trim();
      if (p && !out.some((o) => o.toLowerCase() === p.toLowerCase()))
        out.push(p);
    }
  model.value = out;
}
</script>

<template>
  <div class="artists-input">
    <n-dynamic-tags v-model:value="model" />
    <n-text v-if="!model.length && emptyHint" depth="3">{{ emptyHint }}</n-text>
    <n-button v-if="splittable" text type="primary" size="tiny" @click="split">
      按分隔符拆分
    </n-button>
  </div>
</template>

<style scoped>
.artists-input {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px;
}
</style>
