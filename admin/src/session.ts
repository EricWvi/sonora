import { reactive } from "vue";
import { request, type ApiError, type Status } from "./api";

// What the server said about initialisation and the signed-in admin.
// loaded stays false until the first status request succeeds.
export const session = reactive<Status & { loaded: boolean }>({
  loaded: false,
  initialized: false,
  devMode: false,
  admin: null,
});

// The outcome of loading the status, for the app shell to show.
export const boot = reactive({ error: null as ApiError | null });

let pending: Promise<void> | null = null;

// ensureStatus loads the status once; concurrent callers share the request.
export function ensureStatus(): Promise<void> {
  if (session.loaded) return Promise.resolve();
  if (!pending) {
    boot.error = null;
    pending = request<Status>("GET", "/status")
      .then((st) => {
        Object.assign(session, st, { loaded: true });
      })
      .catch((err: ApiError) => {
        boot.error = err;
      })
      .finally(() => {
        pending = null;
      });
  }
  return pending;
}

export async function reloadStatus() {
  session.loaded = false;
  await ensureStatus();
}

export function signedIn(st: Status) {
  Object.assign(session, st, { loaded: true });
}

export function signedOut() {
  session.admin = null;
}
