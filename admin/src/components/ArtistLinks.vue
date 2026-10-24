<script setup lang="ts">
import { RouterLink } from "vue-router";
import type { Ref } from "../api";

// Links each of several artists separately.
defineProps<{ artists: Ref[]; to: (id: number) => string }>();
</script>

<template>
  <span>
    <template v-for="(a, i) in artists" :key="a.id">
      <template v-if="i"> / </template>
      <router-link :to="to(a.id)">{{ a.name }}</router-link>
    </template>
  </span>
</template>
