import type { FormItemRule } from "naive-ui";

// 照搬 SPlayer utils/rules.ts 中用到的规则。

// 普通文本
export const textRule: FormItemRule = {
  required: true,
  message: "请填写必要信息",
  trigger: ["blur"],
};
