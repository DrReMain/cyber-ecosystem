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

// The redeem-moment shape: day + clock, the grain an operator reconciles
// a hand-out by — the day-only fullDate is too coarse for "when was this
// used".
export function fullTime(ts?: Timestamp | null): string {
  const d = dateOf(ts);
  if (!d) return "-";
  const pad = (n: number) => String(n).padStart(2, "0");
  return `${fullDate(ts)} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`;
}
