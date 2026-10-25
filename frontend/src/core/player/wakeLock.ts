// 阻止系统息屏（设置“播放 → 阻止系统息屏”）。SPlayer 只在 Electron 中用 powerSaveBlocker
// 实现，网页版没有；Sonora 用 Screen Wake Lock API，播放时持有，暂停时释放。页面隐藏时
// 浏览器会自动释放，回到页面后重新申请。

let wanted = false;
let sentinel: WakeLockSentinel | null = null;
let requesting = false;
let listening = false;

const supported = () =>
  typeof navigator !== "undefined" && "wakeLock" in navigator;

async function acquire() {
  if (sentinel || requesting || document.visibilityState !== "visible") return;
  requesting = true;
  try {
    const lock = await navigator.wakeLock.request("screen");
    if (!wanted) {
      await lock.release();
      return;
    }
    sentinel = lock;
    lock.addEventListener("release", () => {
      if (sentinel === lock) sentinel = null;
    });
  } catch (error) {
    console.warn("阻止息屏失败", error);
  } finally {
    requesting = false;
  }
}

/**
 * 是否阻止息屏
 * @param on 设置开启且正在播放
 */
export function keepAwake(on: boolean) {
  if (!supported() || on === wanted) return;
  wanted = on;
  if (!listening) {
    listening = true;
    document.addEventListener("visibilitychange", () => {
      if (wanted) void acquire();
    });
  }
  if (on) {
    void acquire();
  } else if (sentinel) {
    const lock = sentinel;
    sentinel = null;
    void lock.release();
  }
}
