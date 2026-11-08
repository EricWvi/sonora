import { useEventListener } from "@vueuse/core";
import { debounce } from "lodash-es";
import { usePlayerController } from "@/core/player/PlayerController";
import { useShortcutStore, useStatusStore } from "@/stores";
import type { ShortcutList } from "@/stores/shortcut";

/**
 * 页面内快捷键。照搬 SPlayer composables/useInit.ts 的键盘事件：输入框中不响应，
 * 按 store 中的组合键匹配后执行。与 SPlayer 网页版相同，全部快捷键只在页面内生效；
 * 去掉桌面歌词。
 */
export const useShortcut = () => {
  useEventListener(window, "keydown", keyDownEvent);
};

/** 输入焦点所在的元素不响应快捷键 */
const inInput = (event: KeyboardEvent): boolean => {
  const target = event.target as HTMLElement;
  const extendsDom = ["input", "textarea"];
  return extendsDom.includes(target.tagName.toLowerCase());
};

/** 事件是否为这个组合键（修饰键状态也要吻合） */
export const matchesShortcut = (
  event: KeyboardEvent,
  shortcut: string,
): boolean => {
  const parts = shortcut.split("+");
  const isCtrl = event.ctrlKey || event.metaKey;
  const hasCmdOrCtrl = parts.includes("CmdOrCtrl");
  const hasShift = parts.includes("Shift");
  const hasAlt = parts.includes("Alt");
  const mainKey = parts.find(
    (part) => part !== "CmdOrCtrl" && part !== "Shift" && part !== "Alt",
  );
  if (hasCmdOrCtrl !== isCtrl) return false;
  if (hasShift !== event.shiftKey) return false;
  if (hasAlt !== event.altKey) return false;
  if (
    !hasCmdOrCtrl &&
    !hasShift &&
    !hasAlt &&
    (isCtrl || event.shiftKey || event.altKey)
  )
    return false;
  return mainKey === event.code;
};

/** 事件命中的快捷键 */
export const matchedShortcuts = (
  event: KeyboardEvent,
  shortcutList: ShortcutList,
): (keyof ShortcutList)[] =>
  (Object.keys(shortcutList) as (keyof ShortcutList)[]).filter((key) =>
    matchesShortcut(event, shortcutList[key].shortcut),
  );

/**
 * 键盘事件。与 SPlayer 不同（Eric，2026-11-08）：是否阻止默认行为同步决定，命中与执行
 * 放在防抖里。SPlayer 把整个处理函数防抖，preventDefault 落后于浏览器的默认行为：在可
 * 滚动的页面上按空格，页面先滚动，快捷键才触发播放。无差别阻止默认行为还会禁掉 Tab 切
 * 焦点、裸空格滚页等，所以只拦命中快捷键的按键。
 */
const keyDownEvent = (event: KeyboardEvent) => {
  if (inInput(event)) return;
  if (!matchedShortcuts(event, useShortcutStore().shortcutList).length) return;
  event.preventDefault();
  event.stopPropagation();
  executeShortcut(event);
};

const executeShortcut = debounce((event: KeyboardEvent) => {
  const player = usePlayerController();
  const shortcutStore = useShortcutStore();
  const statusStore = useStatusStore();
  for (const shortcutKey of matchedShortcuts(
    event,
    shortcutStore.shortcutList,
  )) {
    const shortcut = shortcutStore.shortcutList[shortcutKey];
    console.log(shortcutKey, `快捷键触发: ${shortcut.name}`);
    switch (shortcutKey) {
      case "playOrPause":
        player.playOrPause();
        break;
      case "playPrev":
        player.nextOrPrev("prev");
        break;
      case "playNext":
        player.nextOrPrev("next");
        break;
      case "seekForward":
        player.seekBy(5000);
        break;
      case "seekBackward":
        player.seekBy(-5000);
        break;
      case "volumeUp":
        player.setVolume("up");
        break;
      case "volumeDown":
        player.setVolume("down");
        break;
      case "openPlayer":
        // 打开播放界面（任意界面）
        statusStore.showFullPlayer = true;
        break;
      case "closePlayer":
        // 关闭播放界面（仅在播放界面时）
        if (statusStore.showFullPlayer) {
          statusStore.showFullPlayer = false;
        }
        break;
      case "openPlayList":
        // 打开播放列表（任意界面）
        statusStore.playListShow = !statusStore.playListShow;
        break;
      default:
        break;
    }
  }
}, 100);
