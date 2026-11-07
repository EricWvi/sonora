<!-- 私人 FM 卡片。样式照搬 SPlayer components/Player/PlayerComponents/PersonalFM.vue。
与 SPlayer 不同（Eric，2026-11-07）：歌曲来自自己的曲库，按专辑排（backend/internal/store/fm.go）；
“不喜欢”改为“跳过这张专辑”，Sonora 没有垃圾桶；没在听 FM 时跳过只换卡片上的专辑，不打断
正在播放的歌；没有选择分组歌单时，专辑名下面多一行提示，点开设置。 -->
<template>
  <n-card class="personal-fm">
    <!-- 封面 -->
    <Transition name="fade" mode="out-in">
      <n-image
        v-if="!settingStore.hiddenCovers.home"
        :key="song?.id"
        :src="song ? coverUrl(song.cover, 'm') : defaultCovers.song"
        class="cover"
        preview-disabled
        lazy
        @load="coverLoaded"
      >
        <template #placeholder>
          <div class="cover-loading">
            <img
              src="/images/song.jpg?asset"
              class="loading-img"
              alt="loading-img"
            />
          </div>
        </template>
      </n-image>
    </Transition>
    <!-- 信息 -->
    <Transition name="fade" mode="out-in">
      <div
        :key="song?.id"
        :class="['info', { 'no-cover': settingStore.hiddenCovers.home }]"
      >
        <n-text class="name text-hidden">
          {{ song?.title || (loaded ? "还没有可以播放的专辑" : "") }}
        </n-text>
        <div v-if="song?.artists.length" class="artists text-hidden">
          <SvgIcon name="Artist" :depth="3" />
          <n-text v-for="ar in song.artists" :key="ar.id" class="ar">
            {{ ar.name }}
          </n-text>
        </div>
        <div v-if="song?.album" class="album text-hidden">
          <SvgIcon name="Album" :depth="3" />
          <n-text class="album-text">{{ song.album.title }}</n-text>
        </div>
        <n-text
          v-if="needsGroups"
          class="hint text-hidden"
          :depth="3"
          @click.stop="openFmGroupManager"
        >
          只播大专辑，去选分组歌单
        </n-text>
        <!-- 功能 -->
        <n-flex :wrap="false" class="menu" align="center">
          <!-- 跳过这张专辑 -->
          <div class="menu-icon" title="跳过这张专辑" v-debounce="skipAlbum">
            <SvgIcon class="icon" size="18" name="ThumbDown" />
          </div>
          <!-- 播放暂停 -->
          <n-button
            :loading="inFm && statusStore.playLoading"
            :focusable="false"
            class="play"
            strong
            secondary
            circle
            @click.stop="fmPlayOrPause"
          >
            <template #icon>
              <Transition name="fade" mode="out-in">
                <SvgIcon :key="playIcon" :name="playIcon" :size="32" />
              </Transition>
            </template>
          </n-button>
          <!-- 下一曲 -->
          <div class="menu-icon" @click.stop="fmPlayNext">
            <SvgIcon size="26" name="SkipNext" />
          </div>
        </n-flex>
        <!-- 图标 -->
        <div class="radio">
          <SvgIcon :depth="3" name="Radio" />
          <n-text :depth="3">私人FM</n-text>
        </div>
      </div>
    </Transition>
  </n-card>
</template>

<script setup lang="ts">
import type { FmState } from "@/api/types";
import { source } from "@/api/source";
import SvgIcon from "@/components/Global/SvgIcon.vue";
import { FM_LIST, fmCurrentTrack, fmRevision } from "@/core/player/fm";
import { usePlayerController } from "@/core/player/PlayerController";
import { useMusicStore, useSettingStore, useStatusStore } from "@/stores";
import { coverUrl, defaultCovers } from "@/utils/cover";
import { coverLoaded } from "@/utils/helper";
import { openFmGroupManager } from "@/utils/modal";
import { debounce } from "lodash-es";

const musicStore = useMusicStore();
const statusStore = useStatusStore();
const settingStore = useSettingStore();
const player = usePlayerController();

const state = shallowRef<FmState | null>(null);
const loaded = ref<boolean>(false);

// 是否正在听 FM
const inFm = computed(() => musicStore.playPlaylistId === FM_LIST);

// 正在听时显示当前歌曲，否则显示回到 FM 时会继续的那首
const song = computed(() =>
  inFm.value
    ? musicStore.playSong
    : state.value
      ? fmCurrentTrack(state.value)
      : undefined,
);

const needsGroups = computed(
  () => loaded.value && !state.value?.groupIds.length,
);

// 播放图标
const playIcon = computed(() =>
  inFm.value && statusStore.playStatus ? "Pause" : "Play",
);

const load = async () => {
  try {
    state.value = await source.fm();
    loaded.value = true;
  } catch (e) {
    console.error("获取私人 FM 失败：", e);
  }
};

// 播放暂停
const fmPlayOrPause = () => {
  if (inFm.value) player.playOrPause();
  else player.playFm();
};

// 下一曲
const fmPlayNext = debounce(async () => {
  if (!inFm.value) await player.playFm();
  if (inFm.value) await player.nextOrPrev("next");
}, 300);

// 跳过这张专辑
const skipAlbum = async () => {
  const next = await player.fmSkipAlbum(state.value?.albums[0]?.id);
  if (next) state.value = next;
};

// 离开 FM 或改了分组歌单之后，卡片要换成服务端记录的专辑
watch([inFm, fmRevision], load);

// 每次进入都重新读取；keep-alive 首次激活与挂载同时发生，跳过
const isActivated = ref<boolean>(false);
onActivated(() => {
  if (isActivated.value) load();
  isActivated.value = true;
});
onMounted(load);
</script>

<style lang="scss" scoped>
.personal-fm {
  position: relative;
  width: 100%;
  height: 200px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 16px;
  overflow: hidden;
  cursor: pointer;
  :deep(.n-card__content) {
    display: flex;
    align-items: center;
    padding: 20px;
    width: 100%;
  }
  .cover {
    width: 160px;
    height: 160px;
    min-width: 160px;
    border-radius: 8px;
    margin-right: 20px;
    display: flex;
    align-items: center;
    justify-content: center;
    overflow: hidden;
    box-shadow: 0 0 10px 6px #00000008;
    :deep(img) {
      width: 100%;
      height: 100%;
      opacity: 0;
      transition: opacity 0.35s ease-in-out;
    }
  }
  .info {
    height: 100%;
    display: flex;
    flex-direction: column;
    .n-text {
      line-height: normal;
    }
    .name {
      font-size: 22px;
      font-weight: bold;
    }
    .artists {
      margin-top: 2px;
      font-size: 14px;
      .n-icon {
        font-size: 18px;
        margin-right: 4px;
        transform: translateY(3px);
      }
      .ar {
        display: inline-flex;
        transition: opacity 0.3s;
        opacity: 0.6;
        &::after {
          content: "/";
          margin: 0 4px;
        }
        &:last-child {
          &::after {
            display: none;
          }
        }
      }
    }
    .album {
      font-size: 14px;
      .n-icon {
        font-size: 18px;
        margin-right: 4px;
        transform: translateY(3px);
      }
      .album-text {
        transition: opacity 0.3s;
        opacity: 0.6;
      }
    }
    .hint {
      margin-top: 2px;
      font-size: 12px;
      cursor: pointer;
      &:hover {
        color: var(--primary-hex);
      }
    }
    .menu {
      margin-top: auto;
      .play {
        width: 46px;
        height: 46px;
        .n-icon {
          color: var(--primary-hex);
          transition: opacity 0.1s ease-in-out;
        }
      }
      .menu-icon {
        display: flex;
        align-items: center;
        justify-content: center;
        width: 38px;
        height: 38px;
        border-radius: 50%;
        color: var(--primary-hex);
        transition:
          background-color 0.3s,
          transform 0.3s;
        cursor: pointer;
        &:hover {
          transform: scale(1.1);
          background-color: rgba(var(--primary), 0.08);
        }
        &:active {
          transform: scale(1);
        }
      }
    }
    &.no-cover {
      width: 100%;
      align-items: center;
      justify-content: center;
      text-align: center;
      .name {
        font-size: 24px;
        margin-bottom: 6px;
      }
      .artists {
        margin-top: 0;
        font-size: 15px;
        opacity: 0.8;
      }
      .menu {
        margin-top: 28px;
        justify-content: center;
        gap: 8px;
        .menu-icon {
          width: 42px;
          height: 42px;
          background-color: rgba(var(--primary), 0.06);
          &:hover {
            background-color: rgba(var(--primary), 0.12);
          }
        }
        .play {
          width: 52px;
          height: 52px;
        }
      }
    }
  }
  .radio {
    position: absolute;
    right: 20px;
    bottom: 14px;
    display: flex;
    align-items: center;
    .n-icon {
      margin-right: 4px;
      transform: translateY(-1px);
    }
  }
  &:hover {
    border-color: rgba(var(--primary), 0.6);
  }
  @media (max-width: 1200px) {
    .radio {
      display: none;
    }
  }
}
</style>
