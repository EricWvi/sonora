<template>
  <Teleport to="body">
    <Transition :name="settingStore.playerExpandAnimation" mode="out-in">
      <div
        v-if="statusStore.showFullPlayer"
        :style="{
          cursor: statusStore.playerMetaShow ? 'auto' : 'none',
          '--lyric-blend-mode': settingStore.lyricsBlendMode,
        }"
        class="full-player"
        @mouseleave="playerLeave"
        @mousemove="playerMove"
        @click="playerMove"
      >
        <!-- 背景 -->
        <PlayerBackground />
        <!-- 菜单 -->
        <PlayerMenu @mouseenter.stop="stopHide" @mouseleave.stop="resumeHide" />
        <!-- 全屏封面 -->
        <PlayerCover v-if="showFullScreenCover" />
        <!-- 主内容 -->
        <Transition name="zoom" mode="out-in">
          <div
            :key="playerContentKey"
            :class="['player-content', playerContentClasses]"
            @mousemove="playerMove"
          >
            <!-- 左侧封面和数据 -->
            <Transition name="zoom">
              <div
                v-if="showLeftContent"
                :key="musicStore.playSong.id"
                class="content-left"
                :style="layoutStyles.left"
              >
                <PlayerCover />
                <PlayerData :center="playerDataCenter" />
              </div>
            </Transition>
            <!-- 右侧歌词 -->
            <div class="content-right" :style="layoutStyles.right">
              <PlayerData
                v-if="showRightPlayerData"
                :center="pureLyricMode || noLrc"
                :light="!(isFullscreenType && noLrc)"
              />
              <PlayerLyric v-if="!noLrc" />
            </div>
          </div>
        </Transition>
        <!-- 控制中心 -->
        <PlayerControl
          @mouseenter.stop="stopHide"
          @mouseleave.stop="resumeHide"
        />
        <!-- 音乐频谱 -->
        <PlayerSpectrum
          v-if="settingStore.showSpectrums"
          :color="
            statusStore.mainColor
              ? `rgb(${statusStore.mainColor})`
              : 'rgb(239 239 239)'
          "
          :show="!statusStore.playerMetaShow"
          :height="60"
        />
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { useStatusStore, useMusicStore, useSettingStore } from "@/stores";

// 照搬 SPlayer components/Player/FullPlayer.vue。去掉评论（含只在评论时出现的顶部实时歌词）、
// 移动版播放器（最小宽度 990px）与 Electron 的阻止息屏。

const musicStore = useMusicStore();
const statusStore = useStatusStore();
const settingStore = useSettingStore();

/** 封面主颜色 */
const mainCoverColor = useCssVar(
  "--main-cover-color",
  document.documentElement,
);

/** 播放器样式是否为全屏封面 */
const isFullscreenType = computed(
  () => settingStore.playerType === "fullscreen",
);

/** 没有歌词 */
const noLrc = computed<boolean>(() => {
  const noNormalLrc = !musicStore.isHasLrc;
  const noYrcAvailable = !musicStore.isHasYrc || !settingStore.showWordLyrics;
  return noNormalLrc && noYrcAvailable;
});

/** 是否处于纯净歌词模式 */
const pureLyricMode = computed<boolean>(
  () => statusStore.pureLyricMode && musicStore.isHasLrc,
);

/** 主内容 key */
const playerContentKey = computed(
  () => `${musicStore.playSong.id}-${statusStore.pureLyricMode}`,
);

/** 主内容 class */
const playerContentClasses = computed(() => ({
  "no-lrc": noLrc.value,
  "full-screen": isFullscreenType.value,
  pure: pureLyricMode.value && musicStore.isHasLrc,
}));

/** 左右布局样式 */
const layoutStyles = computed(() => {
  const ratio = isFullscreenType.value ? 50 : settingStore.playerStyleRatio;
  return {
    left: { width: `${ratio}%`, minWidth: `${ratio}%` },
    right: { width: `${100 - ratio}%`, maxWidth: `${100 - ratio}%` },
  };
});

/** 是否显示左侧封面区域 */
const showLeftContent = computed(
  () => !pureLyricMode.value && !isFullscreenType.value,
);

/** 是否显示右侧 PlayerData */
const showRightPlayerData = computed(
  () => (pureLyricMode.value && musicStore.isHasLrc) || isFullscreenType.value,
);

/** 是否显示全屏封面 */
const showFullScreenCover = computed(
  () => isFullscreenType.value && !pureLyricMode.value,
);

/** 数据是否居中 */
const playerDataCenter = computed<boolean>(
  () =>
    !musicStore.isHasLrc ||
    statusStore.pureLyricMode ||
    settingStore.playerType === "record",
);

const {
  isPending,
  start: startShow,
  stop: stopShow,
} = useTimeoutFn(() => {
  if (settingStore.autoHidePlayerMeta) {
    statusStore.playerMetaShow = false;
  }
}, 3000);

/** 鼠标是否在操作区域（菜单/控制栏） */
const inControlArea = ref(false);

const playerMove = useThrottleFn(
  () => {
    statusStore.playerMetaShow = true;
    if (
      settingStore.autoHidePlayerMeta &&
      !isPending.value &&
      !inControlArea.value
    ) {
      startShow();
    }
  },
  300,
  false,
);

const stopHide = () => {
  inControlArea.value = true;
  stopShow();
  statusStore.playerMetaShow = true;
};

const resumeHide = () => {
  inControlArea.value = false;
  if (settingStore.autoHidePlayerMeta) {
    startShow();
  }
};

const playerLeave = () => {
  if (settingStore.autoHidePlayerMeta) {
    statusStore.playerMetaShow = false;
    stopShow();
  }
};

watch(
  () => statusStore.mainColor,
  (newVal) => {
    mainCoverColor.value = newVal;
  },
);

onMounted(() => {
  mainCoverColor.value = statusStore.mainColor;
});

onBeforeUnmount(() => {
  stopShow();
});
</script>

<style lang="scss" scoped>
.full-player {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  color: rgb(var(--main-cover-color));
  background-color: #00000060;
  backdrop-filter: blur(80px);
  overflow: hidden;
  z-index: 1000;
  .player-content {
    position: absolute;
    display: flex;
    flex-direction: row;
    justify-content: center;
    align-items: center;
    width: 100%;
    height: calc(100vh - 160px);
    transition:
      opacity 0.3s cubic-bezier(0.34, 1.56, 0.64, 1),
      transform 0.3s cubic-bezier(0.34, 1.56, 0.64, 1);
    .content-left {
      position: absolute;
      left: 0;
      flex: 1;
      height: 100%;
      display: flex;
      flex-direction: column;
      align-items: center;
      justify-content: center;
      transition:
        width 0.5s cubic-bezier(0.34, 1.56, 0.64, 1),
        opacity 0.5s cubic-bezier(0.34, 1.56, 0.64, 1),
        transform 0.5s cubic-bezier(0.34, 1.56, 0.64, 1);
    }
    .content-right {
      position: absolute;
      right: 0;
      flex: 1;
      height: 100%;
      display: flex;
      flex-direction: column;
      mix-blend-mode: var(--lyric-blend-mode);
      transition:
        width 0.5s cubic-bezier(0.34, 1.56, 0.64, 1) 0.5s,
        opacity 0.3s ease;
      .player-data {
        margin-top: 0;
        margin-bottom: 26px;
      }
    }
    &.pure {
      .content-right {
        align-items: center;
        width: 100% !important;
        max-width: 100% !important;
      }
    }
    &.no-lrc {
      &:not(.full-screen) {
        .content-left {
          width: 50% !important;
          transform: translateX(50%);
        }
        .content-right {
          opacity: 0;
          pointer-events: none;
        }
      }
      &.full-screen {
        .content-right {
          .player-data {
            width: 100%;
            max-width: 100%;
            transform: translateY(30vh);
          }
        }
      }
    }
  }
}
</style>
