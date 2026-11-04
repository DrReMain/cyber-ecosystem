// A UTC timestamp's local calendar date, YYYY-MM-DD: the walls quote a
// window's last local day, and a UTC slice would hand an early-morning
// local tail (UTC+8 00:00–07:59) the previous day.
export function localDateOf(ms: number): string {
  const d = new Date(ms);
  const pad = (n: number) => String(n).padStart(2, "0");
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`;
}
