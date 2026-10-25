// Pure law for character_counter — the role roster, the counter-relation
// mirror, and the round draw. The relation table is a verbatim mirror of
// the twelve seeded question-bank rows (scenario "<role>|<countersRole>",
// answer counter_score 20): content changes owe both sides, and like the
// offense_defense bank the mirror feeds nothing but the local feedback
// face — the judging key stays the server's. The wire vocabulary is the
// English role token; the view maps tokens to locale labels.

export const CC_ROSTER = ["Tank", "Fighter", "Mage", "Marksman", "Support", "Assassin"] as const;
export type CcRole = (typeof CC_ROSTER)[number];

export interface CounterRelation {
  readonly counter: CcRole;
  readonly countered: CcRole;
  readonly score: number;
}

export const COUNTER_RELATIONS: readonly CounterRelation[] = [
  { counter: "Tank", countered: "Fighter", score: 20 },
  { counter: "Tank", countered: "Assassin", score: 20 },
  { counter: "Fighter", countered: "Marksman", score: 20 },
  { counter: "Fighter", countered: "Mage", score: 20 },
  { counter: "Mage", countered: "Tank", score: 20 },
  { counter: "Mage", countered: "Marksman", score: 20 },
  { counter: "Marksman", countered: "Fighter", score: 20 },
  { counter: "Marksman", countered: "Support", score: 20 },
  { counter: "Support", countered: "Assassin", score: 20 },
  { counter: "Support", countered: "Mage", score: 20 },
  { counter: "Assassin", countered: "Marksman", score: 20 },
  { counter: "Assassin", countered: "Mage", score: 20 },
];

export const CHARACTER_COUNTER_DEFAULTS = {
  roundCount: 3,
  lineupSize: 5,
  speedRefSec: 8,
};

/** Shuffle the roster and take the first `size` — the enemy lineup draws
 * five distinct roles (an in-lineup repeat would fail the server's
 * edited-history gate, so the draw never produces one). */
export function drawEnemyLineup(size: number, rand: () => number): CcRole[] {
  const pool = [...CC_ROSTER];
  const take = Math.min(size, pool.length);
  for (let i = 0; i < take; i += 1) {
    const j = i + Math.floor(rand() * (pool.length - i));
    const a = pool[i] ?? CC_ROSTER[0];
    const b = pool[j] ?? a;
    pool[i] = b;
    pool[j] = a;
  }
  return pool.slice(0, take);
}

export interface CounterHit {
  readonly counter: CcRole;
  readonly countered: CcRole;
  readonly score: number;
}

/** The hitting pairs of a lineup against an enemy — the local face's only
 * judgment; the server re-scores against the seeded key. */
export function hitPairs(
  selected: readonly string[],
  enemy: readonly string[],
  relations: readonly CounterRelation[] = COUNTER_RELATIONS,
): CounterHit[] {
  const hits: CounterHit[] = [];
  for (const relation of relations) {
    if (selected.includes(relation.counter) && enemy.includes(relation.countered)) {
      hits.push(relation);
    }
  }
  return hits;
}

export function counterScore(hits: readonly CounterHit[]): number {
  return hits.reduce((sum, hit) => sum + hit.score, 0);
}
