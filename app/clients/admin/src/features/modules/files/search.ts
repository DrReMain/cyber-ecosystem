import {
  makeSearchParser,
  pageNoField,
  pageSizeField,
  sortField,
} from "@cyber-ecosystem/shared-antd/use-search";
import { z } from "zod";

export const filesSearchSchema = z.looseObject({
  name: z.string().optional(),
  contentType: z.string().optional(),
  status: z.enum(["UPLOADING", "CONFIRMED", "PROCESSING", "FAILED"]).optional(),
  source: z.enum(["CLIENT_UPLOAD", "SERVER_GENERATED"]).optional(),
  pageNo: pageNoField(),
  pageSize: pageSizeField(20),
  sort: sortField("createdAt:desc"),
});

export type FilesSearch = z.output<typeof filesSearchSchema>;

export const parseFilesSearch = makeSearchParser(filesSearchSchema, []);
