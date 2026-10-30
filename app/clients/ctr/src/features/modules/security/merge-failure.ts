import { System } from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/error_reason_pb";
import { systemReason } from "#/domains/error";
import { m } from "#/paraglide/messages";

export function mergeFailureCopy(reason: string | undefined): string {
  switch (systemReason(reason)) {
    case System.MERGE_NOT_MERGEABLE:
      return m.error_SYSTEM_MERGE_NOT_MERGEABLE();
    case System.MERGE_NOT_BOUND:
      return m.error_SYSTEM_MERGE_NOT_BOUND();
    case System.SMS_CODE_INVALID:
      return m.error_SYSTEM_SMS_CODE_INVALID();
    default:
      return m.security_error_generic();
  }
}
