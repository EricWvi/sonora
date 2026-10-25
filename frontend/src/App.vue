<template>
  <ThemeProvider>
    <SignedOut v-if="userStore.signedOut" />
    <AppLayout v-else />
  </ThemeProvider>
</template>

<script setup lang="ts">
import ThemeProvider from "@/components/Global/ThemeProvider.vue";
import AppLayout from "@/layout/AppLayout.vue";
import SignedOut from "@/views/Status/SignedOut.vue";
import { useUserStore } from "@/stores/user";

const userStore = useUserStore();
</script>
