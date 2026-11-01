import {
  makeSearchParser,
  pageNoField,
  pageSizeField,
} from "@cyber-ecosystem/shared-antd/use-search";
import { z } from "zod";

// The version table's consent link lands here with doc/ver prefilled
// (the prototype's ?type=&ver=), so both ride the URL as wire/doc names
// and the version string exactly as the table shows it.
export const consentsSearchSchema = z.looseObject({
  doc: z.enum(["AGREEMENT_DOC_USER", "AGREEMENT_DOC_PRIVACY"]).optional(),
  ver: z.string().max(20).optional(),
  keyword: z.string().max(32).optional(),
  from: z.string().date().optional(),
  to: z.string().date().optional(),
  pageNo: pageNoField(),
  pageSize: pageSizeField(20),
});

export type ConsentsSearch = z.output<typeof consentsSearchSchema>;

export const parseConsentsSearch = makeSearchParser(consentsSearchSchema, []);
