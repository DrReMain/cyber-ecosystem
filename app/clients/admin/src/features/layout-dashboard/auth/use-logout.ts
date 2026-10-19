import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useNavigate, useRouterState } from "@tanstack/react-router";
import { LOGIN_PATH } from "../area";
import { logoutFn } from "./logout.fn";
import { broadcastSessionKick } from "./session-kick";

export function useLogout() {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const href = useRouterState({ select: (s) => s.location.href });

  return useMutation({
    mutationFn: () => logoutFn(),
    onSuccess: async () => {
      queryClient.clear();
      broadcastSessionKick();
      const search = new URLSearchParams({ redirect: href });
      await navigate({ href: `${LOGIN_PATH}?${search.toString()}`, viewTransition: false });
    },
  });
}
