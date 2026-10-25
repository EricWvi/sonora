<template>
  <div class="update-playlist">
    <n-form
      ref="updateFormRef"
      :model="updateFormData"
      :rules="updateFormRules"
    >
      <n-form-item label="歌单名" path="name">
        <n-input
          v-model:value="updateFormData.name"
          placeholder="请输入歌单名"
        />
      </n-form-item>
      <n-form-item label="歌单描述" path="desc">
        <n-input
          v-model:value="updateFormData.desc"
          :autosize="{
            minRows: 3,
            maxRows: 6,
          }"
          :maxlength="800"
          placeholder="请输入歌单描述"
          type="textarea"
          show-count
          clearable
        />
      </n-form-item>
    </n-form>
    <n-button class="create" type="primary" @click="toUpdatePlaylist">
      编辑
    </n-button>
  </div>
</template>

<script setup lang="ts">
// 照搬 SPlayer 的 UpdatePlaylist.vue，去掉网易的歌单分类与“我喜欢的音乐”。
import type { FormInst, FormRules } from "naive-ui";
import { ApiError, loginUrl, source } from "@/api/source";
import { useUserStore } from "@/stores";
import { textRule } from "@/utils/rules";
import { debounce } from "lodash-es";

// 表单类型
interface UpdateFormType {
  name: string;
  desc?: string;
}

const props = defineProps<{
  id: number;
  data: { name: string; description?: string };
}>();

const emit = defineEmits<{ success: [] }>();

const userStore = useUserStore();

// 表单数据
const updateFormRef = ref<FormInst | null>(null);
const updateFormData = ref<UpdateFormType>({
  name: props.data.name,
  desc: props.data.description,
});
const updateFormRules: FormRules = { name: textRule };

// 更新歌单
const toUpdatePlaylist = debounce(
  async (e: MouseEvent) => {
    e.preventDefault();
    // 是否输入
    try {
      await updateFormRef.value?.validate();
    } catch {
      return;
    }
    try {
      await source.updatePlaylist(props.id, {
        name: updateFormData.value.name.trim(),
        description: (updateFormData.value.desc ?? "").trim(),
      });
    } catch (error) {
      if (error instanceof ApiError && error.status === 401) {
        location.assign(loginUrl);
        return;
      }
      window.$message.error(
        error instanceof ApiError ? error.message : "歌单编辑失败，请重试",
      );
      return;
    }
    emit("success");
    window.$message.success("歌单编辑成功");
    await userStore.loadPlaylists();
  },
  300,
  { leading: true, trailing: false },
);
</script>

<style lang="scss" scoped>
.update-playlist {
  .create {
    width: 100%;
  }
}
</style>
