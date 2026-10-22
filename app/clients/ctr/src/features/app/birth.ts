export const MONTHS = Array.from({ length: 12 }, (_, i) => i + 1);

// 6–60 years old: the range must reach below 13 or the guardian gate has no
// reachable input (the prototype widened it for exactly this reason).
export function years(): number[] {
  const now = new Date().getFullYear();
  return Array.from({ length: 55 }, (_, i) => now - 60 + i);
}

// Month-granularity age (the contract stores no day); the pay gate's
// month-level approximation of the birthday-next-day rule lands with PAY3.
export function ageOf(year: number, month: number): number {
  const now = new Date();
  let age = now.getFullYear() - year;
  if (now.getMonth() + 1 < month) {
    age--;
  }
  return age;
}
