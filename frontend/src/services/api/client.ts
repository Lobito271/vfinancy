import { ApiClientError, buildQuery, type ApiClient, type ApiResponse, type RequestOptions } from './types';

export interface ClientConfig {
  baseURL: string;
  defaultHeaders?: Record<string, string>;
  getAuthToken?: () => string | null;
  timeoutMs?: number;
}

async function performRequest<T>(
  config: Required<Pick<ClientConfig, 'baseURL' | 'timeoutMs'>>,
  authToken: string | null,
  defaultHeaders: Record<string, string>,
  path: string,
  opts: RequestOptions,
): Promise<ApiResponse<T>> {
  const url = `${config.baseURL}${path}${buildQuery(opts.query)}`;
  const headers = new Headers({
    'Content-Type': 'application/json',
    Accept: 'application/json',
    ...defaultHeaders,
    ...(opts.headers as Record<string, string> | undefined),
  });
  if (authToken) headers.set('Authorization', `Bearer ${authToken}`);

  const controller = new AbortController();
  const timer = setTimeout(() => controller.abort(), opts.timeoutMs ?? config.timeoutMs);

  try {
    const res = await fetch(url, {
      ...opts,
      headers,
      body: opts.body !== undefined ? JSON.stringify(opts.body) : undefined,
      signal: controller.signal,
    });
    const text = await res.text();
    const data = text ? (JSON.parse(text) as T) : (undefined as T);
    if (!res.ok) {
      throw new ApiClientError(`Request failed: ${res.status}`, {
        status: res.status,
        code: res.statusText,
        details: data,
      });
    }
    return { data, status: res.status, headers: res.headers };
  } catch (err) {
    if (err instanceof ApiClientError) throw err;
    throw new ApiClientError(
      err instanceof Error ? err.message : 'Network error',
      { code: 'NETWORK_ERROR', details: err },
    );
  } finally {
    clearTimeout(timer);
  }
}

export function createApiClient(config: ClientConfig): ApiClient {
  const { baseURL, defaultHeaders = {}, getAuthToken, timeoutMs = 30_000 } = config;
  const finalConfig = { baseURL, timeoutMs };

  return {
    request: <T,>(path: string, opts: RequestOptions = {}) =>
      performRequest<T>(finalConfig, getAuthToken?.() ?? null, defaultHeaders, path, opts),
    get: <T,>(path: string, opts?: RequestOptions) =>
      performRequest<T>(finalConfig, getAuthToken?.() ?? null, defaultHeaders, path, { ...opts, method: 'GET' }).then((r) => r.data),
    post: <T,>(path: string, body?: unknown, opts?: RequestOptions) =>
      performRequest<T>(finalConfig, getAuthToken?.() ?? null, defaultHeaders, path, { ...opts, method: 'POST', body }).then((r) => r.data),
    put: <T,>(path: string, body?: unknown, opts?: RequestOptions) =>
      performRequest<T>(finalConfig, getAuthToken?.() ?? null, defaultHeaders, path, { ...opts, method: 'PUT', body }).then((r) => r.data),
    patch: <T,>(path: string, body?: unknown, opts?: RequestOptions) =>
      performRequest<T>(finalConfig, getAuthToken?.() ?? null, defaultHeaders, path, { ...opts, method: 'PATCH', body }).then((r) => r.data),
    delete: <T,>(path: string, opts?: RequestOptions) =>
      performRequest<T>(finalConfig, getAuthToken?.() ?? null, defaultHeaders, path, { ...opts, method: 'DELETE' }).then((r) => r.data),
  };
}

export const apiClient: ApiClient = createApiClient({
  baseURL: '/api',
  getAuthToken: () => null,
});
