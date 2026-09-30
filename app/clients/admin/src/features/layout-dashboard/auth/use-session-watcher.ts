import { useQuery, useQueryClient } from "@tanstack/react-query";
import { useRouter } from "@tanstack/react-router";
import { useEffect, useRef } from "react";
import { LOGIN_PATH } from "../area";
import { sessionQuery } from "./session.fn";
import { broadcastSessionKick, subscribeSessionKick } from "./session-kick";

const loginHref = (originHref: string) => {
  const search = new URLSearchParams({ expired: "true", redirect: originHref });
  return `${LOGIN_PATH}?${search.toString()}`;
};

const isOnLogin = (pathname: string) => {
  return pathname === LOGIN_PATH || pathname.startsWith(`${LOGIN_PATH}/`);
};

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
          href: loginHref(router.state.location.href),
          viewTransition: false,
        });
      }
    }
    prevStatus.current = status;
  }, [data, router]);

  useEffect(
    () =>
      subscribeSessionKick(() => {
        if (isOnLogin(router.state.location.pathname)) return;
        queryClient.removeQueries({ queryKey: sessionQuery.queryKey });
        void router.navigate({
          href: loginHref(router.state.location.href),
          viewTransition: false,
        });
      }),
    [router, queryClient],
  );
}
