// 单独成模块：fixture 要用到它，而 source.ts 在顶层 await 中加载 fixture，互相导入会卡住求值。
export class ApiError extends Error {
  /** HTTP 状态；0 表示服务不可达。 */
  readonly status: number;
  readonly code: string;
  constructor(status: number, message: string, code = "") {
    super(message);
    this.status = status;
    this.code = code;
  }
}
