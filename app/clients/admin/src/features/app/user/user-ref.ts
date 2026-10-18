export type UserRefView =
  | { kind: "empty" }
  | { kind: "raw"; id: string; deleted: boolean }
  | { kind: "known"; account: string; id: string };

export function userRefView(
  id: string | undefined,
  nameOf: (id: string) => string | undefined,
  directoryKnown: boolean,
): UserRefView {
  const key = id ?? "";
  if (key === "") return { kind: "empty" };
  const account = nameOf(key);
  if (account === undefined) return { kind: "raw", id: key, deleted: directoryKnown };
  return { kind: "known", account, id: key };
}
