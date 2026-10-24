<script setup lang="ts">
import { reactive, ref } from "vue";
import { useRouter } from "vue-router";
import {
  NAlert,
  NButton,
  NCard,
  NForm,
  NFormItem,
  NInput,
  type FormInst,
  type FormRules,
} from "naive-ui";
import { ApiError, request, type Status } from "../api";
import { reloadStatus, signedIn } from "../session";

const router = useRouter();
const formRef = ref<FormInst | null>(null);
const form = reactive({ token: "", username: "", password: "", confirm: "" });
const error = ref("");
const submitting = ref(false);

const rules: FormRules = {
  token: {
    required: true,
    message: "请输入容器日志中的初始化 token",
    trigger: "blur",
  },
  username: [
    { required: true, message: "请输入用户名", trigger: "blur" },
    { max: 32, message: "最多 32 个字符", trigger: "input" },
    { pattern: /^\S+$/, message: "不能包含空白", trigger: "input" },
  ],
  password: [
    { required: true, message: "请输入密码", trigger: "blur" },
    { min: 8, message: "至少 8 个字符", trigger: "blur" },
  ],
  confirm: {
    trigger: ["blur", "input"],
    validator: (_rule, value: string) =>
      value === form.password ? true : new Error("两次输入的密码不一致"),
  },
};

async function submit() {
  try {
    await formRef.value?.validate();
  } catch {
    return;
  }
  submitting.value = true;
  error.value = "";
  try {
    const st = await request<Status>("POST", "/setup", {
      token: form.token.trim(),
      username: form.username.trim(),
      password: form.password,
    });
    signedIn(st);
    router.replace("/");
  } catch (err) {
    const e = err as ApiError;
    if (e.code === "already_initialized") {
      await reloadStatus();
      router.replace("/login");
      return;
    }
    error.value = e.message;
  } finally {
    submitting.value = false;
  }
}
</script>

<template>
  <div class="center">
    <n-card title="初始化 Sonora" style="width: 400px">
      <n-form
        ref="formRef"
        :model="form"
        :rules="rules"
        @submit.prevent="submit"
      >
        <n-form-item label="初始化 token" path="token">
          <n-input
            v-model:value="form.token"
            placeholder="见容器日志"
            autofocus
          />
        </n-form-item>
        <n-form-item label="管理员用户名" path="username">
          <n-input
            v-model:value="form.username"
            :input-props="{ autocomplete: 'username' }"
          />
        </n-form-item>
        <n-form-item label="密码" path="password">
          <n-input
            v-model:value="form.password"
            type="password"
            show-password-on="click"
            :input-props="{ autocomplete: 'new-password' }"
          />
        </n-form-item>
        <n-form-item label="确认密码" path="confirm">
          <n-input
            v-model:value="form.confirm"
            type="password"
            show-password-on="click"
            :input-props="{ autocomplete: 'new-password' }"
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
        <n-button type="primary" attr-type="submit" block :loading="submitting">
          创建管理员
        </n-button>
      </n-form>
    </n-card>
  </div>
</template>
