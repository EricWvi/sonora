<!-- 二维码登录：照搬 SPlayer LoginQRCode.vue。与 SPlayer 不同：二维码内容与登录结果都由
Go 服务处理，播放端不接触 Cookie；服务不可用时显示说明并停止轮询，可以重试。 -->
<template>
  <div class="login-qrcode">
    <div class="qr-img">
      <div
        v-if="qrValue"
        :class="[
          'qr',
          { success: qrStatusCode === 802, error: qrStatusCode === 800 },
        ]"
      >
        <n-qr-code :value="qrValue" :size="160" error-correction-level="H" />
        <Transition name="fade" mode="out-in">
          <div v-if="loginName" class="login-data">
            <n-avatar
              :src="onlineImage(loginAvatar, 100)"
              fallback-src="/images/avatar.jpg"
              class="cover"
              round
            />
            <n-text>{{ loginName }}</n-text>
          </div>
        </Transition>
      </div>
      <n-skeleton v-else class="qr" />
    </div>
    <n-text class="tip" depth="3">{{ failure || qrTipText }}</n-text>
    <n-button
      v-if="failure"
      :focusable="false"
      size="small"
      secondary
      round
      class="retry"
      @click="getQrData"
    >
      重试
    </n-button>
  </div>
</template>

<script setup lang="ts">
import { ApiError, source } from "@/api/source";
import { useUserStore } from "@/stores/user";
import { onlineImage } from "@/utils/onlineImage";

const props = defineProps<{ pause?: boolean }>();
const emit = defineEmits<{ success: [] }>();

const userStore = useUserStore();

// 状态提示
const qrCodeTip = {
  800: "二维码过期，即将重试",
  801: "请打开云音乐 APP 扫码登录",
  802: "扫描成功，请在客户端确认登录",
  803: "登录成功",
} as const;

// 二维码数据
const qrValue = ref("");
const qrUnikey = ref("");
const qrStatusCode = ref<keyof typeof qrCodeTip>(801);
const failure = ref("");

const qrTipText = computed(
  () => qrCodeTip[qrStatusCode.value] || "遇到未知状态，请重试",
);

// 待确认数据
const loginName = ref("");
const loginAvatar = ref("");

function fail(error: unknown): void {
  pauseCheck();
  failure.value =
    error instanceof ApiError ? error.message : "请求失败，请重试";
}

// 获取二维码
async function getQrData(): Promise<void> {
  pauseCheck();
  failure.value = "";
  qrStatusCode.value = 801;
  loginName.value = "";
  loginAvatar.value = "";
  try {
    const { key, url } = await source.neteaseQRStart();
    qrValue.value = url;
    qrUnikey.value = key;
    resumeCheck();
  } catch (error) {
    fail(error);
  }
}

// 检查二维码状态
async function checkQrStatus(): Promise<void> {
  if (!qrUnikey.value || props.pause) return;
  try {
    const { code, nickname, avatarUrl } = await source.neteaseQRCheck(
      qrUnikey.value,
    );
    switch (code) {
      case 800:
        qrStatusCode.value = 800;
        await getQrData();
        break;
      case 801:
        qrStatusCode.value = 801;
        break;
      case 802:
        qrStatusCode.value = 802;
        loginName.value = nickname;
        loginAvatar.value = avatarUrl;
        break;
      case 803:
        qrStatusCode.value = 803;
        pauseCheck();
        await userStore.loadNetease();
        window.$message.success("登录成功");
        emit("success");
        break;
    }
  } catch (error) {
    fail(error);
  }
}

// 控制检查二维码
const { pause: pauseCheck, resume: resumeCheck } = useIntervalFn(
  checkQrStatus,
  1000,
  { immediate: false },
);

onMounted(getQrData);
onBeforeUnmount(pauseCheck);
</script>

<style lang="scss" scoped>
.login-qrcode {
  display: flex;
  flex-direction: column;
  align-items: center;
  .qr-img {
    display: flex;
    margin: 20px 0;
    width: 180px;
    height: 180px;
    border-radius: 12px;
    overflow: hidden;
    .qr {
      position: relative;
      display: flex;
      align-items: center;
      justify-content: center;
      width: 100%;
      height: 100%;
      .n-qr-code {
        padding: 0;
        height: 180px;
        width: 180px;
        min-height: 180px;
        min-width: 180px;
        transition:
          opacity 0.3s,
          filter 0.3s;
        :deep(canvas) {
          width: 100% !important;
          height: 100% !important;
        }
      }
      .login-data {
        position: absolute;
        display: flex;
        flex-direction: column;
        align-items: center;
        z-index: 1;
        .cover {
          width: 40px;
          height: 40px;
          border-radius: 50%;
          overflow: hidden;
          margin-bottom: 8px;
        }
      }
      &.success {
        .n-qr-code {
          opacity: 0.5;
          filter: blur(4px);
        }
      }
    }
  }
  .tip {
    margin-bottom: 12px;
  }
  .retry {
    margin-bottom: 12px;
  }
}
</style>
