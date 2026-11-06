// 导入在线专辑（Eric，2026-11-06；SPlayer 没有这个功能）。服务端在后台逐首下载入库，
// 这里只负责确认、显示进度与结果：进度用 $message 的加载提示，确认与结果用现有的对话框，
// 不新增视觉样式。刷新页面后，用户载入时接着显示未完成的导入（userStore.load）。
import type { MessageApi, MessageReactive } from "naive-ui";
import { ApiError, loginUrl, source } from "@/api/source";
import type { ImportJob } from "@/api/types";

/** 还没完成的导入，按任务 ID */
const watching = new Map<number, ImportJob>();
let timer: ReturnType<typeof setTimeout> | undefined;
const POLL_MS = 1500;
// 消息照搬 SPlayer 只显示一条（Provider 的 max 为 1），所以所有导入共用一条进度提示。
// 新提示会直接挤掉旧的，Naive UI 不通知被挤掉的那条，所以给 $message 包一层：
// 别处每发一条提示，进度就让开，等那条显示完（默认 3 秒）再出现。
let progress: MessageReactive | null = null;
let hiddenUntil = 0;
const MESSAGE_MS = 3000;
let creatingProgress = false;
let wrappedApi: MessageApi | null = null;

const yieldToOtherMessages = () => {
  if (wrappedApi && window.$message === wrappedApi) return;
  const api = window.$message;
  wrappedApi = new Proxy(api, {
    get(target, prop, receiver) {
      const value = Reflect.get(target, prop, receiver);
      if (typeof value !== "function" || prop === "destroyAll") return value;
      return (...args: unknown[]) => {
        if (!creatingProgress && progress) {
          const options = args[1] as { duration?: number } | undefined;
          progress = null;
          hiddenUntil = Date.now() + (options?.duration || MESSAGE_MS);
        }
        return (value as (...a: unknown[]) => unknown).apply(target, args);
      };
    },
  });
  window.$message = wrappedApi;
};

const lines = (text: string[]) => () =>
  h(
    "div",
    text.map((line) => h("p", { style: "margin: 0 0 6px" }, line)),
  );

const progressText = () => {
  const jobs = [...watching.values()];
  const job = jobs.find((j) => j.state === "running") ?? jobs[0];
  const text =
    job.state === "queued"
      ? `等待导入《${job.title}》`
      : `正在导入《${job.title}》 ${job.done}/${job.total}`;
  return jobs.length > 1 ? `${text}，另有 ${jobs.length - 1} 张等待` : text;
};

const renderProgress = () => {
  if (!watching.size) {
    const msg = progress;
    progress = null;
    msg?.destroy();
    return;
  }
  if (progress) {
    progress.content = progressText();
    return;
  }
  if (Date.now() < hiddenUntil) return;
  yieldToOtherMessages();
  creatingProgress = true;
  try {
    progress = window.$message.loading(progressText(), { duration: 0 });
  } finally {
    creatingProgress = false;
  }
};

const summary = (job: ImportJob) =>
  [
    `新增 ${job.created} 首`,
    job.moved && `移入 ${job.moved} 首`,
    job.existing && `已在曲库 ${job.existing} 首`,
    job.unplayable && `跳过 ${job.unplayable} 首无法播放的歌曲`,
    job.failed.length && `${job.failed.length} 首没有导入`,
  ]
    .filter(Boolean)
    .join("，");

const finish = (job: ImportJob) => {
  const failures = job.failed.map((f) => `${f.title}：${f.reason}`);
  if (job.state === "failed") {
    window.$dialog.error({
      title: "导入中止",
      content: lines([
        `《${job.title}》导入中止：${job.error}`,
        `已处理 ${job.done}/${job.total} 首：${summary(job)}。已导入的歌曲会保留，稍后再导入一次即可补齐。`,
        ...failures,
      ]),
      positiveText: "知道了",
    });
  } else if (failures.length) {
    window.$dialog.warning({
      title: "部分歌曲没有导入",
      content: lines([`《${job.title}》：${summary(job)}。`, ...failures]),
      positiveText: "知道了",
    });
  } else {
    window.$message.success(`《${job.title}》导入完成：${summary(job)}`);
  }
};

const showProgress = (job: ImportJob) => {
  watching.set(job.id, job);
  renderProgress();
};

const stopWatching = () => {
  watching.clear();
  renderProgress();
};

const poll = async () => {
  timer = undefined;
  let jobs: ImportJob[];
  try {
    jobs = await source.importJobs();
  } catch (error) {
    if (
      error instanceof ApiError &&
      (error.status === 401 || error.status === 403)
    ) {
      stopWatching();
      return;
    }
    // 服务暂时不可达：保留提示，稍后再试
    timer = setTimeout(poll, POLL_MS * 4);
    return;
  }
  // 先撤下进度，结果提示才不会被它挤掉
  const finished: ImportJob[] = [];
  for (const id of watching.keys()) {
    const job = jobs.find((j) => j.id === id);
    if (!job) {
      // 服务重启后任务不再存在
      watching.delete(id);
    } else if (job.state === "done" || job.state === "failed") {
      watching.delete(id);
      finished.push(job);
    } else {
      watching.set(id, job);
    }
  }
  if (finished.length && progress) {
    const msg = progress;
    progress = null;
    msg.destroy();
    hiddenUntil = Date.now() + MESSAGE_MS;
  }
  finished.forEach(finish);
  renderProgress();
  if (watching.size) timer = setTimeout(poll, POLL_MS);
};

const schedule = () => {
  if (watching.size && !timer) timer = setTimeout(poll, POLL_MS);
};

/** 用户载入后接着显示还没完成的导入（例如刷新了页面）。 */
export const resumeAlbumImports = async () => {
  try {
    const jobs = await source.importJobs();
    for (const job of jobs)
      if (job.state === "queued" || job.state === "running") showProgress(job);
    schedule();
  } catch (error) {
    console.error("❌ 读取导入任务失败:", error);
  }
};

const reportError = (error: unknown, fallback: string) => {
  if (error instanceof ApiError && error.status === 401) {
    location.assign(loginUrl);
    return;
  }
  window.$message.error(error instanceof ApiError ? error.message : fallback);
  console.error("❌", fallback, error);
};

/**
 * 导入在线专辑：先预览数量并确认，再交给服务端在后台导入。
 * @param id 在线专辑 ID
 */
export const importAlbum = async (id: number) => {
  const loading = window.$message.loading("正在读取专辑", { duration: 0 });
  let preview;
  try {
    preview = await source.previewAlbumImport(id);
  } catch (error) {
    reportError(error, "读取专辑失败，请重试");
    return;
  } finally {
    loading.destroy();
  }
  const { title, total, toImport, inLibrary, moving, unplayable } = preview;
  if (toImport === 0 && moving === 0) {
    window.$message.info(
      total === 0
        ? `网易云没有返回《${title}》的歌曲，请稍后重试`
        : unplayable
          ? `《${title}》没有可以导入的歌曲：已在曲库 ${inLibrary} 首，无法播放 ${unplayable} 首`
          : `《${title}》的歌曲都已在曲库中`,
    );
    return;
  }
  const counts = [
    `将导入 ${toImport} 首`,
    inLibrary &&
      `已在曲库 ${inLibrary} 首${moving ? `（其中 ${moving} 首会移入这张专辑）` : ""}`,
    unplayable && `${unplayable} 首无法播放，会跳过`,
  ].filter(Boolean);
  const text = [
    `《${title}》共 ${total} 首：${counts.join("，")}。`,
    preview.target
      ? `歌曲会并入曲库中的专辑《${preview.target.title}》。`
      : "会在曲库中新建这张专辑。",
  ];
  if (preview.sameTitle)
    text.push(
      `曲库中另有 ${preview.sameTitle} 张同名专辑，导入后不会自动合并，需要时在管理后台合并。`,
    );
  text.push("文件由服务器下载，每首歌间隔 5 秒，导入期间可以继续使用。");
  window.$dialog.info({
    title: "导入专辑到曲库",
    content: lines(text),
    positiveText: "导入",
    negativeText: "取消",
    onPositiveClick: async () => {
      try {
        const job = await source.importOnlineAlbum(id);
        // 刚点了导入，进度马上出现，不等“正在读取专辑”的提示
        hiddenUntil = 0;
        showProgress(job);
        schedule();
      } catch (error) {
        reportError(error, "导入失败，请重试");
      }
    },
  });
};
