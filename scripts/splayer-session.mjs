import { readFile } from "node:fs/promises";

const sessionFile = new URL(
  "../.local/credentials/splayer/only_session",
  import.meta.url,
);

// 每次从本机凭据文件读取，确保后续访问使用 Eric 最新的登录会话。
export async function prepareSPlayerSession(
  context,
  baseURL = "http://127.0.0.1:25884",
) {
  const url = new URL(baseURL);
  if (
    !["http:", "https:"].includes(url.protocol) ||
    !["localhost", "127.0.0.1", "[::1]"].includes(url.hostname)
  ) {
    throw new Error(
      "SPlayer reference sessions require a local HTTP(S) instance",
    );
  }

  let raw;
  try {
    raw = await readFile(sessionFile, "utf8");
  } catch {
    throw new Error(
      "SPlayer credentials are unavailable; ask Eric to update the local session file",
    );
  }
  const cookies = raw
    .trim()
    .split(";")
    .map((part) => {
      const separator = part.indexOf("=");
      if (separator < 1) return null;
      const name = part.slice(0, separator).trim();
      const value = part.slice(separator + 1).trim();
      return name && value ? { name, value, url: url.origin } : null;
    })
    .filter(Boolean);
  if (!cookies.some((cookie) => cookie.name === "MUSIC_U")) {
    throw new Error(
      "SPlayer credentials have no MUSIC_U cookie; ask Eric to update the local session file",
    );
  }

  await context.addCookies(cookies);
  // SPlayer 也会读取 localStorage 中的备用值，且只在指定实例的 origin 写入。
  await context.addInitScript(
    ({ cookies, origin }) => {
      if (location.origin !== origin) return;
      for (const { name, value } of cookies)
        localStorage.setItem(`cookie-${name}`, value);
    },
    { cookies, origin: url.origin },
  );
}
