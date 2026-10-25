<template>
  <!-- 全屏封面 -->
  <div
    v-if="settingStore.playerType === 'fullscreen'"
    class="full-screen"
    :style="{
      '--gradient-percent': settingStore.playerFullscreenGradient + '%',
    }"
  >
    <s-image
      :src="getCoverUrl('xl')"
      :alt="musicStore.playSong.title"
      :title="musicStore.playSong.title"
      :lazy="false"
      :width="'100%'"
      :height="'100%'"
    />
  </div>
  <!-- 普通封面 -->
  <div
    v-else
    :class="[
      'player-cover',
      settingStore.playerType,
      { playing: statusStore.playStatus },
    ]"
  >
    <!-- 指针 -->
    <img
      v-if="settingStore.playerType === 'record'"
      class="pointer"
      src="/images/pointer.png?asset"
      alt="pointer"
    />
    <!-- 专辑图片 -->
    <s-image
      :key="getCoverUrl('l')"
      :src="getCoverUrl('l')"
      :observe-visibility="false"
      object-fit="cover"
      class="cover-img"
    />
  </div>
</template>

<script setup lang="ts">
import { useSettingStore, useStatusStore, useMusicStore } from "@/stores";

// 照搬 SPlayer components/Player/PlayerMeta/PlayerCover.vue。去掉本地歌曲封面读取与动态封面：
// 动态封面的视频来自网易接口，Sonora 没有来源。

const musicStore = useMusicStore();
const statusStore = useStatusStore();
const settingStore = useSettingStore();

// 获取封面 URL
const getCoverUrl = (size: "s" | "m" | "l" | "xl" = "l") => {
  return musicStore.getSongCover(size);
};
</script>

<style lang="scss" scoped>
.player-cover {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 70%;
  max-width: 50vh;
  height: auto;
  aspect-ratio: 1 / 1;
  transition:
    transform 0.5s cubic-bezier(0.34, 1.56, 0.64, 1),
    width 0.3s;
  .cover-img {
    width: 100%;
    height: 100%;
    z-index: 1;
    box-shadow: 0 0 20px 10px rgba(0, 0, 0, 0.1);
    transition: opacity 0.1s ease-in-out;
  }
  .dynamic-cover {
    position: absolute;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    border-radius: 32px;
    overflow: hidden;
    z-index: 1;
    opacity: 0;
    transition: opacity 0.8s ease-in-out;
    backface-visibility: hidden;
    transform: translateZ(0);
    &.loaded {
      opacity: 1;
    }
  }
  &.record {
    position: relative;
    max-width: 46vh;
    margin-bottom: 4%;
    .pointer {
      position: absolute;
      width: 30%;
      left: 46%;
      top: -22%;
      transform: rotate(-20deg);
      transform-origin: 10% 10%;
      z-index: 2;
      transition: transform 0.3s;
    }
    .cover-img {
      animation: playerCoverRotate 30s linear infinite;
      animation-play-state: paused;
      border-radius: 50%;
      border: 1vh solid #ffffff30;
      background:
        linear-gradient(black 0%, transparent, black 98%),
        radial-gradient(
          #000 52%,
          #555,
          #000,
          #555,
          #000,
          #555,
          #000,
          #555,
          #000,
          #555,
          #000,
          #555,
          #000,
          #555,
          #000,
          #555,
          #000,
          #555,
          #000,
          #555,
          #000,
          #555,
          #000,
          #555,
          #000,
          #555,
          #000,
          #555,
          #000,
          #555,
          #000,
          #555,
          #000,
          #555,
          #000,
          #555,
          #000,
          #555,
          #000,
          #555,
          #000,
          #555,
          #000,
          #555,
          #000,
          #555,
          #000,
          #555,
          #000,
          #555,
          #000,
          #555,
          #000,
          #555,
          #000,
          #555,
          #000,
          #555,
          #000,
          #555,
          #000,
          #555
        );
      background-clip: content-box;
      // width: 46vh;
      // height: 46vh;
      // min-width: 46vh;
      width: 100%;
      height: 100%;
      display: flex;
      justify-content: center;
      align-items: center;
      :deep(img) {
        border: 1vh solid #ffffff40;
        border-radius: 50%;
        width: 70%;
        height: 70%;
        object-fit: cover;
      }
      .cover-loading {
        position: absolute;
        height: 100%;
        padding-bottom: 0;
        .loading-img {
          top: auto;
          left: auto;
        }
      }
    }
  }
  &.cover {
    border-radius: 32px;
    overflow: hidden;
    transform: scale(0.9);
    &.playing {
      transform: scale(1);
    }
  }
  &.playing {
    .pointer {
      transform: rotate(-8deg);
    }
    .cover-img {
      animation-play-state: running;
    }
  }
}
.full-screen {
  position: fixed;
  left: 0;
  top: 0;
  height: 100vh;
  width: 60vw;
  z-index: 0;
  mask-image: linear-gradient(
    to right,
    #000 var(--gradient-percent),
    transparent 100%
  );
  -webkit-mask-image: linear-gradient(
    to right,
    #000 var(--gradient-percent),
    transparent 100%
  );
  :deep(img) {
    object-fit: cover;
    width: 100%;
    height: 100%;
  }
}
</style>
