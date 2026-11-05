<!-- 登录网易云：照搬 SPlayer Login.vue 的结构与样式。与 SPlayer 不同：不做手机号，
所以没有标签页，二维码下面只有“Cookie 登录”；不做 UID 登录；登录结果保存在服务端。 -->
<template>
  <div class="login">
    <Logo :size="60" class="logo" />
    <LoginQRCode :pause="qrPause" @success="emit('close')" />
    <n-flex align="center" class="other">
      <n-button
        :focusable="false"
        size="small"
        quaternary
        round
        @click="cookieLogin"
      >
        Cookie 登录
      </n-button>
    </n-flex>
    <n-button
      :focusable="false"
      class="close"
      strong
      secondary
      round
      @click="emit('close')"
    >
      <template #icon>
        <SvgIcon name="WindowClose" />
      </template>
      取消
    </n-button>
  </div>
</template>

<script setup lang="ts">
import Logo from "@/components/Layout/Logo.vue";
import SvgIcon from "@/components/Global/SvgIcon.vue";
import LoginPaste from "./LoginPaste.vue";
import LoginQRCode from "./LoginQRCode.vue";

const emit = defineEmits<{ close: [] }>();

// Cookie 弹窗打开时暂停检查二维码
const qrPause = ref(false);

function cookieLogin(): void {
  qrPause.value = true;
  const modal = window.$modal.create({
    title: "Cookie 登录",
    preset: "card",
    transformOrigin: "center",
    style: { width: "400px" },
    content: () =>
      h(LoginPaste, {
        onClose: () => modal.destroy(),
        onSuccess: () => {
          modal.destroy();
          emit("close");
        },
      }),
    onClose: () => {
      qrPause.value = false;
      modal.destroy();
    },
  });
}
</script>

<style lang="scss" scoped>
.login {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-direction: column;
  .logo {
    width: 60px;
    height: 60px;
    margin: 20px auto 10px auto;
  }
  .other {
    margin: 0 0 20px;
    .n-button {
      width: 140px;
    }
  }
  .close {
    margin-bottom: 8px;
  }
}
</style>
