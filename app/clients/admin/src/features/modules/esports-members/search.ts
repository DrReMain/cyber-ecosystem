import {
  makeSearchParser,
  pageNoField,
  pageSizeField,
} from "@cyber-ecosystem/shared-antd/use-search";
import { z } from "zod";

export const membersSearchSchema = z.looseObject({
  tab: z.enum(["all", "active", "expiring", "expired"]).optional(),
  source: z.enum(["order", "manual", "legacy_code"]).optional(),
  pageNo: pageNoField(),
  pageSize: pageSizeField(20),
});

export type MembersSearch = z.output<typeof membersSearchSchema>;

export const parseMembersSearch = makeSearchParser(membersSearchSchema, []);
