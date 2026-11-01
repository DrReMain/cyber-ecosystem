import {
  makeSearchParser,
  pageNoField,
  pageSizeField,
} from "@cyber-ecosystem/shared-antd/use-search";
import { z } from "zod";

export const batchCodesSearchSchema = z.looseObject({
  // The three-way segment; absent = all codes of the batch.
  status: z.enum(["unused", "redeemed"]).optional(),
  // The find-one search box; absent = the whole (filtered) table.
  code: z.string().optional(),
  pageNo: pageNoField(),
  pageSize: pageSizeField(50),
});

export type BatchCodesSearch = z.output<typeof batchCodesSearchSchema>;

export const parseBatchCodesSearch = makeSearchParser(batchCodesSearchSchema, []);
