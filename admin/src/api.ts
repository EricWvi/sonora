// Client for the admin API (/admin/api/v1). Every failure becomes an
// ApiError; a network failure has status 0 and code "unreachable".

export const API = "/admin/api/v1";

export class ApiError extends Error {
  constructor(
    readonly status: number,
    readonly code: string,
    message: string,
  ) {
    super(message);
  }
}

// Called when a request finds the session gone, so the app can return to
// the sign-in page.
let onUnauthenticated: () => void = () => {};
export function handleUnauthenticated(fn: () => void) {
  onUnauthenticated = fn;
}

export async function toError(res: Response): Promise<ApiError> {
  let code = "http_" + res.status;
  let message = `请求失败（${res.status}）`;
  try {
    const body = await res.json();
    code = body.code ?? code;
    message = body.error ?? message;
  } catch {
    // Not a Sonora error body, e.g. a proxy page.
  }
  if (res.status === 502 || res.status === 503 || res.status === 504) {
    return new ApiError(0, "unreachable", "无法连接 Sonora 服务");
  }
  return new ApiError(res.status, code, message);
}

export function checkUnauthenticated(err: ApiError) {
  if (err.status === 401 && err.code === "unauthenticated") onUnauthenticated();
}

export async function request<T>(
  method: string,
  path: string,
  body?: unknown,
): Promise<T> {
  const init: RequestInit = { method, headers: {} };
  if (body instanceof Blob || typeof body === "string") {
    init.body = body;
  } else if (body !== undefined) {
    init.body = JSON.stringify(body);
    (init.headers as Record<string, string>)["Content-Type"] =
      "application/json";
  }
  let res: Response;
  try {
    res = await fetch(API + path, init);
  } catch {
    throw new ApiError(0, "unreachable", "无法连接 Sonora 服务");
  }
  if (!res.ok) {
    const err = await toError(res);
    checkUnauthenticated(err);
    throw err;
  }
  if (res.status === 204) return undefined as T;
  return res.json();
}

/** Thumbnail edge lengths the service serves besides the original. */
export type CoverSize = 96 | 300;

export const coverURL = (name: string, size?: CoverSize) =>
  `${API}/covers/${name}` + (size ? `?size=${size}` : "");
export const audioURL = (id: number) => `${API}/tracks/${id}/audio`;

// --- types ------------------------------------------------------------------

export interface Admin {
  id: number;
  username: string;
}

export interface Status {
  initialized: boolean;
  devMode: boolean;
  admin: Admin | null;
}

export interface Cover {
  name: string;
  /** Blurred placeholder (ThumbHash, base64); empty when unavailable. */
  thumbhash: string;
}

export interface Ref {
  id: number;
  name: string;
}

export interface Track {
  id: number;
  title: string;
  artists: Ref[];
  album: {
    id: number;
    title: string;
    artists: Ref[];
    cover: Cover | null;
  } | null;
  trackNo: number;
  discNo: number;
  year: number;
  durationMs: number;
  format: string;
  size: number;
  cover: Cover | null;
  hasLyrics: boolean;
  /** Empty when the track has no lyrics. */
  lyricsFormat: "lrc" | "ttml" | "";
  originalName: string;
  createdAt: string;
}

export interface TrackDetail extends Track {
  lyrics: string;
}

export interface TrackInput {
  title: string;
  artists: string[];
  album: string;
  /** Empty means the same as artists. */
  albumArtists: string[];
  trackNo: number;
  discNo: number;
  year: number;
}

export interface Album {
  id: number;
  title: string;
  artists: Ref[];
  year: number;
  cover: Cover | null;
  // An album that only holds singles; the player's album page hides it.
  single: boolean;
  // Written by the admin; the player's album page shows it.
  description: string;
  trackCount: number;
  durationMs: number;
  createdAt: string;
}

export interface Artist {
  id: number;
  name: string;
  image: Cover | null;
  trackCount: number;
  albumCount: number;
  createdAt: string;
}

export interface Page<T> {
  items: T[];
  total: number;
}

export interface Overview {
  tracks: number;
  albums: number;
  artists: number;
  missingTags: number;
  durationMs: number;
  mediaBytes: number;
  coverBytes: number;
}

export interface IngestResult {
  status: "created" | "duplicate";
  track: TrackDetail;
}

export function query(
  params: Record<string, string | number | boolean | undefined>,
) {
  const q = new URLSearchParams();
  for (const [k, v] of Object.entries(params)) {
    if (v !== undefined && v !== "" && v !== false) q.set(k, String(v));
  }
  const s = q.toString();
  return s ? "?" + s : "";
}

// --- formatting -------------------------------------------------------------

export function formatDuration(ms: number) {
  const total = Math.round(ms / 1000);
  const h = Math.floor(total / 3600);
  const m = Math.floor((total % 3600) / 60);
  const s = String(total % 60).padStart(2, "0");
  return h ? `${h}:${String(m).padStart(2, "0")}:${s}` : `${m}:${s}`;
}

export function formatBytes(n: number) {
  const units = ["B", "KB", "MB", "GB", "TB"];
  let i = 0;
  while (n >= 1024 && i < units.length - 1) {
    n /= 1024;
    i++;
  }
  return `${i ? n.toFixed(1) : n} ${units[i]}`;
}

export const artistNames = (artists: Ref[]) =>
  artists.map((a) => a.name).join(" / ");
