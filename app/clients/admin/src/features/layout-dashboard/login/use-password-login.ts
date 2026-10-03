import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useNavigate, useSearch } from "@tanstack/react-router";
import { safeRedirect } from "#/features/login-default";
import { AREA_PATH, LOGIN_PATH, LOGIN_ROUTE_ID } from "../area";
import { defaultLanding } from "../auth/landing";
import { sessionQuery } from "../auth/session.fn";
import { loginFn } from "./login.fn";

export function usePasswordLogin() {
  const navigate = useNavigate();
  const search = useSearch({ from: LOGIN_ROUTE_ID });
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (input: { email: string; password: string }) => loginFn({ data: input }),
    onSuccess: async () => {
      queryClient.clear();
      const home =
        search.redirect === undefined
          ? await queryClient
              .query(sessionQuery)
              .then((s) => (s.status === "authed" ? defaultLanding(s.permissions) : AREA_PATH))
              .catch(() => AREA_PATH)
          : AREA_PATH;
      await navigate({
        href: safeRedirect(search.redirect, { home, self: LOGIN_PATH }),
      });
    },
  });
}
