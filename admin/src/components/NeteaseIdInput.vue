<script setup lang="ts">
import { NInputNumber } from "naive-ui";

// A NetEase ID: a positive integer, or null when the item has none. Clearing
// the box clears the ID.
const model = defineModel<number | null>({ required: true });
</script>

<template>
  <n-input-number
    v-model:value="model"
    :min="1"
    :precision="0"
    :show-button="false"
    clearable
    placeholder="未关联网易云"
  />
</template>
