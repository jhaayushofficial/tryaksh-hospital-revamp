import { logger } from "./logger";

export const API_BASE: string =
  import.meta.env.VITE_API_BASE || "http://localhost:8080/api/v1";

/** The error envelope every endpoint returns. */
interface ErrorEnvelope {
  error?: { code?: string; message?: string };
  request_id?: string;
}

/**
 * A failed API call. `requestId` is the server's correlation id, which is what
 * a user should quote when reporting a problem.
 */
export class ApiError extends Error {
  readonly status: number;
  readonly code: string;
  readonly requestId?: string;

  constructor(message: string, status: number, code: string, requestId?: string) {
    super(message);
    this.name = "ApiError";
    this.status = status;
    this.code = code;
    this.requestId = requestId;
  }
}

export interface RequestOptions extends Omit<RequestInit, "body"> {
  body?: unknown;
  /** Bearer token to send. */
  token?: string | null;
}

/**
 * Single entry point for every call to the API: it attaches JSON headers and
 * the bearer token, turns a non-2xx response into an ApiError carrying the
 * server's message and request id, and logs both outcomes.
 */
export async function request<T>(path: string, options: RequestOptions = {}): Promise<T> {
  const { body, token, headers, ...rest } = options;
  const method = rest.method ?? (body === undefined ? "GET" : "POST");

  const requestHeaders = new Headers(headers);
  if (body !== undefined) requestHeaders.set("Content-Type", "application/json");
  if (token) requestHeaders.set("Authorization", `Bearer ${token}`);

  const startedAt = performance.now();

  let response: Response;
  try {
    response = await fetch(`${API_BASE}${path}`, {
      ...rest,
      method,
      headers: requestHeaders,
      body: body === undefined ? undefined : JSON.stringify(body),
    });
  } catch (cause) {
    // fetch only rejects when the request never completed — offline, DNS,
    // CORS. That is a different failure from a 4xx and is worth its own message.
    logger.error("network request failed", { method, path, detail: String(cause) });
    throw new ApiError(
      "Can't reach the clinic's server. Check your connection and try again.",
      0,
      "NETWORK_ERROR",
    );
  }

  const requestId = response.headers.get("X-Request-Id") ?? undefined;
  const durationMs = Math.round(performance.now() - startedAt);

  if (!response.ok) {
    const envelope = (await response.json().catch(() => ({}))) as ErrorEnvelope;
    const message = envelope.error?.message || `Request failed (${response.status})`;

    logger.warn("api request failed", {
      method,
      path,
      status: response.status,
      code: envelope.error?.code,
      requestId: envelope.request_id ?? requestId,
      durationMs,
    });

    throw new ApiError(
      message,
      response.status,
      envelope.error?.code ?? "UNKNOWN",
      envelope.request_id ?? requestId,
    );
  }

  logger.debug("api request ok", { method, path, status: response.status, durationMs });

  // 204 and empty bodies are valid successes (DELETE, logout).
  if (response.status === 204) return null as T;
  const text = await response.text();
  return (text ? JSON.parse(text) : null) as T;
}
