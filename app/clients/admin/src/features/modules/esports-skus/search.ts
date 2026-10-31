import { makeSearchParser } from "@cyber-ecosystem/shared-antd/use-search";
import { z } from "zod";

export const skusSearchSchema = z.looseObject({
  shelf: z.enum(["all", "on", "off"]).optional(),
});

export type SkusSearch = z.output<typeof skusSearchSchema>;

export const parseSkusSearch = makeSearchParser(skusSearchSchema, []);
