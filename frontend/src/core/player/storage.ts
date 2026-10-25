import type { QueueEntry } from "./queue";

// 播放队列与播放位置保存在浏览器中（阶段 2 的“播放行为”）。SPlayer 把播放列表存在
// IndexedDB（localforage），当前序号与位置存在 localStorage 的 status store；Sonora 都放进
// IndexedDB，不引入 localforage。队列只在变化时写入，位置写得频繁，分开保存。

/** 保存的队列 */
export interface SavedQueue {
  playList: QueueEntry[];
  /** 随机播放前的顺序；不在随机模式时为空 */
  originalPlayList: QueueEntry[];
  playIndex: number;
  /** 列表来源，见 music store 的 playPlaylistId */
  playPlaylistId: string;
}

/** 保存的播放位置，属于某个条目 */
export interface SavedPosition {
  entryId: number;
  positionMs: number;
}

export interface PlayerStorage {
  load(): Promise<{ queue: SavedQueue; position: SavedPosition | null } | null>;
  saveQueue(queue: SavedQueue): Promise<void>;
  savePosition(position: SavedPosition): Promise<void>;
  clear(): Promise<void>;
}

const DB_NAME = "sonora";
const STORE = "player";
const QUEUE_KEY = "queue";
const POSITION_KEY = "position";

let dbPromise: Promise<IDBDatabase> | null = null;

function openDb(): Promise<IDBDatabase> {
  dbPromise ??= new Promise((resolve, reject) => {
    const req = indexedDB.open(DB_NAME, 1);
    req.onupgradeneeded = () => req.result.createObjectStore(STORE);
    req.onsuccess = () => resolve(req.result);
    req.onerror = () => reject(req.error);
  });
  return dbPromise;
}

async function transact<T>(
  mode: IDBTransactionMode,
  run: (store: IDBObjectStore) => IDBRequest<T> | void,
): Promise<T | undefined> {
  const db = await openDb();
  return new Promise((resolve, reject) => {
    const tx = db.transaction(STORE, mode);
    const req = run(tx.objectStore(STORE));
    tx.oncomplete = () => resolve(req ? req.result : undefined);
    tx.onerror = () => reject(tx.error);
    tx.onabort = () => reject(tx.error);
  });
}

/** 保存在 IndexedDB 中 */
export const browserStorage: PlayerStorage = {
  async load() {
    const [queue, position] = await Promise.all([
      transact<SavedQueue>("readonly", (s) => s.get(QUEUE_KEY)),
      transact<SavedPosition>("readonly", (s) => s.get(POSITION_KEY)),
    ]);
    if (!queue || !Array.isArray(queue.playList)) return null;
    return { queue, position: position ?? null };
  },
  async saveQueue(queue) {
    // 歌曲对象可能是响应式代理，先转成普通对象
    await transact("readwrite", (s) => {
      s.put(JSON.parse(JSON.stringify(queue)), QUEUE_KEY);
    });
  },
  async savePosition(position) {
    await transact("readwrite", (s) => {
      s.put({ ...position }, POSITION_KEY);
    });
  },
  async clear() {
    await transact("readwrite", (s) => {
      s.delete(QUEUE_KEY);
      s.delete(POSITION_KEY);
    });
  },
};

/** 只保存在内存中，用于测试；fixture 用它，不改动浏览器中保存的队列 */
export function memoryStorage(): PlayerStorage & {
  queue: SavedQueue | null;
  position: SavedPosition | null;
} {
  return {
    queue: null,
    position: null,
    async load() {
      if (!this.queue) return null;
      return structuredClone({ queue: this.queue, position: this.position });
    },
    async saveQueue(queue) {
      this.queue = JSON.parse(JSON.stringify(queue));
    },
    async savePosition(position) {
      this.position = { ...position };
    },
    async clear() {
      this.queue = null;
      this.position = null;
    },
  };
}
