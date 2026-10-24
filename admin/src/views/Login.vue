<script setup lang="ts">
import { reactive, ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import { NAlert, NButton, NCard, NForm, NFormItem, NInput } from "naive-ui";
import { ApiError, request, type Status } from "../api";
import { session, signedIn } from "../session";

const route = useRoute();
const router = useRouter();
const form = reactive({
  username: session.devMode ? "admin" : "",
  password: session.devMode ? "admin" : "",
});
const error = ref("");
const submitting = ref(false);

async function submit() {
  if (!form.username || !form.password) {
    error.value = "请输入用户名和密码";
    return;
  }
  submitting.value = true;
  error.value = "";
  try {
    const st = await request<Status>("POST", "/login", form);
    signedIn(st);
    const next = typeof route.query.next === "string" ? route.query.next : "/";
    router.replace(next.startsWith("/") ? next : "/");
  } catch (err) {
    error.value = (err as ApiError).message;
  } finally {
    submitting.value = false;
  }
}
</script>

<template>
  <div class="center">
    <n-card title="Sonora 管理后台" style="width: 360px">
      <n-alert
        v-if="route.query.expired && !error"
        type="warning"
        :show-icon="false"
        style="margin-bottom: 16px"
      >
        登录已过期，请重新登录
      </n-alert>
      <n-alert
        v-if="session.devMode"
        type="info"
        :show-icon="false"
        style="margin-bottom: 16px"
      >
        开发模式：admin / admin
      </n-alert>
      <n-form @submit.prevent="submit">
        <n-form-item label="用户名">
          <n-input
            v-model:value="form.username"
            autofocus
            :input-props="{ autocomplete: 'username' }"
          />
        </n-form-item>
        <n-form-item label="密码">
          <n-input
            v-model:value="form.password"
            type="password"
            show-password-on="click"
            :input-props="{ autocomplete: 'current-password' }"
          />
        </n-form-item>
        <n-alert
          v-if="error"
          type="error"
          :show-icon="false"
          style="margin-bottom: 16px"
        >
          {{ error }}
        </n-alert>
        <n-button type="primary" attr-type="submit" block :loading="submitting"
          >登录</n-button
        >
      </n-form>
    </n-card>
  </div>
</template>
