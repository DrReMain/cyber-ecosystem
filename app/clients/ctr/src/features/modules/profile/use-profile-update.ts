import { useMutation } from "@connectrpc/connect-query";
import { updateProfile } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/profile-ProfileService_connectquery";
import { useQueryClient } from "@tanstack/react-query";
import { sessionQuery } from "#/features/auth/session.fn";

export function useProfileUpdate() {
  const queryClient = useQueryClient();
  return useMutation(updateProfile, {
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: sessionQuery.queryKey });
    },
  });
}
