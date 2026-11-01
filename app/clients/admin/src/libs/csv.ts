// The one CSV shape the activation faces share: quoted cells, doubled
// quotes, a BOM so Excel opens Chinese headers correctly on the first try.
const escapeCell = (v: string | number | undefined | null) =>
  `"${String(v ?? "").replaceAll('"', '""')}"`;

export const downloadCsv = (
  filename: string,
  header: string[],
  rows: (string | number | undefined | null)[][],
) => {
  const csv = [header, ...rows].map((r) => r.map(escapeCell).join(",")).join("\n");
  const blob = new Blob([`﻿${csv}`], { type: "text/csv" });
  const a = document.createElement("a");
  a.href = URL.createObjectURL(blob);
  a.download = filename;
  a.click();
  URL.revokeObjectURL(a.href);
};
