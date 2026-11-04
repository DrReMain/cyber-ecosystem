import { useQuery as useConnectQuery } from "@connectrpc/connect-query";
import { getUnreadCount } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/message-MessageService_connectquery";
import { useQuery } from "@tanstack/react-query";
import { sessionQuery } from "#/features/auth/session.fn";

// The letterbox's unread count as the chrome asks it: guests skip the
// ask (the badge is a member's fact), and every surface — the tabbar
// slot, the home bell — shares this one cache entry, so mark-read
// invalidations walk them all back to zero together.
export function useUnreadCount(): number | undefined {
  const session = useQuery({ ...sessionQuery, throwOnError: false });
  const authed = session.data?.status === "authed";
  const unreadQuery = useConnectQuery(getUnreadCount, {}, { enabled: authed, throwOnError: false });
  return unreadQuery.data?.count;
}
