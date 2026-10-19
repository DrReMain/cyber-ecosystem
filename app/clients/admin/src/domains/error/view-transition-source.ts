import { makeEnvelope, type SourceAdapter } from "@cyber-ecosystem/shared-error";

const ABORT_PREFIX = "Transition was aborted";

export const viewTransitionSource: SourceAdapter = {
  source: "view-transition",
  recognize: (error) => {
    if (!(error instanceof Error || error instanceof DOMException)) return undefined;
    if (!error.message.startsWith(ABORT_PREFIX)) return undefined;
    return makeEnvelope("canceled", error.message, { reason: "VIEW_TRANSITION_ABORTED" });
  },
};
