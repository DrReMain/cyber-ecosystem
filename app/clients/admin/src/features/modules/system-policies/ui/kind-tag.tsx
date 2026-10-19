import type { Policy } from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/policy_pb";
import { Tag } from "antd";
import { m } from "#/paraglide/messages";
import { kindOf } from "../params";

export function KindTag({ policy }: Readonly<{ policy: Policy }>) {
  const kind = kindOf(policy);
  if (kind === "time_window") return <Tag color="blue">{m.system_policies_kind_time_window()}</Tag>;
  if (kind === "calendar") return <Tag color="gold">{m.system_policies_kind_calendar()}</Tag>;
  return <Tag>-</Tag>;
}
