import { useMutation } from "@connectrpc/connect-query";
import { submitTestAttempt } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/test-TestService_connectquery";

/**
 * The verdict loop's write half — one submit per natural attempt end. Toast
 * feedback stays off: the end face's verdict card owns the error face
 * inline (view(error) + resubmit), so a toast would double-report the same
 * refusal. Session-flow refusals (expired session, exhausted budget)
 * arrive as normal error envelopes and render through the same card.
 */
export function useSubmitAttempt() {
  return useMutation(submitTestAttempt, { meta: { feedback: false } });
}
