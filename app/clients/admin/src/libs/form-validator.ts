import { makeEnvelope } from "@cyber-ecosystem/shared-error";
import type { output, ZodType } from "zod";

export function zodValidator<T extends ZodType>(schema: T) {
  return (input: unknown): output<T> => {
    const parsed = schema.safeParse(input);
    if (!parsed.success)
      throw makeEnvelope("business", "", { reason: "GENERAL_ERROR_VALIDATION_FAILED" });
    return parsed.data;
  };
}
