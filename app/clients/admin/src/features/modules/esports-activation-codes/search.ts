import { ReportType } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/entitlement_pb";
import {
  makeSearchParser,
  pageNoField,
  pageSizeField,
} from "@cyber-ecosystem/shared-antd/use-search";
import { z } from "zod";

export const codesSearchSchema = z.looseObject({
  reportType: z
    .enum(["REPORT_TYPE_COGNITIVE", "REPORT_TYPE_CAREER", "REPORT_TYPE_SCHOOL"])
    .optional(),
  // The URL carries the three-way pick; the request folds it to a boolean
  // only when a side is chosen.
  state: z.enum(["live", "disabled"]).optional(),
  pageNo: pageNoField(),
  pageSize: pageSizeField(20),
});

export type CodesSearch = z.output<typeof codesSearchSchema>;

export const parseCodesSearch = makeSearchParser(codesSearchSchema, []);

// The URL carries wire names (readable, and what the transport answers);
// requests carry the generated enum — this module owns the one translation.
export const reportTypeFromWire = (wire?: string): ReportType | undefined => {
  switch (wire) {
    case "REPORT_TYPE_COGNITIVE":
      return ReportType.COGNITIVE;
    case "REPORT_TYPE_CAREER":
      return ReportType.CAREER;
    case "REPORT_TYPE_SCHOOL":
      return ReportType.SCHOOL;
    default:
      return undefined;
  }
};
