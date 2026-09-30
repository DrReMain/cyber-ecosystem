import { Tag } from "antd";
import { m } from "#/paraglide/messages";

export type UserRefView =
  | { kind: "empty" }
  | { kind: "raw"; id: string; deleted: boolean }
  | { kind: "known"; email: string; id: string };

export function userRefView(
  id: string | undefined,
  nameOf: (id: string) => string | undefined,
  directoryKnown: boolean,
): UserRefView {
  const key = id ?? "";
  if (key === "") return { kind: "empty" };
  const email = nameOf(key);
  if (email === undefined) return { kind: "raw", id: key, deleted: directoryKnown };
  return { kind: "known", email, id: key };
}

export function UserDeletedTag() {
  return <Tag className="m-0">{m.common_user_deleted()}</Tag>;
}
