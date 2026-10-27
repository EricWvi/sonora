<template>
  <n-flex :size="8" align="center" class="right-menu">
    <!-- 其他控制：只剩均衡器，改为直接打开（Eric 确认，2026-10-27） -->
    <div
      v-if="settingStore.fullscreenPlayerElements.moreSettings"
      class="menu-icon hidden"
      title="均衡器"
      @click.stop="openEqualizer"
    >
      <SvgIcon name="Eq" />
    </div>
    <!-- 音量 -->
    <n-popover :show-arrow="false" :style="{ padding: 0 }">
      <template #trigger>
        <div
          class="menu-icon hidden"
          @click.stop="player.toggleMute"
          @wheel="player.setVolume"
        >
          <SvgIcon :name="statusStore.playVolumeIcon" />
        </div>
      </template>
      <div class="volume-change" @wheel="player.setVolume">
        <n-slider
          v-model:value="statusStore.playVolume"
          :tooltip="false"
          :min="0"
          :max="1"
          :step="0.01"
          vertical
          @update:value="(val: number) => player.setVolume(val)"
        />
        <n-text class="slider-num hidden"
          >{{ statusStore.playVolumePercent }}%</n-text
        >
      </div>
    </n-popover>
    <!-- 播放列表 -->
    <n-badge
      :value="dataStore.playList?.length ?? 0"
      :show="settingStore.showPlaylistCount"
      :max="9999"
      :style="{
        marginRight: settingStore.showPlaylistCount ? '12px' : null,
      }"
    >
      <div
        class="menu-icon"
        @click.stop="statusStore.playListShow = !statusStore.playListShow"
      >
        <SvgIcon name="PlayList" />
      </div>
    </n-badge>
  </n-flex>
</template>

<script setup lang="ts">
import { usePlayerController } from "@/core/player/PlayerController";
import { useDataStore, useSettingStore, useStatusStore } from "@/stores";
import { openEqualizer } from "@/utils/modal";

// 照搬 SPlayer components/Player/PlayerRightMenu.vue。去掉音质与桌面歌词；“其他控制”只保留
// 均衡器，不做自动关闭、AB 循环与播放速度（2026-10-25）。只剩一项的悬停菜单像提示，点按钮
// 却没有反应，改为均衡器图标、点击直接打开（Eric 确认，2026-10-27）。

const dataStore = useDataStore();
const statusStore = useStatusStore();
const settingStore = useSettingStore();
const player = usePlayerController();
</script>

<style scoped lang="scss">
.right-menu {
  .menu-icon {
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 8px;
    border-radius: 8px;
    transition:
      background-color 0.3s,
      transform 0.3s;
    cursor: pointer;
    .n-icon {
      font-size: 22px;
      color: var(--primary-hex);
    }
    &:hover {
      transform: scale(1.1);
      background-color: rgba(var(--primary), 0.28);
    }
    &:active {
      transform: scale(1);
    }
  }
  :deep(.n-badge-sup) {
    background-color: rgba(var(--primary), 0.28);
    backdrop-filter: blur(20px);
    // font-size: 10px;
    .n-base-slot-machine {
      color: var(--primary-hex);
    }
  }
  @media (max-width: 810px) {
    .hidden {
      display: none;
    }
  }
}
.volume-change {
  padding: 12px;
  display: flex;
  flex-direction: column;
  height: 180px;
  width: 58px;
  align-items: center;
  .slider-num {
    margin-top: 8px;
    font-size: 13px;
    white-space: nowrap;
  }
}
</style>
