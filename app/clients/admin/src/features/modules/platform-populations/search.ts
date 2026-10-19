import { makeSearchParser } from "@cyber-ecosystem/shared-antd/use-search";
import { z } from "zod";

export const populationsSearchSchema = z.looseObject({
  keyword: z.string().optional(),
});

export const parsePopulationsSearch = makeSearchParser(populationsSearchSchema);
