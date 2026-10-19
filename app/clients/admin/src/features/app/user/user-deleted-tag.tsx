import { Tag } from "antd";
import { m } from "#/paraglide/messages";

export function UserDeletedTag() {
  return <Tag className="m-0">{m.common_user_deleted()}</Tag>;
}
