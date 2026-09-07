/**
 * Browser-side logging.
 *
 * Every failure the user can see should also be recorded here with enough
 * context to match it to a backend record: the API returns an X-Request-Id on
 * every response, and `ApiError` carries it, so a support call quoting a
 * reference can be traced across both sides.
 *
 * In production only warnings and errors are emitted, and `report` is the
 * single seam to forward them to Sentry or similar.
 */

type Level = "debug" | "info" | "warn" | "error";

const LEVELS: Record<Level, number> = { debug: 10, info: 20, warn: 30, error: 40 };

const isProd = import.meta.env.PROD;
const threshold = isProd ? LEVELS.warn : LEVELS.debug;

export interface LogContext {
  [key: string]: unknown;
}

function emit(level: Level, message: string, context?: LogContext) {
  if (LEVELS[level] < threshold) return;

  const payload = { level, message, ...context };

  // console.error/warn keep the browser's stack grouping, which a single
  // console.log would lose.
  if (level === "error") console.error("[tryaksh]", payload);
  else if (level === "warn") console.warn("[tryaksh]", payload);
  else console.info("[tryaksh]", payload);
}

export const logger = {
  debug: (message: string, context?: LogContext) => emit("debug", message, context),
  info: (message: string, context?: LogContext) => emit("info", message, context),
  warn: (message: string, context?: LogContext) => emit("warn", message, context),
  error: (message: string, context?: LogContext) => emit("error", message, context),
};

/**
 * Records a caught error and returns the message to show the user. Call it at
 * the point where the failure becomes visible, so the log and the UI agree.
 */
export function report(message: string, error: unknown, context?: LogContext): string {
  const detail = error instanceof Error ? error.message : String(error);
  const requestId =
    error && typeof error === "object" && "requestId" in error
      ? (error as { requestId?: string }).requestId
      : undefined;

  logger.error(message, { ...context, detail, requestId });
  return detail || message;
}
