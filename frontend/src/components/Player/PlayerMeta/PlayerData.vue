<template>
  <div
    :class="['player-data', settingStore.playerType, { center, light }]"
    :style="{ marginLeft: leftMargin }"
  >
    <!-- 名称 -->
    <div class="name">
      <span class="name-text text-hidden">
        {{
          settingStore.hideBracketedContent
            ? removeBrackets(musicStore.playSong.title)
            : musicStore.playSong.title || "未知曲目"
        }}
      </span>
    </div>
    <n-flex :align="center ? 'center' : undefined" size="small" vertical>
      <!-- 歌手 -->
      <div class="artists">
        <SvgIcon :depth="3" name="Artist" size="20" />
        <div class="ar-list">
          <span
            v-for="ar in musicStore.playSong.artists"
            :key="ar.id"
            class="ar"
            @click="jumpPage({ name: 'artist', query: { id: ar.id } })"
          >
            {{
              settingStore.hideBracketedContent
                ? removeBrackets(ar.name)
                : ar.name
            }}
          </span>
        </div>
      </div>
      <!-- 专辑 -->
      <div class="album">
        <SvgIcon :depth="3" name="Album" size="20" />
        <span
          v-if="musicStore.playSong.album"
          class="name-text text-hidden"
          @click="
            jumpPage({
              name: 'album',
              query: { id: musicStore.playSong.album.id },
            })
          "
        >
          {{
            (settingStore.hideBracketedContent
              ? removeBrackets(musicStore.playSong.album.title)
              : musicStore.playSong.album.title) || "未知专辑"
          }}
        </span>
        <span v-else class="name-text text-hidden">未知专辑</span>
      </div>
    </n-flex>
  </div>
</template>

<script setup lang="ts">
import type { RouteLocationRaw } from "vue-router";
import { useMusicStore, useStatusStore, useSettingStore } from "@/stores";
import { debounce } from "lodash-es";
import { removeBrackets } from "@/utils/format";

// 照搬 SPlayer components/Player/PlayerMeta/PlayerData.vue。歌名下的别名与整行标签（音质、
// 歌词类型与来源、音源）按 Eric 的决定不显示；去掉云盘与解锁提示、电台。

const props = defineProps<{
  /** 数据居中 */
  center?: boolean;
  /** 少量数据模式 */
  light?: boolean;
}>();

const router = useRouter();
const musicStore = useMusicStore();
const statusStore = useStatusStore();
const settingStore = useSettingStore();

// 左侧外边距
const leftMargin = computed(() => {
  if (props.center || !props.light) return "0px";
  const offset = settingStore.lyricHorizontalOffset;
  return settingStore.useAMLyrics ? `${offset + 40}px` : `${offset + 10}px`;
});

const jumpPage = debounce(
  (go: RouteLocationRaw) => {
    if (!go) return;
    statusStore.showFullPlayer = false;
    router.push(go);
  },
  300,
  {
    leading: true,
    trailing: false,
  },
);
</script>

<style lang="scss" scoped>
.player-data {
  display: flex;
  flex-direction: column;
  width: 70%;
  max-width: 50vh;
  margin-top: 24px;
  padding: 0 2px;
  // mix-blend-mode: plus-lighter;
  .n-icon {
    color: rgb(var(--main-cover-color));
  }
  .name {
    position: relative;
    display: flex;
    align-items: center;
    margin-left: 4px;
    .name-text {
      font-size: 26px;
      font-weight: bold;
    }
    .n-icon {
      margin-left: 12px;
      transform: translateY(1px);
      cursor: pointer;
    }
  }
  .alia {
    margin: 6px 0 6px 4px;
    opacity: 0.6;
    font-size: 18px;
    line-clamp: 1;
    -webkit-line-clamp: 1;
  }
  .artists {
    display: flex;
    align-items: center;
    .n-icon {
      margin-right: 4px;
    }
    .ar-list {
      display: -webkit-box;
      line-clamp: 2;
      -webkit-box-orient: vertical;
      -webkit-line-clamp: 2;
      overflow: hidden;
      word-break: break-all;
      .ar {
        font-size: 16px;
        opacity: 0.7;
        display: inline-flex;
        transition: opacity 0.3s;
        cursor: pointer;
        &::after {
          content: "/";
          margin: 0 4px;
          transition: none;
        }
        &:last-child {
          &::after {
            display: none;
          }
        }
        &:hover {
          opacity: 1;
        }
      }
    }
  }
  .album,
  .dj {
    font-size: 16px;
    display: flex;
    align-items: center;
    .n-icon {
      margin-right: 4px;
    }
    .name-text {
      opacity: 0.7;
      transition: opacity 0.3s;
      line-clamp: 1;
      -webkit-line-clamp: 1;
      cursor: pointer;
      &:hover {
        opacity: 1;
      }
    }
  }
  .play-meta {
    padding: 4px 4px;
    opacity: 0.6;
    .meta-item {
      font-size: 12px;
      border-radius: 8px;
      padding: 2px 6px;
      border: 1px solid rgba(var(--main-cover-color), 0.6);
      &.clickable {
        cursor: pointer;
        transition: all 0.2s ease;
        &:hover {
          background-color: rgba(var(--main-cover-color), 0.08);
          border-color: rgb(var(--main-cover-color));
        }
        &.loading {
          opacity: 0.6;
          cursor: wait;
        }
      }
    }
  }
  &.record {
    width: 100%;
    padding: 0 80px 0 24px;
    .name {
      .name-text {
        font-size: 30px;
      }
      .extra-info {
        position: absolute;
        right: -34px;
        display: flex;
        align-items: center;
        justify-content: center;
      }
    }
    @media (max-width: 990px) {
      padding: 0 2px;
    }
  }
  &.center {
    align-items: center;
    padding: 0 40px;
    .name {
      text-align: center;
    }
  }
  &.light {
    .name {
      .name-text {
        line-clamp: 1;
        -webkit-line-clamp: 1;
      }
      .extra-info {
        display: none;
      }
    }
    .alia {
      display: none;
    }
  }
}
.player-tip {
  max-width: 240px;
  padding: 12px 20px;
  border-radius: 12px;
  color: rgb(var(--main-cover-color));
  background-color: rgba(var(--main-cover-color), 0.18);
  backdrop-filter: blur(10px);
}
</style>
