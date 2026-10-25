<!-- 歌手跳转 -->
<template>
  <div class="jump-artist">
    <Transition name="fade" mode="out-in">
      <div v-if="artistData?.length" class="ar-list">
        <n-card
          v-for="(item, index) in artistData"
          :key="index"
          class="ar-item"
          hoverable
          @click="jumpArtist(item.id)"
        >
          <n-avatar
            :src="item.cover || '/images/artist.jpg?asset'"
            class="cover"
            round
          />
          <n-text class="name">
            {{
              settingStore.hideBracketedContent
                ? removeBrackets(item.name)
                : item.name
            }}
          </n-text>
        </n-card>
      </div>
      <div v-else class="ar-list">
        <n-skeleton class="ar-item" :repeat="2" text />
      </div>
    </Transition>
  </div>
</template>

<script setup lang="ts">
// 照搬 SPlayer 的 JumpArtist.vue。Sonora 的歌手都是带 ID 的数组，去掉按名称
// 在线搜索的分支；图片取管理后台上传的艺术家图片。
import type { ArtistRef } from "@/api/types";
import { source } from "@/api/source";
import { useSettingStore } from "@/stores";
import { removeBrackets } from "@/utils/format";

const props = defineProps<{
  artist: ArtistRef[];
}>();

const emit = defineEmits<{
  close: [];
}>();

const router = useRouter();
const settingStore = useSettingStore();

const artistData = ref<{ id: number; name: string; cover?: string }[]>([]);

// 获取歌手信息
const getArtistData = async () => {
  if (!props.artist) return;
  const artistPromises = props.artist.map(async (ar) => {
    try {
      const { image } = await source.artist(ar.id);
      return {
        id: ar.id,
        name: ar.name,
        cover: image ? source.coverUrl(image, 100) : undefined,
      };
    } catch (error) {
      console.error(`获取歌手 ${ar.name} (${ar.id}) 的详情失败:`, error);
      return { id: ar.id, name: ar.name, cover: undefined };
    }
  });
  artistData.value = await Promise.all(artistPromises);
};

// 跳转至歌手
const jumpArtist = (id: number) => {
  if (!id) return;
  emit("close");
  router.push({ name: "artist", query: { id } });
};

onMounted(getArtistData);
</script>

<style lang="scss" scoped>
.ar-list {
  .n-alert {
    margin-bottom: 12px;
  }
  :deep(.ar-item) {
    height: 60px;
    margin-bottom: 12px;
    border-radius: 12px;
    cursor: pointer;
    &:last-child {
      margin-bottom: 0;
    }
    .n-card__content {
      padding: 12px 16px;
      display: flex;
      align-items: center;
    }
    .cover {
      margin-right: 8px;
    }
    .name {
      font-size: 16px;
    }
    &:hover {
      border-color: rgba(var(--primary), 0.58);
    }
  }
}
</style>
