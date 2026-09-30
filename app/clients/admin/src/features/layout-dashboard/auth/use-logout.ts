import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useNavigate } from "@tanstack/react-router";
import { LOGIN_PATH } from "../area";
import { logoutFn } from "./logout.fn";
import { broadcastSessionKick } from "./session-kick";

export function useLogout() {
  const navigate = useNavigate();
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: () => logoutFn(),
    onSuccess: async () => {
      queryClient.clear();
      broadcastSessionKick();
      await navigate({ href: LOGIN_PATH, viewTransition: false });
    },
  });
}
