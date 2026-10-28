import { createQueryOptions } from "@connectrpc/connect-query";
import { getSessionReport } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/report-ReportService_connectquery";
import type { QueryClient } from "@tanstack/react-query";
import { connectTransport } from "#/services/connect";

// The session read-back face's client seam. One input shape — the session
// id — so the key built here plus this module's invalidation reaches every
// cache entry the UI can create; the test module invalidates it beside the
// match session after a submit (the new attempt is the scoreboard's next
// line), and the match and session-done faces read it for their receipts.

export function sessionReportKey(sessionId: string) {
  return createQueryOptions(getSessionReport, { sessionId }, { transport: connectTransport })
    .queryKey;
}

export function invalidateSessionReport(queryClient: QueryClient, sessionId: string): void {
  void queryClient.invalidateQueries({ queryKey: sessionReportKey(sessionId) });
}
