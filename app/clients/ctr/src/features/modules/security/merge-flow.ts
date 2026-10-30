import type { MergeOfferDTO } from "./security.fn";

// The merge confirm page's handoff, in-memory on purpose: the offer verdict
// is a one-shot read (the wechat branch burned its proof to materialize),
// so it can neither ride the URL nor be re-derived — a fresh visit with an
// empty hand honestly bounces back to the security page.
export interface MergeFlow {
  offer?: MergeOfferDTO;
  phone?: string;
}

let flow: MergeFlow = {};

export function setMergeFlow(next: MergeFlow): void {
  flow = next;
}

export function readMergeFlow(): MergeFlow {
  return flow;
}

export function clearMergeFlow(): void {
  flow = {};
}
