import { useQuery, useQueryClient } from "@tanstack/react-query";
import { useRouter } from "@tanstack/react-router";
import { useEffect, useRef } from "react";
import { sessionQuery } from "./session.fn";
import { broadcastSessionKick, subscribeSessionKick } from "./session-kick";

const isOnLogin = (pathname: string) => {
  return pathname === "/login" || pathname.startsWith("/login/");
};

// Admin's session watcher, mounted app-wide (AppLayout hangs it once at
// the root route): the session probe stays live — react-query revalidates
// on focus, remount, and reconnect — and the moment it drops out of
// authed the app boots to login with the expired notice, so a server-side
// ban or revoke reaches the member without a manual refresh. Root mount
// means public surfaces run it too: an anonymous probe short-circuits at
// the cookie check, the login tree never acts on a transition, and a
// kick only boots a tab that believed itself authed (a public tab just
// drops its cached probe).
export function useSessionWatcher() {
  const router = useRouter();
  const queryClient = useQueryClient();
  const { data } = useQuery({ ...sessionQuery, throwOnError: false });
  const prevStatus = useRef<string | null>(null);

  useEffect(() => {
    const status = data?.status ?? null;
    if (status === "unreachable") return;
    if (prevStatus.current === "authed" && status !== null && status !== "authed") {
      if (!isOnLogin(router.state.location.pathname)) {
        broadcastSessionKick();
        void router.navigate({
          search: { expired: true, next: router.state.location.href },
          to: "/login",
          viewTransition: false,
        });
      }
    }
    prevStatus.current = status;
  }, [data, router]);

  useEffect(
    () =>
      subscribeSessionKick(() => {
        // A deliberate sign-out carries no expired notice — the surprise
        // flag is reserved for sessions killed server-side.
        if (isOnLogin(router.state.location.pathname)) return;
        if (prevStatus.current !== "authed") return;
        queryClient.removeQueries({ queryKey: sessionQuery.queryKey });
        void router.navigate({
          search: { next: router.state.location.href },
          to: "/login",
          viewTransition: false,
        });
      }),
    [router, queryClient],
  );
}
