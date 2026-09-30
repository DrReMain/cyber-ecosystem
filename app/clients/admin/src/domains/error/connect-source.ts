import { Code, ConnectError } from "@connectrpc/connect";
import { ErrorInfoSchema } from "@cyber-ecosystem/gen-connect-ts/google/rpc/error_details_pb";
import { type ErrorKind, makeEnvelope, type SourceAdapter } from "@cyber-ecosystem/shared-error";

const CODE_KIND: Partial<Record<Code, { kind: ErrorKind; http: number }>> = {
  [Code.DeadlineExceeded]: { kind: "network", http: 504 },
  [Code.Unavailable]: { kind: "network", http: 503 },
  [Code.Unknown]: { kind: "server", http: 500 },
  [Code.Aborted]: { kind: "server", http: 409 },
  [Code.Unimplemented]: { kind: "server", http: 501 },
  [Code.Internal]: { kind: "server", http: 500 },
  [Code.DataLoss]: { kind: "server", http: 500 },
  [Code.Canceled]: { kind: "canceled", http: 499 },
  [Code.InvalidArgument]: { kind: "business", http: 400 },
  [Code.NotFound]: { kind: "business", http: 404 },
  [Code.AlreadyExists]: { kind: "business", http: 409 },
  [Code.ResourceExhausted]: { kind: "business", http: 429 },
  [Code.FailedPrecondition]: { kind: "business", http: 400 },
  [Code.OutOfRange]: { kind: "business", http: 400 },
  [Code.Unauthenticated]: { kind: "auth", http: 401 },
  [Code.PermissionDenied]: { kind: "forbidden", http: 403 },
};

function toWireWord(enumName: string): string {
  return enumName
    .replace(/([a-z0-9])([A-Z])/g, "$1_$2")
    .replace(/-/g, "_")
    .toLowerCase();
}

export const connectSource: SourceAdapter = {
  source: "connect",
  recognize(errorLike) {
    if (!(errorLike instanceof ConnectError)) return undefined;
    const reason = errorLike.findDetails(ErrorInfoSchema)[0]?.reason;
    const network = errorLike.code === Code.Unknown && errorLike.cause instanceof TypeError;
    const mapped = network ? undefined : CODE_KIND[errorLike.code];
    const message = errorLike.rawMessage !== "" ? errorLike.rawMessage : (reason ?? "");
    const codeName = Code[errorLike.code];
    return makeEnvelope(network ? "network" : (mapped?.kind ?? "unknown"), message, {
      source: "connect",
      reason,
      code: codeName !== undefined ? toWireWord(codeName) : undefined,
      http: mapped?.http,
      stack: errorLike.stack,
      cause: errorLike.cause,
    });
  },
};
