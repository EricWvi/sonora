<template>
  <div class="player-control">
    <Transition name="fade" mode="out-in">
      <div
        v-show="statusStore.playerMetaShow"
        class="control-content"
        @click.stop
      >
        <n-flex class="left" align="center">
          <!-- 收起 -->
          <div
            class="menu-icon"
            @click.stop="statusStore.showFullPlayer = false"
          >
            <SvgIcon name="Down" />
          </div>
          <!-- 喜欢歌曲 -->
          <div
            v-if="settingStore.fullscreenPlayerElements.like"
            class="menu-icon"
            @click="
              toLikeSong(
                musicStore.playSong,
                !dataStore.isLikeSong(musicStore.playSong),
              )
            "
          >
            <SvgIcon
              :name="
                dataStore.isLikeSong(musicStore.playSong)
                  ? 'Favorite'
                  : 'FavoriteBorder'
              "
            />
          </div>
          <!-- 添加到歌单 -->
          <div
            v-if="settingStore.fullscreenPlayerElements.addToPlaylist"
            class="menu-icon"
            @click.stop="openPlaylistAdd([musicStore.playSong])"
          >
            <SvgIcon name="AddList" />
          </div>
        </n-flex>
        <div class="center">
          <div class="btn">
            <!-- 随机按钮 -->
            <div
              class="btn-icon mode-icon"
              @click.stop="player.toggleShuffle()"
            >
              <SvgIcon
                :name="statusStore.shuffleIcon"
                :size="20"
                :depth="statusStore.shuffleMode === 'off' ? 3 : 1"
              />
            </div>
            <!-- 上一曲 -->
            <div class="btn-icon" v-debounce="() => player.nextOrPrev('prev')">
              <SvgIcon :size="26" name="SkipPrev" />
            </div>
            <!-- 播放暂停 -->
            <n-button
              :loading="statusStore.playLoading"
              :focusable="false"
              :keyboard="false"
              class="play-pause"
              type="primary"
              strong
              secondary
              circle
              @click.stop="player.playOrPause()"
            >
              <template #icon>
                <Transition name="fade" mode="out-in">
                  <SvgIcon
                    :key="statusStore.playStatus ? 'Pause' : 'Play'"
                    :name="statusStore.playStatus ? 'Pause' : 'Play'"
                    :size="28"
                  />
                </Transition>
              </template>
            </n-button>
            <!-- 下一曲 -->
            <div class="btn-icon" v-debounce="() => player.nextOrPrev('next')">
              <SvgIcon :size="26" name="SkipNext" />
            </div>
            <!-- 循环按钮 -->
            <div class="btn-icon mode-icon" @click.stop="player.toggleRepeat()">
              <SvgIcon
                :name="statusStore.repeatIcon"
                :size="20"
                :depth="statusStore.repeatMode === 'off' ? 3 : 1"
              />
            </div>
          </div>
          <!-- 进度条 -->
          <div class="slider">
            <span @click="toggleTimeFormat">{{ timeDisplay[0] }}</span>
            <PlayerSlider :show-tooltip="false" />
            <div class="time-container">
              <span @click="toggleTimeFormat">{{ timeDisplay[1] }}</span>
            </div>
          </div>
        </div>
        <n-flex class="right" align="center" justify="end">
          <!-- 功能区 -->
          <PlayerRightMenu />
        </n-flex>
      </div>
    </Transition>
  </div>
</template>

<script setup lang="ts">
import { toLikeSong } from "@/utils/auth";
import { usePlayerController } from "@/core/player/PlayerController";
import {
  useDataStore,
  useMusicStore,
  useStatusStore,
  useSettingStore,
} from "@/stores";
import { useTimeFormat } from "@/composables/useTimeFormat";
import { openPlaylistAdd } from "@/utils/modal";

// 照搬 SPlayer components/Player/PlayerControl.vue。去掉下载、评论、私人 FM 的不喜欢、电台判断
// 与自动混音（Sonora 不做）；喜欢在阶段 4.4 接入。

const dataStore = useDataStore();
const musicStore = useMusicStore();
const statusStore = useStatusStore();
const settingStore = useSettingStore();

const player = usePlayerController();

const { timeDisplay, toggleTimeFormat } = useTimeFormat();
</script>

<style lang="scss" scoped>
.player-control {
  position: absolute;
  bottom: 0;
  width: 100%;
  height: 80px;
  overflow: hidden;
  .control-content {
    width: 100%;
    height: 100%;
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    align-items: center;
  }
  .left,
  .right {
    opacity: 1;
    height: 100%;
    padding: 0 30px;
    transition: opacity 0.3s;
    :deep(.menu-icon) {
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
        font-size: 24px;
        color: rgb(var(--main-cover-color));
      }
      &:hover {
        transform: scale(1.1);
        background-color: rgba(var(--main-cover-color), 0.14);
      }
      &:active {
        transform: scale(1);
      }
    }
    :deep(.n-badge-sup) {
      background-color: rgba(var(--main-cover-color), 0.14);
      .n-base-slot-machine {
        color: rgb(var(--main-cover-color));
      }
    }
    :deep(.right-menu) {
      .n-badge-sup {
        background-color: rgba(var(--main-cover-color), 0.14);
        .n-base-slot-machine {
          color: rgb(var(--main-cover-color));
        }
      }
      .quality-tag {
        color: rgb(var(--main-cover-color));
        background-color: transparent !important;
        border-color: rgba(var(--main-cover-color), 0.1) !important;
      }
    }
  }
  .center {
    height: 100%;
    max-height: 80px;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    .btn {
      display: flex;
      flex-direction: row;
      align-items: center;
      .btn-icon {
        display: flex;
        align-items: center;
        justify-content: center;
        width: 38px;
        height: 38px;
        border-radius: 50%;
        will-change: transform;
        transition:
          background-color 0.3s,
          transform 0.3s;
        cursor: pointer;
        margin: 0 4px;
        .n-icon {
          color: rgb(var(--main-cover-color));
        }
        &:hover {
          transform: scale(1.1);
          background-color: rgba(var(--main-cover-color), 0.14);
        }
        &:active {
          transform: scale(1);
        }
      }
      .play-pause {
        --n-width: 44px;
        --n-height: 44px;
        --n-color: rgba(var(--main-cover-color), 0.14);
        --n-color-hover: rgba(var(--main-cover-color), 0.2);
        --n-color-focus: rgba(var(--main-cover-color), 0.2);
        --n-color-pressed: rgba(var(--main-cover-color), 0.12);
        margin: 0 12px;
        transition:
          background-color 0.3s,
          transform 0.3s;
        .n-icon {
          color: rgb(var(--main-cover-color));
          transition: opacity 0.1s ease-in-out;
        }
        :deep(.n-base-loading) {
          color: rgb(var(--main-cover-color));
        }
        &:hover {
          transform: scale(1.1);
        }
        &:active {
          transform: scale(1);
        }
      }
    }
    .slider {
      display: flex;
      flex-direction: row;
      align-items: center;
      width: 100%;
      max-width: 480px;
      font-size: 12px;
      position: relative;
      .n-slider {
        margin: 6px 8px;
        --n-handle-size: 12px;
        --n-rail-height: 4px;
      }
      .time-container {
        position: relative;
        min-width: 40px;
        height: 100%;
        display: flex;
        justify-content: flex-end;
        align-items: center;

        span {
          position: absolute;
          right: 0;
          white-space: nowrap;
          opacity: 0.6;
          cursor: pointer;
        }
      }
    }
  }
  &:hover {
    .left,
    .right {
      opacity: 1;
    }
  }
}
</style>
