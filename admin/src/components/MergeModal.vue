<script setup lang="ts">
import { ref, watch } from "vue";
import { NButton, NModal, NRadio, NRadioGroup, NSpace, NText } from "naive-ui";

// Picks which of the selected items the others merge into.
const props = defineProps<{
  show: boolean;
  title: string;
  items: { id: number; label: string }[];
  merge: (target: number, sources: number[]) => Promise<void>;
}>();
const emit = defineEmits<{ close: [] }>();

const target = ref<number | null>(null);
const busy = ref(false);
watch(
  () => props.show,
  (v) => v && (target.value = props.items[0]?.id ?? null),
);

async function confirm() {
  if (target.value == null) return;
  busy.value = true;
  try {
    await props.merge(
      target.value,
      props.items.map((i) => i.id).filter((id) => id !== target.value),
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
