<template>
  <n-flex
    :wrap="false"
    align="center"
    style="width: 460px; justify-content: flex-end"
  >
    <!-- 本地/页面内快捷键 -->
    <n-popover trigger="focus">
      <template #trigger>
        <n-input
          :value="shortcutItem.shortcut"
          placeholder="未设置"
          readonly
          class="shortcut-input"
          @focus="onFocus"
          @blur="onBlur"
          @keydown.stop="onKeyDown"
          @keyup="keyHandled = ''"
        />
      </template>
      <n-text>正在设置快捷键，按 Backspace 删除快捷键</n-text>
    </n-popover>
  </n-flex>
</template>

<script setup lang="ts">
// 照搬 SPlayer components/Setting/components/ShortcutRecorder.vue 的页面内快捷键部分，
// 去掉全局快捷键（注册、占用检查）。
import { useShortcutStore } from "@/stores";
import type { ShortcutList } from "@/stores/shortcut";
import { some } from "lodash-es";
import { ref, computed } from "vue";

const props = defineProps<{
  shortcutKey: keyof ShortcutList;
}>();

const shortcutStore = useShortcutStore();
const shortcutItem = computed(
  () => shortcutStore.shortcutList[props.shortcutKey],
);

// 选中状态
const isFocus = ref(false);

// 按键标志位
const keyHandled = ref<string>("");

// 获取按下的快捷键
const getShortcut = (e: KeyboardEvent): string => {
  const allowedCodes = [
    // 字母 a-z
    "KeyA",
    "KeyB",
    "KeyC",
    "KeyD",
    "KeyE",
    "KeyF",
    "KeyG",
    "KeyH",
    "KeyI",
    "KeyJ",
    "KeyK",
    "KeyL",
    "KeyM",
    "KeyN",
    "KeyO",
    "KeyP",
    "KeyQ",
    "KeyR",
    "KeyS",
    "KeyT",
    "KeyU",
    "KeyV",
    "KeyW",
    "KeyX",
    "KeyY",
    "KeyZ",
    // 数字 0-9
    "Digit0",
    "Digit1",
    "Digit2",
    "Digit3",
    "Digit4",
    "Digit5",
    "Digit6",
    "Digit7",
    "Digit8",
    "Digit9",
    "Numpad0",
    "Numpad1",
    "Numpad2",
    "Numpad3",
    "Numpad4",
    "Numpad5",
    "Numpad6",
    "Numpad7",
    "Numpad8",
    "Numpad9",
    // 功能键
    "Space",
    "ArrowLeft",
    "ArrowUp",
    "ArrowRight",
    "ArrowDown",
    "Escape",
    // Funtion keys
    "F1",
    "F2",
    "F3",
    "F4",
    "F5",
    "F6",
    "F7",
    "F8",
    "F9",
    "F10",
    "F11",
    "F12",
  ];
  if (!allowedCodes.includes(e.code)) return "";
  return e.code;
};

const onFocus = () => {
  isFocus.value = true;
};

const onBlur = () => {
  isFocus.value = false;
};

// 快捷键是否重复
const isRepeat = (shortcut: string): boolean => {
  return some(
    Object.values(shortcutStore.shortcutList),
    (item) => item.shortcut === shortcut,
  );
};

const changeShortcut = (shortcut: string) => {
  shortcutStore.shortcutList[props.shortcutKey].shortcut = shortcut;
};

const isSameShortcut = (shortcut: string) => {
  return shortcutStore.shortcutList[props.shortcutKey].shortcut === shortcut;
};

const onKeyDown = (e: KeyboardEvent) => {
  e.preventDefault();
  e.stopPropagation();
  if (!isFocus.value) return;

  if (e.code === keyHandled.value) return;
  keyHandled.value = e.code;

  const blur = () => {
    keyHandled.value = "";
    const target = e.target;
    if (target instanceof HTMLElement) target.blur();
  };

  if (e.code === "Backspace") {
    changeShortcut("");
    window.$message.success("快捷键已删除");
    blur();
    return;
  }

  const isCtrl = e.ctrlKey || e.metaKey;
  const isShift = e.shiftKey;
  const isAlt = e.altKey;

  const keyCode = getShortcut(e);
  if (!keyCode) return;

  const shortcut = [
    isCtrl && "CmdOrCtrl",
    isShift && "Shift",
    isAlt && "Alt",
    keyCode,
  ]
    .filter(Boolean)
    .join("+");

  if (isSameShortcut(shortcut)) {
    window.$message.info("快捷键相同");
    blur();
    return;
  }

  if (isRepeat(shortcut)) {
    window.$message.warning("快捷键设置冲突");
    blur();
    return;
  }

  changeShortcut(shortcut);
  window.$message.success("快捷键设置成功");

  blur();
};
</script>

<style scoped lang="scss">
.shortcut-input {
  flex: 1;
  text-align: center;
  :deep(input) {
    text-align: center;
  }
}
.n-flex {
  gap: 12px;
}
</style>
