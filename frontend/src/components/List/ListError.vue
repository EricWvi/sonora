<!-- 列表加载失败。样式沿用 SPlayer 流媒体页面的空状态（n-empty 加主色按钮）。
内容已被删除（missing）时重试没有意义，按钮换成 404 页的“返回上一级”。
没有登录网易云（login）时同样如此，按钮换成“登录网易云”，登录成功后自动重新读取。
SPlayer 的这些页面只有登录后才能进入，没有对应的提示，是 Sonora 新增。 -->
<template>
  <n-empty :description="error" size="large" class="list-error">
    <template #extra>
      <n-button
        v-if="missing"
        type="primary"
        strong
        secondary
        @click="router.go(-1)"
      >
        返回上一级
      </n-button>
      <n-button
        v-else-if="login"
        type="primary"
        strong
        secondary
        @click="openOnlineLogin"
      >
        登录网易云
      </n-button>
      <n-button v-else type="primary" strong secondary @click="emit('retry')">
        <template #icon>
          <SvgIcon name="Refresh" />
        </template>
        重试
      </n-button>
    </template>
  </n-empty>
</template>

<script setup lang="ts">
import { useUserStore } from "@/stores/user";
import { openOnlineLogin } from "@/utils/modal";

const props = defineProps<{
  error: string;
  missing?: boolean;
  login?: boolean;
}>();
const emit = defineEmits<{ retry: [] }>();
const router = useRouter();
const userStore = useUserStore();

// 看状态对象而不是 onlineSignedIn：列表先发现登录失效时，账号状态可能还是 signedIn，
// 重新登录后这个值不变。两种登录方式都会换一个新的状态对象。
watch(
  () => userStore.netease,
  (netease) => {
    if (props.login && netease.state === "signedIn") emit("retry");
  },
);
</script>

<style lang="scss" scoped>
.list-error {
  margin-top: 60px;
  // 详情页给空状态留出头部的位置（main.scss 的列表页通用样式）；出错时不显示头部
  padding-top: 0 !important;
}
</style>
