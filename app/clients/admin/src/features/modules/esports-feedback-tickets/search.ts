import {
  makeSearchParser,
  pageNoField,
  pageSizeField,
} from "@cyber-ecosystem/shared-antd/use-search";
import { z } from "zod";

export const ticketsSearchSchema = z.looseObject({
  status: z.enum(["pending", "replied", "closed"]).optional(),
  pageNo: pageNoField(),
  pageSize: pageSizeField(20),
});

export type TicketsSearch = z.output<typeof ticketsSearchSchema>;

// The queue's only filter is the state segment — no keyword field rides
// the URL because the contract offers no text search to back one.
export const parseTicketsSearch = makeSearchParser(ticketsSearchSchema, []);
