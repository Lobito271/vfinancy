export interface ApiError extends Error {
  status?: number;
  code?: string;
  details?: unknown;
}

export class ApiClientError extends Error implements ApiError {
  status?: number;
  code?: string;
  details?: unknown;

  constructor(message: string, opts: Partial<ApiError> = {}) {
    super(message);
    this.name = 'ApiClientError';
    this.status = opts.status;
    this.code = opts.code;
    this.details = opts.details;
  }
}

export interface RequestOptions extends Omit<RequestInit, 'body'> {
  body?: unknown;
  query?: Record<string, string | number | boolean | undefined | null>;
  timeoutMs?: number;
}

export interface ApiResponse<T> {
  data: T;
  status: number;
  headers: Headers;
}

export interface ApiClient {
  request<T>(path: string, opts?: RequestOptions): Promise<ApiResponse<T>>;
  get<T>(path: string, opts?: RequestOptions): Promise<T>;
  post<T>(path: string, body?: unknown, opts?: RequestOptions): Promise<T>;
  put<T>(path: string, body?: unknown, opts?: RequestOptions): Promise<T>;
  patch<T>(path: string, body?: unknown, opts?: RequestOptions): Promise<T>;
  delete<T>(path: string, opts?: RequestOptions): Promise<T>;
}

export function buildQuery(q: RequestOptions['query']): string {
  if (!q) return '';
  const params = new URLSearchParams();
  for (const [k, v] of Object.entries(q)) {
    if (v === undefined || v === null) continue;
    params.append(k, String(v));
  }
  const s = params.toString();
  return s ? `?${s}` : '';
}
