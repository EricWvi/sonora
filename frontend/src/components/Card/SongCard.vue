<!-- 歌曲列表的一行。照搬 SPlayer components/Card/SongCard.vue，歌曲是播放端 API 的 Track；
不迁入音质、原唱翻唱、特权、云盘、MV、脏标标签与电台、移动端分支。 -->
<template>
  <div class="song-card">
    <div
      :class="['song-content', { play: sameTrack(musicStore.playSong, song) }]"
    >
      <!-- 序号 -->
      <div class="num" @dblclick.stop>
        <n-text v-if="!sameTrack(musicStore.playSong, song)" depth="3">
          {{ index + 1 }}
        </n-text>
        <SvgIcon v-else :size="22" name="Music" />
        <!-- 播放暂停 -->
        <SvgIcon
          :size="28"
          :name="statusStore.playStatus ? 'Pause' : 'Play'"
          class="status"
          @click="player.playOrPause()"
        />
        <!-- 播放 -->
        <SvgIcon
          :size="28"
          name="Play"
          class="play"
          @click="player.addNextSong(song, true)"
        />
      </div>
      <!-- 标题 -->
      <div class="title">
        <!-- 封面 -->
        <s-image
          v-if="!hiddenCover"
          :key="coverSrc"
          :src="coverSrc"
          class="cover"
        />
        <!-- 信息 -->
        <n-flex size="small" class="info" vertical>
          <!-- 名称 -->
          <div class="name">
            <n-ellipsis
              :line-clamp="1"
              :tooltip="{
                placement: 'top',
                width: 'trigger',
              }"
              class="name-text"
            >
              {{
                settingStore.hideBracketedContent
                  ? removeBrackets(song?.title)
                  : song?.title || "未知曲目"
              }}
            </n-ellipsis>
          </div>
          <n-flex :size="4" :wrap="false" class="desc" align="center">
            <!-- 歌手 -->
            <template v-if="settingStore.showSongArtist">
              <div v-if="song.artists.length" class="artists">
                <n-text
                  v-for="ar in song.artists"
                  :key="ar.id"
                  class="ar"
                  @click="openJumpArtist(song.artists, ar.id, isOnline(song))"
                >
                  {{
                    settingStore.hideBracketedContent
                      ? removeBrackets(ar.name)
                      : ar.name
                  }}
                </n-text>
              </div>
              <div v-else class="artists">
                <n-text class="ar">未知艺术家</n-text>
              </div>
            </template>
          </n-flex>
        </n-flex>
      </div>
      <!-- 专辑 -->
      <div
        v-if="!hiddenAlbum && !isSmallScreen && settingStore.showSongAlbum"
        class="album text-hidden"
      >
        <n-text
          v-if="song.album"
          class="album-text"
          @click="router.push(albumRoute(song))"
        >
          {{ albumName }}
        </n-text>
        <n-text v-else class="album-text">
          {{ albumName }}
        </n-text>
      </div>
      <!-- 操作 -->
      <div
        v-if="settingStore.showSongOperations"
        class="actions"
        @click.stop
        @dblclick.stop
      >
        <!-- 喜欢歌曲 -->
        <SvgIcon
          :name="dataStore.isLikeSong(song) ? 'Favorite' : 'FavoriteBorder'"
          :size="20"
          @click.stop="toLikeSong(song, !dataStore.isLikeSong(song))"
          @delclick.stop
        />
      </div>
      <!-- 时长 -->
      <n-text
        v-if="!isSmallScreen && settingStore.showSongDuration"
        class="meta"
        depth="3"
      >
        {{ msToTime(song.durationMs) }}
      </n-text>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { Track } from "@/api/types";
import {
  useStatusStore,
  useMusicStore,
  useDataStore,
  useSettingStore,
} from "@/stores";
import { openJumpArtist } from "@/utils/modal";
import { removeBrackets } from "@/utils/format";
import { toLikeSong } from "@/utils/auth";
import { albumRoute, isOnline, sameTrack } from "@/utils/track";
import { coverUrl } from "@/utils/cover";
import { msToTime } from "@/utils/time";
import { usePlayerController } from "@/core/player/PlayerController";

const props = defineProps<{
  // 歌曲
  song: Track;
  // 索引
  index: number;
  // 隐藏信息
  hiddenCover?: boolean;
  hiddenAlbum?: boolean;
}>();

// SPlayer 的 useMobile：小屏（< 768px）隐藏专辑与时长；Sonora 最小宽度 990px，不会出现
const isSmallScreen = useMediaQuery("(max-width: 767.98px)");
const router = useRouter();
const dataStore = useDataStore();
const musicStore = useMusicStore();
const statusStore = useStatusStore();
const settingStore = useSettingStore();

const player = usePlayerController();

// 歌曲数据
const song = toRef(props, "song");

// 封面：SPlayer 的 coverSize.s
const coverSrc = computed(() => coverUrl(song.value.cover, "s"));

// 专辑名称
const albumName = computed(() => {
  const name = song.value.album?.title;
  return (
    (settingStore.hideBracketedContent ? removeBrackets(name) : name) ||
    "未知专辑"
  );
});
</script>

<style lang="scss" scoped>
.song-card {
  height: 90px;
  cursor: pointer;
  .song-content {
    display: flex;
    flex-direction: row;
    align-items: center;
    justify-content: space-between;
    padding: 8px 12px;
    flex: 1;
    border-radius: 12px;
    border: 2px solid rgba(var(--primary), 0.12);
    background-color: var(--surface-container-hex);
    transition:
      transform 0.1s,
      background-color 0.3s var(--n-bezier),
      border-color 0.3s var(--n-bezier);
    &.play {
      border-color: rgba(var(--primary), 0.58);
      background-color: rgba(var(--primary), 0.28);
    }
    // &:active {
    //   transform: scale(0.99);
    // }
    &:hover {
      border-color: rgba(var(--primary), 0.58);
      .num {
        .n-text,
        .n-icon {
          opacity: 0;
        }
        .play {
          opacity: 1;
          transform: scale(1);
        }
      }
      &.play {
        .num {
          .play {
            display: none;
          }
          .status {
            opacity: 1;
            transform: scale(1);
          }
        }
      }
    }
  }
  .num {
    position: relative;
    display: flex;
    justify-content: center;
    align-items: center;
    width: 40px;
    min-width: 40px;
    font-weight: bold;
    margin-right: 12px;
    .n-icon {
      transition:
        opacity 0.3s,
        transform 0.3s;
      :deep(.svg-container) {
        color: var(--primary-hex);
      }
    }
    .status,
    .play {
      position: absolute;
      opacity: 0;
      transform: scale(0.8);
      transition:
        opacity 0.3s,
        transform 0.3s;
      &:active {
        opacity: 0.6 !important;
      }
    }
  }
  .title {
    flex: 1;
    min-width: 0;
    display: flex;
    align-items: center;
    padding: 4px 20px 4px 0;
    .cover {
      width: 50px;
      height: 50px;
      min-width: 50px;
      border-radius: 8px;
      margin-right: 12px;
      display: flex;
      align-items: center;
      justify-content: center;
      overflow: hidden;
    }
    .info {
      min-width: 0;
      .name {
        display: flex;
        flex-direction: row;
        align-items: center;
        line-height: normal;
        font-size: 16px;
      }
      .desc {
        min-width: 0;
        margin-top: 2px;
        font-size: 13px;
        .n-tag {
          --n-height: 18px;
          font-size: 10px;
          cursor: pointer;
          pointer-events: none;
          &:last-child {
            margin-right: 0;
          }
        }
        .quality {
          font-size: 10px;
        }
        .cloud {
          padding: 0 10px;
          align-items: center;
          justify-content: center;
          :deep(.n-tag__icon) {
            margin-right: 0;
            width: 100%;
          }
          .n-icon {
            font-size: 12px;
            color: var(--n-text-color);
          }
        }
        .mv {
          pointer-events: auto;
        }
      }
      .artists {
        flex: 1;
        min-width: 0;
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
        .ar {
          display: inline;
          transition: opacity 0.3s;
          opacity: 0.6;
          cursor: pointer;
          &::after {
            content: "/";
            margin: 0 4px;
          }
          &:last-child {
            &::after {
              display: none;
            }
          }
          &:hover {
            opacity: 0.8;
          }
        }
      }
    }
    .sort {
      margin-left: 6px;
      &::after {
        content: " )";
      }
      &::before {
        content: "( ";
      }
    }
  }
  .album {
    flex: 1;
    min-width: 0;
    line-clamp: 2;
    -webkit-line-clamp: 2;
    padding-right: 20px;
    &:hover {
      .album-text {
        color: var(--primary-hex);
      }
    }
  }
  .actions {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 40px;
    .n-icon {
      color: var(--primary-hex);
      transition: transform 0.3s;
      cursor: pointer;
      &:hover {
        transform: scale(1.15);
      }
      &:active {
        transform: scale(1);
      }
    }
  }
  .meta {
    width: 50px;
    font-size: 13px;
    text-align: center;
    &.size {
      width: 60px;
    }
    &.date {
      width: 80px;
    }
  }
  &.header {
    border: none;
    background-color: transparent;
    .n-text {
      opacity: 0.6;
    }
    .title {
      position: relative;
      padding: 0 20px 0 0;
      &.has-sort {
        &::after {
          content: "";
          position: absolute;
          opacity: 0;
          top: 0;
          left: -8px;
          width: 100%;
          height: 100%;
          border-radius: 8px;
          background-color: rgba(var(--primary), 0.08);
          transition: opacity 0.3s;
        }
        &:hover {
          &::after {
            opacity: 1;
          }
        }
      }
    }
  }
}
</style>
