import { API, ApiError, checkUnauthenticated, type IngestResult } from "./api";

// uploadTrack sends one audio file, and optionally its sidecar lyrics,
// reporting upload progress from 0 to 1.
export function uploadTrack(
  file: File,
  lyrics: File | null,
  onProgress: (fraction: number) => void,
): Promise<IngestResult> {
  const form = new FormData();
  if (lyrics) form.append("lyrics", lyrics, lyrics.name);
  form.append("file", file, file.name);
  return new Promise((resolve, reject) => {
    const xhr = new XMLHttpRequest();
    xhr.open("POST", API + "/tracks");
    xhr.responseType = "json";
    xhr.upload.onprogress = (e) => {
      if (e.lengthComputable) onProgress(e.loaded / e.total);
    };
    xhr.onerror = () =>
      reject(new ApiError(0, "unreachable", "无法连接 Sonora 服务"));
    xhr.onload = () => {
      const body = xhr.response;
      if (xhr.status === 200 || xhr.status === 201) {
        resolve(body as IngestResult);
        return;
      }
      const err =
        xhr.status >= 502 && xhr.status <= 504
          ? new ApiError(0, "unreachable", "无法连接 Sonora 服务")
          : new ApiError(
              xhr.status,
              body?.code ?? "http_" + xhr.status,
              body?.error ?? `上传失败（${xhr.status}）`,
            );
      checkUnauthenticated(err);
      reject(err);
    };
    xhr.send(form);
  });
}

export function stem(name: string) {
  const dot = name.lastIndexOf(".");
  return (dot > 0 ? name.slice(0, dot) : name).toLowerCase();
}

export const isLyrics = (name: string) => /\.(lrc|ttml|yrc)$/i.test(name);
// When a song has several lyrics files, the highest rank wins: TTML, then
// YRC, then LRC.
export const lyricsRank = (name: string) =>
  /\.ttml$/i.test(name) ? 2 : /\.yrc$/i.test(name) ? 1 : 0;
