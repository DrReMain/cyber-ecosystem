import {
  makeSearchParser,
  pageNoField,
  pageSizeField,
} from "@cyber-ecosystem/shared-antd/use-search";
import { z } from "zod";

export const ordersSearchSchema = z.looseObject({
  state: z.enum(["all", "paid", "pending"]).optional(),
  scope: z.enum(["all", "trial", "period"]).optional(),
  keyword: z.string().optional(),
  pageNo: pageNoField(),
  pageSize: pageSizeField(20),
});

export type OrdersSearch = z.output<typeof ordersSearchSchema>;

export const parseOrdersSearch = makeSearchParser(ordersSearchSchema, ["keyword"]);
