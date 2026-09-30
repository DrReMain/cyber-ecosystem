import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useNavigate, useSearch } from "@tanstack/react-router";
import { safeRedirect } from "#/features/login-default";
import { AREA_PATH, LOGIN_PATH, LOGIN_ROUTE_ID } from "../area";
import { loginFn } from "./login.fn";

export function usePasswordLogin() {
  const navigate = useNavigate();
  const search = useSearch({ from: LOGIN_ROUTE_ID });
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (input: { email: string; password: string }) => loginFn({ data: input }),
    onSuccess: async () => {
      queryClient.clear();
      await navigate({
        href: safeRedirect(search.redirect, { home: AREA_PATH, self: LOGIN_PATH }),
      });
    },
  });
}
