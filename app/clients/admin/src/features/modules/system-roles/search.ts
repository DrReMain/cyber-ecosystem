import { makeSearchParser } from "@cyber-ecosystem/shared-antd/use-search";
import { z } from "zod";

export const rolesSearchSchema = z.looseObject({
  keyword: z.string().optional(),
  status: z.enum(["on", "off"]).optional(),
});

export const parseRolesSearch = makeSearchParser(rolesSearchSchema);
