import { m } from "#/paraglide/messages";

// One report-type vocabulary for the module's three faces — the list
// filter, the generate modal, and the table's type column.
export const reportTypeOptions = [
  { label: () => m.esports_report_type_cognitive(), value: "REPORT_TYPE_COGNITIVE" },
  { label: () => m.esports_report_type_career(), value: "REPORT_TYPE_CAREER" },
  { label: () => m.esports_report_type_school(), value: "REPORT_TYPE_SCHOOL" },
];
