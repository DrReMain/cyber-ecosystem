import { Code, ConnectError } from "@connectrpc/connect";
import { ErrorInfoSchema } from "@cyber-ecosystem/gen-ts/google/rpc/error_details_pb";

// AppError — the unified shape EVERY channel's thrown value normalizes into
// (request errors AND non-request errors). Consumers (copy projection,
// reporter, fallbacks) face this one shape only.
//
// Field contract (each backed by a live consumer):
//   reason   — UPPER_SNAKE identity when the backend spoke (undefined =
//              no identity: JS bugs, render throws, gateway-down)
//   network  — infra/gateway failure tier for fallback copy (two-tier:
//              network vs unknown; a finer taxonomy waits for real consumers)
//   message  — most readable raw text (wire message is always empty for
//              backend errors — human copy comes from reason via projection)
//   original — the raw thrown value: fidelity, future dedup key, reporter
//              attachment
export interface AppError {
  reason?: string;
  network: boolean;
  message: string;
  original: unknown;
}

// normalize — single entry, shape-dispatch (callers never know the protocol).
// Rules verified against live-wire observations (A2/BFF experiments):
//   ConnectError          → findDetails reason; code 14 without reason = down
//   plain object          → body itself (hey-api throws the body); reason at
//                           top level; status (attached by the http error
//                           interceptor) ≥500 without reason = down
//   Error / any other     → no identity (BFF-connect crossings lose identity
//                           by the server-fn boundary — that's B8's territory,
//                           not something to hack back from message text)
export function normalize(error: unknown): AppError {
  if (error instanceof ConnectError) {
    const reason = error.findDetails(ErrorInfoSchema)[0]?.reason;
    return {
      reason,
      network: reason === undefined && error.code === Code.Unavailable,
      message: error.rawMessage,
      original: error,
    };
  }

  if (typeof error === "object" && error !== null) {
    const { reason, status, message } = error as {
      reason?: unknown;
      status?: unknown;
      message?: unknown;
    };
    return {
      reason: typeof reason === "string" ? reason : undefined,
      network: typeof status === "number" && status >= 500 && typeof reason !== "string",
      message: typeof message === "string" ? message : "",
      original: error,
    };
  }

  return {
    network: false,
    message: error instanceof Error ? error.message : String(error),
    original: error,
  };
}
