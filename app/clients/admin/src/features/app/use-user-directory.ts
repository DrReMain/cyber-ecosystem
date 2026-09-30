import { useQuery } from "@connectrpc/connect-query";
import { listUsers } from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/user-UserService_connectquery";
import { useMemo } from "react";

export function useUserDirectory(enabled: boolean) {
  const usersQuery = useQuery(listUsers, { page: { all: true } }, { enabled });
  const byId = useMemo(() => {
    const map = new Map<string, string>();
    for (const user of usersQuery.data?.list ?? []) {
      if (user.id !== undefined && user.id !== "") {
        map.set(user.id, user.email ?? user.id);
      }
    }
    return map;
  }, [usersQuery.data]);
  const byEmail = useMemo(() => {
    const map = new Map<string, string>();
    for (const user of usersQuery.data?.list ?? []) {
      if (user.email !== undefined && user.email !== "") {
        map.set(user.email, user.id ?? user.email);
      }
    }
    return map;
  }, [usersQuery.data]);
  return { known: usersQuery.isSuccess, byId, byEmail };
}
