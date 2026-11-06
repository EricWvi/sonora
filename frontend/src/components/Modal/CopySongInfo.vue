<!-- 歌曲详情复制，照搬 SPlayer components/Modal/CopySongInfo.vue，只用于在线歌曲（Eric，2026-11-06）。
与 SPlayer 不同：直接用列表中的歌曲信息，不再请求网易云，所以没有别名，发布时间只有年份（“发行年份”）。 -->
<template>
  <div class="copy-song-info">
    <div>
      <n-scrollbar style="max-height: 70vh">
        <n-form :size="'small'" ref="formRef">
          <n-form-item label="歌曲名称">
            <n-input-group>
              <n-input
                :value="song.title"
                readonly
                placeholder="暂无歌曲名称"
              />
              <n-button
                type="primary"
                strong
                secondary
                @click="copyText(song.title, '歌曲名称')"
              >
                <template #icon>
                  <SvgIcon name="Copy" />
                </template>
              </n-button>
            </n-input-group>
          </n-form-item>
          <n-divider class="divider"> 制作人员 </n-divider>
          <template v-for="(artist, index) in song.artists" :key="index">
            <n-grid :cols="24" :x-gap="12">
              <n-form-item-gi
                :span="14"
                :label="song.artists.length > 1 ? `歌手 ${index + 1}` : '歌手'"
              >
                <n-input-group>
                  <n-input :value="artist.name" readonly />
                  <n-button
                    type="primary"
                    strong
                    secondary
                    @click="copyText(artist.name, '歌手名称')"
                  >
                    <template #icon>
                      <SvgIcon name="Copy" />
                    </template>
                  </n-button>
                </n-input-group>
              </n-form-item-gi>
              <n-form-item-gi :span="10" label="ID" v-if="artist.id">
                <n-input-group>
                  <n-input :value="String(artist.id)" readonly />
                  <n-button
                    type="primary"
                    strong
                    secondary
                    @click="copyText(String(artist.id), '歌手ID')"
                  >
                    <template #icon>
                      <SvgIcon name="Copy" />
                    </template>
                  </n-button>
                </n-input-group>
              </n-form-item-gi>
            </n-grid>
          </template>
          <n-grid :cols="24" :x-gap="12" v-if="song.album">
            <n-form-item-gi :span="14" label="专辑">
              <n-input-group>
                <n-input
                  :value="song.album.title"
                  readonly
                  placeholder="暂无专辑信息"
                />
                <n-button
                  type="primary"
                  strong
                  secondary
                  @click="copyText(song.album.title, '专辑名称')"
                >
                  <template #icon>
                    <SvgIcon name="Copy" />
                  </template>
                </n-button>
              </n-input-group>
            </n-form-item-gi>
            <n-form-item-gi :span="10" label="ID" v-if="song.album.id">
              <n-input-group>
                <n-input :value="String(song.album.id)" readonly />
                <n-button
                  type="primary"
                  strong
                  secondary
                  @click="copyText(String(song.album.id), '专辑ID')"
                >
                  <template #icon>
                    <SvgIcon name="Copy" />
                  </template>
                </n-button>
              </n-input-group>
            </n-form-item-gi>
          </n-grid>
          <n-divider class="divider"> 歌曲信息 </n-divider>
          <n-grid :cols="24" :x-gap="12">
            <n-form-item-gi :span="12" label="歌曲 ID">
              <n-input-group>
                <n-input :value="String(song.id || '')" readonly />
                <n-button
                  type="primary"
                  strong
                  secondary
                  @click="copyText(String(song.id), '歌曲ID')"
                >
                  <template #icon>
                    <SvgIcon name="Copy" />
                  </template>
                </n-button>
              </n-input-group>
            </n-form-item-gi>
            <n-form-item-gi :span="12" label="时长">
              <n-input-group>
                <n-input :value="duration" readonly />
                <n-button
                  type="primary"
                  strong
                  secondary
                  @click="copyText(duration, '时长')"
                >
                  <template #icon>
                    <SvgIcon name="Copy" />
                  </template>
                </n-button>
              </n-input-group>
            </n-form-item-gi>
          </n-grid>
          <n-grid :cols="24" :x-gap="12">
            <n-form-item-gi :span="24" label="发行年份" v-if="song.year">
              <n-input-group>
                <n-input :value="String(song.year)" readonly />
                <n-button
                  type="primary"
                  strong
                  secondary
                  @click="copyText(String(song.year), '发行年份')"
                >
                  <template #icon>
                    <SvgIcon name="Copy" />
                  </template>
                </n-button>
              </n-input-group>
            </n-form-item-gi>
          </n-grid>
          <n-form-item label="歌曲链接">
            <n-input-group>
              <n-input :value="songLink" readonly placeholder="暂无链接" />
              <n-button
                type="primary"
                strong
                secondary
                @click="copyText(songLink, '链接')"
              >
                <template #icon>
                  <SvgIcon name="Copy" />
                </template>
              </n-button>
            </n-input-group>
          </n-form-item>
        </n-form>
      </n-scrollbar>
    </div>
    <n-button block @click="handleCopyAll" type="primary" secondary>
      复制全部信息
    </n-button>
  </div>
</template>

<script setup lang="ts">
import type { Track } from "@/api/types";
import { copyData } from "@/utils/helper";
import { onlineLink } from "@/utils/onlineLink";
import { msToTime } from "@/utils/time";

const props = defineProps<{ song: Track }>();

const duration = computed(() =>
  props.song.durationMs ? msToTime(props.song.durationMs) : "",
);
const songLink = computed(() => onlineLink("song", props.song.id));

const copyText = (text: string | undefined, label: string) => {
  if (text) copyData(text, `已复制${label}`);
};

// 复制全部
const handleCopyAll = () => {
  const { song } = props;
  const lines = [
    `歌曲：${song.title}`,
    `歌手：${song.artists.map((a) => `${a.name}${a.id ? ` (ID: ${a.id})` : ""}`).join(" / ")}`,
    song.album
      ? `专辑：${song.album.title}${song.album.id ? ` (ID: ${song.album.id})` : ""}`
      : "",
    `歌曲ID：${song.id}`,
    duration.value ? `时长：${duration.value}` : "",
    song.year ? `发行年份：${song.year}` : "",
    songLink.value ? `链接：${songLink.value}` : "",
  ].filter((line) => line);
  copyData(lines, "已复制全部信息");
};
</script>

<style lang="scss" scoped>
.copy-song-info {
  display: flex;
  flex-direction: column;
  width: 100%;
  .divider {
    font-size: 14px;
    margin: 16px 0 12px 0;
  }
}
</style>
