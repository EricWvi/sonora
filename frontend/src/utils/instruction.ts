import { useIntersectionObserver } from "@vueuse/core";
import { debounce, throttle } from "lodash-es";
import { DirectiveBinding } from "vue";

interface HTMLElementWithDebounce extends HTMLElement {
  _debouncedFn?: EventListener;
  _debounceValue?: EventListener;
}

// 防抖指令。与 SPlayer 不同，点击时调用最新绑定的函数：列表项按 key 复用元素，
// 删除或移动歌曲后，挂载时的函数仍带着旧的序号（SPlayer 的播放队列会因此播错歌）。
export const debounceDirective = {
  mounted(el: HTMLElementWithDebounce, binding: DirectiveBinding) {
    if (typeof binding.value !== "function") {
      console.warn("v-debounce directive expects a function as the value");
      return;
    }

    const delay = binding.arg ? parseInt(binding.arg) : 300;
    el._debounceValue = binding.value;
    const debouncedFn = debounce(
      (event: Event) => el._debounceValue?.(event),
      delay,
      {
        leading: true,
        trailing: false,
      },
    );

    el.addEventListener("click", debouncedFn);
    el._debouncedFn = debouncedFn;
  },
  updated(el: HTMLElementWithDebounce, binding: DirectiveBinding) {
    if (typeof binding.value === "function") el._debounceValue = binding.value;
  },
  beforeUnmount(el: HTMLElementWithDebounce) {
    if (el._debouncedFn) {
      el.removeEventListener("click", el._debouncedFn);
      delete el._debouncedFn;
      delete el._debounceValue;
    }
  },
};

interface HTMLElementWithThrottle extends HTMLElement {
  _throttledFn?: EventListener;
}

// 节流指令
export const throttleDirective = {
  mounted(el: HTMLElementWithThrottle, binding: DirectiveBinding) {
    if (typeof binding.value !== "function") {
      console.warn("v-throttle directive expects a function as the value");
      return;
    }

    const delay = binding.arg ? parseInt(binding.arg) : 300;
    const throttledFn = throttle(binding.value, delay, {
      leading: true,
      trailing: false,
    });

    el.addEventListener("click", throttledFn);
    el._throttledFn = throttledFn;
  },
  beforeUnmount(el: HTMLElementWithThrottle) {
    if (el._throttledFn) {
      el.removeEventListener("click", el._throttledFn);
      delete el._throttledFn;
    }
  },
};

// 元素可见
export const visibleDirective = {
  mounted(el: HTMLElement, binding: DirectiveBinding) {
    const { modifiers, value } = binding;

    const handleIntersection = (entries: IntersectionObserverEntry[]) => {
      entries.forEach((entry) => {
        const isVisible = entry.isIntersecting;
        el.classList.toggle("hidden", !isVisible);
        if (typeof value === "function") value(isVisible);
        if (modifiers.once && isVisible) stop();
      });
    };
    const { stop } = useIntersectionObserver(el, handleIntersection, {
      threshold: 0.1,
    });
  },
};
