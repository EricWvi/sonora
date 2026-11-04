import type {
  DialogApi,
  LoadingBarApi,
  MessageApi,
  ModalApi,
  NotificationApi,
} from "naive-ui";

declare global {
  interface Window {
    // Naive UI 的全局工具，由 Provider 挂载
    $message: MessageApi;
    $dialog: DialogApi;
    $notification: NotificationApi;
    $loadingBar: LoadingBarApi;
    $modal: ModalApi;
  }

  interface ImportMetaEnv {
    // 构建时注入的版本号，见 Dockerfile 与 scripts/build.sh
    readonly VITE_APP_VERSION?: string;
  }
}

export {};
