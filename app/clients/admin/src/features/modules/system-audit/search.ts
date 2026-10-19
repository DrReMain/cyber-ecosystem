import {
  makeSearchParser,
  pageNoField,
  pageSizeField,
  sortField,
} from "@cyber-ecosystem/shared-antd/use-search";
import { z } from "zod";

export const auditSearchSchema = z.looseObject({
  actor: z.string().optional(),
  operation: z.string().optional(),
  status: z.number().optional(),
  denyReason: z.string().optional(),
  lens: z.enum(["all", "denied"]).default("all"),
  createdAtA: z.number().optional(),
  createdAtZ: z.number().optional(),
  pageNo: pageNoField(),
  pageSize: pageSizeField(20),
  sort: sortField("createdAt:desc"),
});

export type AuditSearch = z.output<typeof auditSearchSchema>;

export const parseAuditSearch = makeSearchParser(auditSearchSchema, [
  "createdAtA",
  "createdAtZ",
  "status",
]);
