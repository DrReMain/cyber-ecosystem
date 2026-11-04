export const MONTHS = Array.from({ length: 12 }, (_, i) => i + 1);

// 6–60 years old: the range must reach below 13 or the guardian gate has no
// reachable input (the prototype widened it for exactly this reason).
export function years(): number[] {
  const now = new Date().getFullYear();
  return Array.from({ length: 55 }, (_, i) => now - 60 + i);
}

// Month-granularity age (the contract stores no day).
export function ageOf(year: number, month: number): number {
  const now = new Date();
  let age = now.getFullYear() - year;
  if (now.getMonth() + 1 < month) {
    age--;
  }
  return age;
}

// The pay gate's reckoning, the mirror of the server's conservative rule:
// the line clears only from the month AFTER the birthday month (次月起算),
// so it is stricter than ageOf >= line — the birthday month itself still
// blocks. The server's CreateOrder refusal stays the authority; this mirror
// only keeps the wall's pre-check from pitching a purchase the gate refuses.
export function ageLineReached(year: number, month: number, line: number): boolean {
  const now = new Date();
  const elapsed = now.getFullYear() - year;
  if (elapsed !== line) {
    return elapsed > line;
  }
  return now.getMonth() + 1 > month;
}
