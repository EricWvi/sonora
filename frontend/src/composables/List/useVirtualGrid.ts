/** 网格中应渲染的行，以及上下留白 */
export interface GridWindow {
  start: number;
  end: number;
  padTop: number;
  padBottom: number;
}

/**
 * 计算视口附近应渲染的行
 * @param heights 已实测的行高，未实测的为 undefined
 * @param estimate 未实测行的估计高度
 * @param gap 行间距
 * @param rowCount 总行数
 * @param viewTop 视口顶部相对网格顶部的位置
 * @param viewBottom 视口底部相对网格顶部的位置
 * @param buffer 视口上下额外渲染的行数
 */
export const gridWindow = (
  heights: (number | undefined)[],
  estimate: number,
  gap: number,
  rowCount: number,
  viewTop: number,
  viewBottom: number,
  buffer: number,
): GridWindow => {
  if (rowCount <= 0) return { start: 0, end: -1, padTop: 0, padBottom: 0 };
  const height = (i: number) => heights[i] ?? estimate;
  const tops: number[] = [];
  let top = 0;
  for (let i = 0; i < rowCount; i++) {
    tops.push(top);
    top += height(i) + gap;
  }
  const total = top - gap;
  let start = tops.findIndex((t, i) => t + height(i) > viewTop);
  if (start < 0) start = rowCount - 1;
  let end = start;
  while (end + 1 < rowCount && tops[end + 1] < viewBottom) end++;
  start = Math.max(0, start - buffer);
  end = Math.min(rowCount - 1, end + buffer);
  return {
    start,
    end,
    padTop: tops[start],
    padBottom: total - (tops[end] + height(end)),
  };
};

export interface VirtualGridOptions<T> {
  /** 全部数据 */
  items: () => T[];
  /** 列的最小宽度，与原网格的 minmax 相同 */
  minWidth: number;
  /** 行与列的间距 */
  gap: number;
  /** 视口上下额外渲染的行数 */
  buffer?: number;
  /** 渲染到最后几行时调用，用于载入下一页 */
  onReachEnd?: () => void;
  /** 按列数限制显示的条数，例如首页栏目只显示补满的整行 */
  maxItems?: (columns: number) => number;
}

/**
 * 虚拟网格：只渲染所在滚动容器视口附近的行，上下用留白撑出整个网格的高度。
 * 用于专辑与艺术家网格（阶段 6）。页面照旧整体滚动，行高实测后缓存，
 * 列宽变化时重新实测。
 */
export const useVirtualGrid = <T>(options: VirtualGridOptions<T>) => {
  const { minWidth, gap, buffer = 2 } = options;

  /** 网格容器，模板中以 ref 绑定 */
  const container = ref<HTMLElement | null>(null);
  const { width } = useElementSize(container);

  const columns = computed(() =>
    Math.max(1, Math.floor((width.value + gap) / (minWidth + gap))),
  );
  // 实际显示的条目
  const visible = computed(() => {
    const all = options.items();
    return options.maxItems
      ? all.slice(0, options.maxItems(columns.value))
      : all;
  });
  const rowCount = computed(() =>
    Math.ceil(visible.value.length / columns.value),
  );

  // 已实测的行高
  const heights = shallowRef<(number | undefined)[]>([]);
  const estimate = computed(() => {
    const known = heights.value.filter((h): h is number => h !== undefined);
    if (known.length) return known.reduce((a, b) => a + b, 0) / known.length;
    return width.value / columns.value + 80;
  });

  // 视口相对网格顶部的位置
  const view = shallowRef({ top: 0, bottom: 0 });

  const range = computed(() =>
    gridWindow(
      heights.value,
      estimate.value,
      gap,
      rowCount.value,
      view.value.top,
      view.value.bottom,
      buffer,
    ),
  );

  /** 应渲染的行 */
  const rows = computed(() => {
    const items = visible.value;
    const c = columns.value;
    const out: { index: number; items: T[] }[] = [];
    for (let r = range.value.start; r <= range.value.end; r++) {
      out.push({ index: r, items: items.slice(r * c, (r + 1) * c) });
    }
    return out;
  });
  const padTop = computed(() => range.value.padTop);
  const padBottom = computed(() => range.value.padBottom);

  // 所在的滚动容器；页面被 KeepAlive 缓存时不跟随滚动
  const scroller = shallowRef<HTMLElement | null>(null);
  let active = true;
  let frame = 0;

  const update = () => {
    const el = scroller.value;
    if (!active || !container.value || !el) return;
    const offset =
      container.value.getBoundingClientRect().top -
      el.getBoundingClientRect().top;
    view.value = { top: -offset, bottom: -offset + el.clientHeight };
  };
  const schedule = () => {
    if (frame) return;
    frame = requestAnimationFrame(() => {
      frame = 0;
      update();
    });
  };

  // 实测已渲染的行；data-row 为行号
  const measure = () => {
    if (!container.value) return;
    const next = [...heights.value];
    let changed = false;
    for (const el of container.value.querySelectorAll<HTMLElement>(
      ":scope > [data-row]",
    )) {
      const h = el.offsetHeight;
      const i = Number(el.dataset.row);
      if (h && next[i] !== h) {
        next[i] = h;
        changed = true;
      }
    }
    if (changed) heights.value = next;
  };

  /** 清空已实测的行高，例如卡片样式改变后 */
  const reset = () => {
    heights.value = [];
  };

  const attach = (el: HTMLElement | null) => {
    scroller.value?.removeEventListener("scroll", schedule);
    let parent = el?.parentElement ?? null;
    while (parent && !/(auto|scroll)/.test(getComputedStyle(parent).overflowY))
      parent = parent.parentElement;
    scroller.value = parent;
    parent?.addEventListener("scroll", schedule, { passive: true });
    update();
  };

  watch(container, attach, { flush: "post" });
  watch(width, reset);
  watch(rows, measure, { flush: "post" });
  // 卡片高度变化（如字体载入）会改变容器高度
  useResizeObserver(container, measure);
  useResizeObserver(scroller, schedule);
  watch(range, (r) => {
    if (rowCount.value && r.end >= rowCount.value - 1 - buffer)
      options.onReachEnd?.();
  });

  onActivated(() => {
    active = true;
    schedule();
  });
  onDeactivated(() => {
    active = false;
  });
  onBeforeUnmount(() => {
    scroller.value?.removeEventListener("scroll", schedule);
    cancelAnimationFrame(frame);
  });

  return {
    container,
    columns,
    rows,
    padTop,
    padBottom,
    reset,
    items: visible,
  };
};
