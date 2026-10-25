// 播放端的依赖边界：不导入网易 API、Cookie 工具或 reference/ 下的原项目代码。
// ESLint 检查源码里的导入与网易接口地址（task lint:frontend），
// 构建插件检查实际进入构建的模块与产物（vite build）。
import { fileURLToPath } from "node:url";

const referenceDir = fileURLToPath(new URL("../reference/", import.meta.url));

// 正则里的斜杠写成 \x2F：ESLint 的选择器语法不支持正则中的 "/"
const forbiddenModules = [
  {
    pattern: /(^|\x2F)reference\x2F/,
    message: "reference/ 是原项目参照，不作为 Sonora 的运行依赖。",
  },
  {
    pattern: /netease/i,
    message: "播放端不接网易 API，数据只经 api/source.ts 读取 Sonora 服务。",
  },
  {
    pattern: /cookie/i,
    message: "播放端不读写 Cookie，会话 Cookie 由浏览器与服务端处理。",
  },
];

const forbiddenEndpoint = {
  pattern: /music\.163\.com|api\x2Fnetease|MUSIC_U/,
  message: "播放端不请求网易接口，也不处理网易登录 Cookie。",
};

/** @param {RegExp} pattern */
const selectorRegex = (pattern) => `/${pattern.source}/${pattern.flags}`;

/** ESLint 规则，用于 eslint.config.mjs */
export const boundaryRules = {
  "no-restricted-imports": [
    "error",
    {
      patterns: forbiddenModules.map(({ pattern, message }) => ({
        regex: pattern.source,
        caseSensitive: !pattern.flags.includes("i"),
        message,
      })),
    },
  ],
  "no-restricted-syntax": [
    "error",
    ...forbiddenModules.flatMap(({ pattern, message }) => [
      {
        selector: `ImportExpression > Literal[value=${selectorRegex(pattern)}]`,
        message,
      },
      {
        selector: `CallExpression[callee.object.type='MetaProperty'] Literal[value=${selectorRegex(pattern)}]`,
        message,
      },
    ]),
    {
      selector: `Literal[value=${selectorRegex(forbiddenEndpoint.pattern)}]`,
      message: forbiddenEndpoint.message,
    },
    {
      selector: `TemplateElement[value.raw=${selectorRegex(forbiddenEndpoint.pattern)}]`,
      message: forbiddenEndpoint.message,
    },
  ],
};

/**
 * 模块路径违反边界时返回原因
 * @param {string} id
 */
const moduleProblem = (id) => {
  const file = id.replace(/^\0/, "").replace(/\?.*$/, "");
  if (file.startsWith(referenceDir)) return forbiddenModules[0].message;
  // node_modules 与 src 下的路径只看网易与 Cookie 两项
  return forbiddenModules.slice(1).find(({ pattern }) => pattern.test(file))
    ?.message;
};

/**
 * 构建插件：模块加载时检查路径，构建结束时检查 SCSS 等引入的文件，
 * 生成产物时检查网易接口地址
 * @returns {import("vite").Plugin}
 */
export const boundaryPlugin = () => ({
  name: "sonora-boundaries",
  apply: "build",
  enforce: "pre",
  load(id) {
    const problem = moduleProblem(id);
    if (problem) this.error(`依赖边界检查失败：${id}: ${problem}`);
    return null;
  },
  buildEnd(error) {
    if (error) return;
    const problems = this.getWatchFiles().flatMap((file) => {
      const problem = moduleProblem(file);
      return problem ? [`${file}: ${problem}`] : [];
    });
    if (problems.length)
      this.error(`依赖边界检查失败：\n${problems.join("\n")}`);
  },
  generateBundle(_options, bundle) {
    for (const output of Object.values(bundle)) {
      if (
        output.type === "chunk" &&
        forbiddenEndpoint.pattern.test(output.code)
      ) {
        this.error(
          `依赖边界检查失败：${output.fileName}: ${forbiddenEndpoint.message}`,
        );
      }
    }
  },
});
