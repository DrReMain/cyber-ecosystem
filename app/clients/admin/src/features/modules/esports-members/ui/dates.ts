import { type Timestamp, timestampMs } from "@bufbuild/protobuf/wkt";

// Transport delivers WKT fields JSON-shaped (an ISO string here), not as
// Timestamp instances — the same narrowing the table column layer applies.
export const dateOf = (ts?: Timestamp | null): Date | null => {
  if (ts == null) return null;
  if (typeof ts === "string") {
    const d = new Date(ts);
    return Number.isNaN(d.getTime()) ? null : d;
  }
  return new Date(timestampMs(ts));
};

export const shortOf = (d: Date | null): string =>
  d ? `${String(d.getMonth() + 1).padStart(2, "0")}-${String(d.getDate()).padStart(2, "0")}` : "-";

export function shortDate(ts?: Timestamp | null): string {
  return shortOf(dateOf(ts));
}

export function fullDate(ts?: Timestamp | null): string {
  const d = dateOf(ts);
  return d
    ? `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, "0")}-${String(d.getDate()).padStart(2, "0")}`
    : "-";
}
