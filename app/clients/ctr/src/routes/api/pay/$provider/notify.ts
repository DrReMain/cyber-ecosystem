import { createClient } from "@connectrpc/connect";
import { OrderService } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/order_pb";
import { createFileRoute } from "@tanstack/react-router";
import { createCtrTransport } from "#/services/connect/transport";

// The channel's callback face at the edge. The push arrives as raw channel
// HTTP — a signed envelope — and the only job here is byte-faithful
// forwarding into the service's public notify RPC: the envelope's channel
// signature is the entire authentication, and rewriting anything about it
// (headers, body bytes) would void it before the service ever sees it.
export const Route = createFileRoute("/api/pay/$provider/notify")({
  server: {
    handlers: {
      POST: async ({ request, params }) => {
        // The service's 64KiB body ceiling, enforced before the read: the
        // edge must not buffer into memory a push the service would only
        // reject. A chunked push without a declared length still lands on
        // the service-side check.
        const declared = Number(request.headers.get("content-length") ?? "0");
        if (declared > 65536) {
          return fail(413, "payload too large");
        }
        const body = new Uint8Array(await request.arrayBuffer());
        const headers: Record<string, string> = {};
        request.headers.forEach((value, key) => {
          headers[key] = value;
        });
        try {
          const res = await createClient(OrderService, createCtrTransport()).notifyPayment({
            provider: params.provider,
            headers,
            body,
          });
          if (res.handled) {
            return ack();
          }
          // A well-formed push this build does not fold: no ack, so the
          // channel's own retry schedule keeps it visible.
          return fail(501, "unprocessed event");
        } catch {
          // Forwarding or channel-verification failure — never leak the
          // cause to the caller; the channel retries, the service logs.
          return fail(500, "notify processing failed");
        }
      },
    },
  },
});

function ack() {
  return new Response(JSON.stringify({ code: "SUCCESS", message: "成功" }), {
    status: 200,
    headers: { "Content-Type": "application/json" },
  });
}

function fail(status: number, message: string) {
  return new Response(JSON.stringify({ code: "FAIL", message }), {
    status,
    headers: { "Content-Type": "application/json" },
  });
}
