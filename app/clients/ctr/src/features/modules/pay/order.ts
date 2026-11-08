import type { Sku } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/sku_pb";
import { fenToYuan } from "#/features/app/fen";

// Runtime string enums — the transport delivers enums as value-name
// strings while the generated types keep numeric enums; these constants
// restate what actually arrives.
export const SKU_KIND_TRIAL = "SKU_KIND_TRIAL";
export const SKU_KIND_MEMBERSHIP = "SKU_KIND_MEMBERSHIP";
export const ORDER_SCOPE_TRIAL = "ORDER_SCOPE_TRIAL";
export const ORDER_STATE_PAID = "ORDER_STATE_PAID";
export const ORDER_STATE_CLOSED = "ORDER_STATE_CLOSED";
export const CREDENTIAL_INVOKE = "PAY_CREDENTIAL_KIND_INVOKE";
export const CREDENTIAL_JUMP_URL = "PAY_CREDENTIAL_KIND_JUMP_URL";
export const CREDENTIAL_QR_CODE = "PAY_CREDENTIAL_KIND_QR_CODE";

// Not a channel: a mobile browser outside WeChat cannot pay — the H5
// permission is deferred by the provider, and a QR shown on the same phone
// has nothing to scan it. The wall routes this value to the "open inside
// WeChat" guide instead of ordering on a channel the provider would refuse.
export const CHANNEL_DEGRADED = "wechat_degraded";

export interface PurchaseFace {
  row?: Sku;
  yuan: string;
}

// The wall's two purchase faces, priced off the catalog with the same
// pick rule the report wall pitches (the priciest trial row, the
// cheapest membership row) so both faces quote one number per kind —
// no picked row stays an invented number ("—").
export function purchaseFaces(skus: readonly Sku[]): { trial: PurchaseFace; member: PurchaseFace } {
  const ofKind = (kind: string) =>
    skus
      .filter((s) => String(s.kind) === kind)
      .sort((a, b) => Number(a.priceFen ?? 0) - Number(b.priceFen ?? 0));
  const trial = ofKind(SKU_KIND_TRIAL).at(-1);
  const member = ofKind(SKU_KIND_MEMBERSHIP)[0];
  const face = (row?: Sku): PurchaseFace => ({
    row,
    yuan: row?.priceFen !== undefined ? fenToYuan(row.priceFen) : "—",
  });
  return { trial: face(trial), member: face(member) };
}

// The channel is an environment fact the caller decides and the server
// validates against the active provider's set — provider-blind by
// contract, so the same three names serve wechat and the mock backend
// (the mock mirrors wechat's set string for string).
export function channelOf(userAgent: string): string {
  if (/MicroMessenger/i.test(userAgent)) {
    return "wechat_jsapi";
  }
  if (/Android|iPhone|iPad|Mobile/i.test(userAgent)) {
    // The H5 leg stays reachable to the server; while the provider's H5
    // permission is deferred this branch degrades instead of sending it.
    return CHANNEL_DEGRADED;
  }
  return "wechat_native";
}

// The in-WeChat continuation (JSAPI). The bridge object exists only
// inside WeChat's webview; the credential carries the package string the
// invoke needs. The full parameter assembly (appId/timeStamp/nonceStr/
// paySign) belongs to the real-channel integration leg — registered as a
// gap, this branch keeps the honest shape.
export function invokeWechatPay(packageValue: string, onDone: (ok: boolean) => void): void {
  const bridge = (
    window as {
      WeixinJSBridge?: {
        invoke: (
          api: string,
          params: Record<string, string>,
          cb: (r: { err_msg?: string }) => void,
        ) => void;
      };
    }
  ).WeixinJSBridge;
  bridge?.invoke("getBrandWCPayRequest", { package: packageValue }, (r) => {
    onDone(r.err_msg === "get_brand_wcpay_request:ok");
  });
}
