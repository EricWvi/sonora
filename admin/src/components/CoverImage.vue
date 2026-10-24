<script setup lang="ts">
import { computed, ref, watch } from "vue";
import { thumbHashToDataURL } from "thumbhash";
import { coverURL, type Cover, type CoverSize } from "../api";

// Fills its box with a cover: the blurred placeholder at once, then the
// image fading in once it has loaded.
const props = defineProps<{ cover: Cover; size?: CoverSize }>();

const placeholders = new Map<string, string>();
function placeholder(hash: string) {
  let url = placeholders.get(hash);
  if (url === undefined) {
    try {
      const bytes = Uint8Array.from(atob(hash), (c) => c.charCodeAt(0));
      url = thumbHashToDataURL(bytes);
    } catch {
      url = "";
    }
    placeholders.set(hash, url);
  }
  return url;
}

const src = computed(() => coverURL(props.cover.name, props.size));
const background = computed(() => {
  const url = props.cover.thumbhash && placeholder(props.cover.thumbhash);
  return url ? { backgroundImage: `url(${url})` } : {};
});
const loaded = ref(false);
watch(src, () => (loaded.value = false));
</script>

<template>
  <div class="cover-image" :style="background">
    <img
      :src="src"
      alt=""
      loading="lazy"
      decoding="async"
      :class="{ loaded }"
      @load="loaded = true"
    />
  </div>
</template>

<style scoped>
.cover-image {
  width: 100%;
  height: 100%;
  background-size: cover;
  background-position: center;
}
img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
  opacity: 0;
  transition: opacity 0.3s;
}
img.loaded {
  opacity: 1;
}
</style>
