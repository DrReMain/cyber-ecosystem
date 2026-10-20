// Same 6–60 law as profile/setup/age.tsx (the range must reach below 13 or
// the guardian gate has no reachable input); re-declared here because
// modules never import each other's internals.

export const MONTHS = Array.from({ length: 12 }, (_, i) => i + 1);

export function years(): number[] {
  const now = new Date().getFullYear();
  return Array.from({ length: 55 }, (_, i) => now - 60 + i);
}

export function ageOf(year: number, month: number): number {
  const now = new Date();
  let age = now.getFullYear() - year;
  if (now.getMonth() + 1 < month) {
    age--;
  }
  return age;
}
