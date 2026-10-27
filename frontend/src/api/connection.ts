// 与 Go 服务的连接状态（路线图的 Service Connection）：请求遇到 401 时通知会话失效，
// 服务从不可达恢复后通知订阅者重新读取（阶段 5 的异常状态）。
type Listener = () => void;

const unauthorizedListeners = new Set<Listener>();
const reconnectListeners = new Set<Listener>();
let unreachable = false;

export const connection = {
  /** 请求没有到达服务（网络错误，或代理在服务停止时返回的 5xx） */
  unreachable() {
    unreachable = true;
  },
  /** 服务有响应；此前不可达时通知恢复 */
  reached() {
    if (!unreachable) return;
    unreachable = false;
    for (const listener of reconnectListeners) listener();
  },
  /** 服务返回 401：会话已失效 */
  unauthorized() {
    for (const listener of unauthorizedListeners) listener();
  },
  /** 会话失效时调用 `listener`，返回取消订阅的函数 */
  onUnauthorized(listener: Listener) {
    unauthorizedListeners.add(listener);
    return () => unauthorizedListeners.delete(listener);
  },
  /** 服务从不可达恢复时调用 `listener`，返回取消订阅的函数 */
  onReconnect(listener: Listener) {
    reconnectListeners.add(listener);
    return () => reconnectListeners.delete(listener);
  },
};
