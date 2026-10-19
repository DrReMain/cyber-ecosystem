import { makeSearchParser } from "@cyber-ecosystem/shared-antd/use-search";
import { z } from "zod";

export const usersSearchSchema = z.looseObject({
  keyword: z.string().optional(),
  status: z.enum(["on", "off"]).optional(),
  deptId: z.string().optional(),
});

export const parseUsersSearch = makeSearchParser(usersSearchSchema);
