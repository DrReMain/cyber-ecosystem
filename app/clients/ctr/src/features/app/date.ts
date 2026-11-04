// A UTC timestamp's local calendar date, YYYY-MM-DD: the walls quote a
// window's last local day, and a UTC slice would hand an early-morning
// local tail (UTC+8 00:00–07:59) the previous day.
export function localDateOf(ms: number): string {
  const d = new Date(ms);
  const pad = (n: number) => String(n).padStart(2, "0");
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`;
}

// The same local-calendar read with the minute — the pay result's receipt
// row quotes the paid moment, where an hour's slice is a wrong receipt,
// not a wrong deadline.
export function localDateTimeOf(ms: number): string {
  const d = new Date(ms);
  const pad = (n: number) => String(n).padStart(2, "0");
  return `${localDateOf(ms)} ${pad(d.getHours())}:${pad(d.getMinutes())}`;
}

// The letterbox and feedback threads quote the prototype's compact stamps:
// MM-DD on list rows, MM-DD HH:mm on ticket turns — a year prefix would
// crowd rows that already carry a dot and a badge.
export function localShortDateOf(ms: number): string {
  const d = new Date(ms);
  const pad = (n: number) => String(n).padStart(2, "0");
  return `${pad(d.getMonth() + 1)}-${pad(d.getDate())}`;
}

export function localShortDateTimeOf(ms: number): string {
  const d = new Date(ms);
  const pad = (n: number) => String(n).padStart(2, "0");
  return `${localShortDateOf(ms)} ${pad(d.getHours())}:${pad(d.getMinutes())}`;
}
