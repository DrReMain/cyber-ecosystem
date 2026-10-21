import { System } from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/error_reason_pb";

// Wire reasons carry their service prefix ("SYSTEM_..."); pages branch on
// the typed enum so a drifted comparison is a compile error, not a silent
// fall-through.
export function systemReason(reason: string | undefined): System | undefined {
  if (reason === undefined || !reason.startsWith("SYSTEM_")) {
    return undefined;
  }
  const byName = System as unknown as Record<string, System | undefined>;
  return byName[reason.slice("SYSTEM_".length)];
}
