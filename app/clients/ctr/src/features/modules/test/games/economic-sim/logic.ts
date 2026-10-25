// Pure law for economic_sim — the stepped-decay minion, the four tap
// windows, and the window's payout. The decay is stepped, not continuous
// (a hundred minus whole speed steps every interval), which preserves the
// legacy geometry exactly: a melee minion's health only ever lands on
// multiples of ten (so its perfect window is unreachable and good is the
// best it offers), while a ranged minion stepping by twelve does land on
// four — the only perfect tap in the game. The payout mirrors the server
// scorer's floor arithmetic, so every fact the engine mints passes the
// server's per-tap gold cross-check by construction.

export type EsMinionType = "melee" | "ranged" | "cannon";

export type EsWindow = "perfect" | "good" | "early" | "too_early";

export const ECONOMIC_SIM_DEFAULTS = {
  decayIntervalMs: 100,
  decayMelee: 10,
  decayRanged: 12,
  decayCannon: 15,
  respawnDelayMs: 200,
  perfectWindowMelee: 5,
  perfectWindowCannon: 8,
  goodWindowMelee: 10,
  goodWindowCannon: 15,
  goldMelee: 21,
  goldCannon: 90,
  earlyHealthCap: 30,
  healthPenaltyPct: 5,
  itemCost1: 400,
  itemCost2: 800,
  itemCost3: 1200,
  maxItems: 3,
  perfectBonusMs: 1000,
};

export type EsParams = typeof ECONOMIC_SIM_DEFAULTS;

export function decaySpeed(type: EsMinionType, params: EsParams): number {
  switch (type) {
    case "cannon":
      return params.decayCannon;
    case "ranged":
      return params.decayRanged;
    default:
      return params.decayMelee;
  }
}

/** Stepped remaining health: whole speed steps per elapsed interval,
 * clamped to 0-100. A negative elapsed (pre-spawn) reads as full. */
export function healthAt(
  elapsedSinceSpawnMs: number,
  type: EsMinionType,
  params: EsParams,
): number {
  const steps = Math.floor(Math.max(0, elapsedSinceSpawnMs) / params.decayIntervalMs);
  return Math.max(0, Math.min(100, 100 - steps * decaySpeed(type, params)));
}

/** The four windows the tap lands in; perfect and good confirm a last hit. */
export function windowOf(health: number, type: EsMinionType, params: EsParams): EsWindow {
  const perfect = type === "cannon" ? params.perfectWindowCannon : params.perfectWindowMelee;
  const good = type === "cannon" ? params.goodWindowCannon : params.goodWindowMelee;
  if (health <= perfect) {
    return "perfect";
  }
  if (health <= good) {
    return "good";
  }
  if (health <= params.earlyHealthCap) {
    return "early";
  }
  return "too_early";
}

export function confirmsHit(window: EsWindow): boolean {
  return window === "perfect" || window === "good";
}

/** The window's payout — floored shares, mirroring the server's arithmetic
 * so the per-tap gold cross-check passes by construction. */
export function goldOf(window: EsWindow, type: EsMinionType, params: EsParams): number {
  const base = type === "cannon" ? params.goldCannon : params.goldMelee;
  switch (window) {
    case "perfect":
      return base;
    case "good":
      return Math.floor(base * 0.6);
    case "early":
      return Math.floor(base * 0.3);
    default:
      return 0;
  }
}

/** The next minion's type draw — the legacy weights: a tenth cannon, four
 * tenths ranged, half melee. */
export function nextMinionType(rand: () => number): EsMinionType {
  const roll = rand();
  if (roll > 0.9) {
    return "cannon";
  }
  if (roll > 0.5) {
    return "ranged";
  }
  return "melee";
}
