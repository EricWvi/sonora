<script setup lang="ts">
import { ref, watch } from "vue";
import {
  NButton,
  NModal,
  NRadio,
  NRadioGroup,
  NSpace,
  NSwitch,
  NText,
} from "naive-ui";

// Picks which of the selected items the others merge into. An option,
// when given, is a switch that starts on and is passed to merge.
const props = defineProps<{
  show: boolean;
  title: string;
  items: { id: number; label: string }[];
  option?: { label: string; hint: string };
  merge: (target: number, sources: number[], option: boolean) => Promise<void>;
}>();
const emit = defineEmits<{ close: [] }>();

const target = ref<number | null>(null);
const optionOn = ref(true);
const busy = ref(false);
watch(
  () => props.show,
  (v) => {
    if (!v) return;
    target.value = props.items[0]?.id ?? null;
    optionOn.value = true;
  },
);

async function confirm() {
  if (target.value == null) return;
  busy.value = true;
  try {
    await props.merge(
      target.value,
      props.items.map((i) => i.id).filter((id) => id !== target.value),
      optionOn.value,
    );
    emit("close");
  } finally {
    busy.value = false;
  }
}
</script>

<template>
  <n-modal
    :show="show"
    preset="card"
    :title="title"
    style="width: 440px"
    @update:show="(v: boolean) => !v && emit('close')"
  >
    <n-text depth="3">保留：</n-text>
    <n-radio-group
      v-model:value="target"
      style="margin-top: 8px; display: block"
    >
      <n-space vertical>
        <n-radio
          v-for="it in items"
          :key="it.id"
          :value="it.id"
          :label="it.label"
        />
      </n-space>
    </n-radio-group>
    <n-space v-if="props.option" vertical :size="4" style="margin-top: 16px">
      <n-space align="center" :size="8">
        <n-switch v-model:value="optionOn" />
        <n-text>{{ props.option.label }}</n-text>
      </n-space>
      <n-text depth="3" style="font-size: 12px">{{ props.option.hint }}</n-text>
    </n-space>
    <template #footer>
      <n-space justify="end">
        <n-button @click="emit('close')">取消</n-button>
        <n-button type="primary" :loading="busy" @click="confirm"
          >合并</n-button
        >
      </n-space>
    </template>
  </n-modal>
</template>
