<template>
  <div class="create-playlist">
    <n-form ref="formRef" :model="formData" :rules="formRules">
      <n-form-item label="歌单名称" path="name">
        <n-input v-model:value="formData.name" placeholder="请输入歌单名称" />
      </n-form-item>
      <n-form-item v-if="online" label="设为隐私歌单" label-placement="left">
        <n-switch v-model:value="formData.private" />
      </n-form-item>
      <n-form-item v-else label="歌单描述" path="description">
        <n-input
          v-model:value="formData.description"
          type="textarea"
          placeholder="请输入歌单描述（选填）"
          :autosize="{ minRows: 2, maxRows: 4 }"
        />
      </n-form-item>
    </n-form>
    <n-button class="create" type="primary" @click="toCreatePlaylist">
      新建
    </n-button>
  </div>
</template>

<script setup lang="ts">
// 照搬 SPlayer 的 CreatePlaylist.vue。Sonora 的歌单只有名称与描述，沿用其本地歌单的表单；
// 在线歌单沿用其在线表单的名称与隐私，与 SPlayer 不同的是没有“歌单类型”
// （视频歌单不做，共享歌单没有对应功能）。
import type { FormInst, FormRules } from "naive-ui";
import { ApiError, loginUrl, source } from "@/api/source";
import { useUserStore } from "@/stores";
import { textRule } from "@/utils/rules";
import { debounce } from "lodash-es";

const props = withDefaults(defineProps<{ online?: boolean }>(), {
  online: false,
});

const emit = defineEmits<{ close: [] }>();

// 表单类型
interface FormType {
  name: string;
  description: string;
  private: boolean;
}

const userStore = useUserStore();

const formRef = ref<FormInst | null>(null);
const formData = ref<FormType>({
  name: "",
  description: "",
  private: false,
});
const formRules: FormRules = { name: textRule };

// 新建歌单
const toCreatePlaylist = debounce(
  async (e: MouseEvent) => {
    e.preventDefault();
    try {
      await formRef.value?.validate();
    } catch {
      // 验证失败，不做处理
      return;
    }
    try {
      const name = formData.value.name.trim();
      await (props.online
        ? source.createOnlinePlaylist({
            name,
            private: formData.value.private,
          })
        : source.createPlaylist({
            name,
            description: formData.value.description.trim(),
          }));
    } catch (error) {
      if (error instanceof ApiError && error.status === 401) {
        location.assign(loginUrl);
        return;
      }
      window.$message.error(
        error instanceof ApiError ? error.message : "新建歌单失败，请重试",
      );
      return;
    }
    emit("close");
    window.$message.success("新建歌单成功");
    await (props.online
      ? userStore.loadOnlinePlaylists()
      : userStore.loadPlaylists());
  },
  300,
  { leading: true, trailing: false },
);
</script>

<style lang="scss" scoped>
.create-playlist {
  .n-form {
    margin-top: 12px;
  }
  .create {
    width: 100%;
  }
  .n-empty {
    padding: 40px 0;
  }
}
</style>
