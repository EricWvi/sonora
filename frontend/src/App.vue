<template>
  <Provider>
    <SignedOut v-if="userStore.signedOut" />
    <AppLayout v-else />
  </Provider>
</template>

<script setup lang="ts">
import Provider from "@/components/Global/Provider.vue";
import AppLayout from "@/layout/AppLayout.vue";
import SignedOut from "@/views/Status/SignedOut.vue";
import { useUserStore } from "@/stores/user";

const userStore = useUserStore();
</script>
