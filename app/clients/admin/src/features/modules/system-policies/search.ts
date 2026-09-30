import { makeSearchParser } from "@cyber-ecosystem/shared-antd/use-search";
import { z } from "zod";

export const policiesSearchSchema = z.looseObject({
  keyword: z.string().optional(),
  kind: z.enum(["time_window", "calendar"]).optional(),
});

export const parsePoliciesSearch = makeSearchParser(policiesSearchSchema);
