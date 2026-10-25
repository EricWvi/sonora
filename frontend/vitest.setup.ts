// store 的持久化配置在定义时引用 localStorage；Node 中没有，给一个内存实现。
const items = new Map<string, string>();
globalThis.localStorage = {
  get length() {
    return items.size;
  },
  clear: () => items.clear(),
  getItem: (key) => items.get(key) ?? null,
  key: (index) => [...items.keys()][index] ?? null,
  removeItem: (key) => void items.delete(key),
  setItem: (key, value) => void items.set(key, String(value)),
};
