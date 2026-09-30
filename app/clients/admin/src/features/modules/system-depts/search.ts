import { makeSearchParser } from "@cyber-ecosystem/shared-antd/use-search";
import { z } from "zod";

export const deptsSearchSchema = z.looseObject({
  keyword: z.string().optional(),
});

export const parseDeptsSearch = makeSearchParser(deptsSearchSchema);
