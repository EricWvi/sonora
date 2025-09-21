import { QueryClient, type QueryFunction } from "@tanstack/react-query";

async function throwIfResNotOk(res: Response) {
  if (!res.ok) {
    const text = (await res.text()) || res.statusText;
    throw new Error(`${res.status}: ${text}`);
  }
}

export async function getRequest(
  url: string,
  retries = 3,
  baseTimeout = 2000,
  maxTimeout = 10000
): Promise<any> {
  const method = "GET";
  let errorMsg = `${method} ${url} failed after ${retries} attempts`;

  for (let attempt = 0; attempt < retries; attempt++) {
    // Exponential backoff timeout (2s → 4s → 8s)
    const timeout = Math.min(baseTimeout * 2 ** attempt, maxTimeout);
    // Setup abort controller for timeout
    const controller = new AbortController();
    const id = setTimeout(() => controller.abort(), timeout);

    try {
      const res = await fetch(url, {
        method,
        credentials: "include",
        signal: controller.signal,
      });

      if (!res.ok) {
        errorMsg = `${method} ${url} failed with status ${res.status}`;
        break; // fail fast on bad status
      }

      const data = await res.json();
      if (data.code !== 200) {
        errorMsg = `${method} ${url} failed with code ${data.code}`;
        break; // fail fast on bad status
      }

      return data.message; // success
    } catch (err) {
      if (attempt < retries - 1) {
        console.warn(`Retrying ${method} ${url} (attempt ${attempt + 2})...`);
        continue;
      }
    } finally {
      clearTimeout(id);
    }
  }

  // All retries failed → show toast and rethrow
  throw new Error(errorMsg);
}

export async function postRequest(
  url: string,
  body: unknown,
  retries = 3,
  baseTimeout = 3000,
  maxTimeout = 10000
): Promise<any> {
  const method = "POST";
  let errorMsg = `${method} ${url} failed after ${retries} attempts`;

  for (let attempt = 0; attempt < retries; attempt++) {
    // Exponential backoff timeout (3s → 6s → max 10s)
    const timeout = Math.min(baseTimeout * 2 ** attempt, maxTimeout);
    // Setup abort controller for timeout
    const controller = new AbortController();
    const id = setTimeout(() => controller.abort(), timeout);

    try {
      const res = await fetch(url, {
        method,
        headers: {
          "Content-Type": "application/json",
          "Idempotency-Key": crypto.randomUUID(),
        },
        body: JSON.stringify(body),
        credentials: "include",
        signal: controller.signal,
      });

      if (!res.ok) {
        errorMsg = `${method} ${url} failed with status ${res.status}`;
        break; // fail fast on bad status
      }

      const data = await res.json();
      if (data.code !== 200) {
        errorMsg = `${method} ${url} failed with code ${data.code}`;
        break; // fail fast on bad status
      }

      return data.message; // success
    } catch (err) {
      if (attempt < retries - 1) {
        console.warn(`Retrying ${method} ${url} (attempt ${attempt + 2})...`);
        continue;
      }
    } finally {
      clearTimeout(id);
    }
  }

  // All retries failed → show toast and rethrow
  throw new Error(errorMsg);
}

type UnauthorizedBehavior = "returnNull" | "throw";
export const getQueryFn: <T>(options: {
  on401: UnauthorizedBehavior;
}) => QueryFunction<T> =
  ({ on401: unauthorizedBehavior }) =>
  async ({ queryKey }) => {
    const res = await fetch(queryKey[0] as string, {
      credentials: "include",
    });

    if (unauthorizedBehavior === "returnNull" && res.status === 401) {
      return null;
    }

    await throwIfResNotOk(res);
    return await res.json();
  };

export const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      queryFn: getQueryFn({ on401: "throw" }),
      refetchInterval: false,
      refetchOnWindowFocus: false,
      staleTime: Infinity,
      retry: false,
    },
    mutations: {
      retry: false,
    },
  },
});
