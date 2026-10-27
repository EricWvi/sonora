import { describe, expect, it, vi } from "vitest";
import { connection } from "./connection";

describe("connection", () => {
  it("只在不可达之后的第一次响应时通知恢复", () => {
    const reconnected = vi.fn();
    const off = connection.onReconnect(reconnected);
    connection.reached();
    expect(reconnected).not.toHaveBeenCalled();
    connection.unreachable();
    connection.unreachable();
    connection.reached();
    connection.reached();
    expect(reconnected).toHaveBeenCalledTimes(1);
    off();
    connection.unreachable();
    connection.reached();
    expect(reconnected).toHaveBeenCalledTimes(1);
  });

  it("401 时通知会话失效", () => {
    const unauthorized = vi.fn();
    const off = connection.onUnauthorized(unauthorized);
    connection.unauthorized();
    expect(unauthorized).toHaveBeenCalledTimes(1);
    off();
    connection.unauthorized();
    expect(unauthorized).toHaveBeenCalledTimes(1);
  });
});
