import {
  makeSearchParser,
  pageNoField,
  pageSizeField,
} from "@cyber-ecosystem/shared-antd/use-search";
import { z } from "zod";

export const connectionsSearchSchema = z.looseObject({
  pageNo: pageNoField(),
  pageSize: pageSizeField(20),
});

export type ConnectionsSearch = z.output<typeof connectionsSearchSchema>;

export const parseConnectionsSearch = makeSearchParser(connectionsSearchSchema, []);
