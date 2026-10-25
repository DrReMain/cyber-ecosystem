// offense_defense pure logic. The game itself is a decision card: each
// round shows one scenario, the player picks a stance inside the window,
// and the fact stream carries the pick with its round-relative time. The
// correctness judgment never lives here — the server scores against the
// seeded answer key.

import type { OdScenario } from "./scenarios";

export const OFFENSE_DEFENSE_DEFAULTS = {
  totalRounds: 10,
  timePerRoundMs: 2000,
  feedbackMs: 800,
} as const;

// Legacy pace conversions: the old client ran 10 rounds at a 15-second
// window on an untimed global clock; the config-driven 30-second run
// clock pins the new window at the spec's decision window (the 1200 ms
// standard-mode narrowing waits for a difficulty-aware boot — the engine
// reads one window today).
export function pickScenarios(
  bank: readonly OdScenario[],
  count: number,
  rand: () => number,
): OdScenario[] {
  const idx = bank.map((_, i) => i);
  // Partial Fisher–Yates: the first `count` slots of one shuffle, so a
  // fixed rand draws a fixed lineup with no duplicates.
  const take = Math.min(count, idx.length);
  for (let i = 0; i < take; i += 1) {
    const j = i + Math.floor(rand() * (idx.length - i));
    const drawn = idx[j] ?? 0;
    idx[j] = idx[i] ?? 0;
    idx[i] = drawn;
  }
  return idx.slice(0, take).map((i) => bank[i] as OdScenario);
}
