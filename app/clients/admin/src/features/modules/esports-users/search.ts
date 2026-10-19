import {
  makeSearchParser,
  pageNoField,
  pageSizeField,
} from "@cyber-ecosystem/shared-antd/use-search";
import { z } from "zod";

export const usersSearchSchema = z.looseObject({
  keyword: z.string().optional(),
  status: z.enum(["on", "off"]).optional(),
  pageNo: pageNoField(),
  pageSize: pageSizeField(20),
});

export type UsersSearch = z.output<typeof usersSearchSchema>;

export const parseUsersSearch = makeSearchParser(usersSearchSchema, []);
