// The client-side stand-in arranger: composes the four-ability lineup from
// the test catalog by the genre order rule. Every rule here mirrors the
// frozen domain law (§0.5: per-ability one test, genre pools by category,
// universal feeds both) and is replaced wholesale by the E4 session engine
// — same shape, server authority.

export type Genre = "fps" | "moba";

export interface ArrangedStep {
  ability: string;
  code: string;
  name: string;
}

export interface CatalogTest {
  code?: string | null;
  name?: string | null;
  category?: string | null;
  ability?: string | null;
}

// render_order per genre — the prototype's frozen sequence (§0.5).
export const GENRE_ORDER: Readonly<Record<Genre, readonly string[]>> = {
  fps: ["reactionSpeed", "aimingAccuracy", "trackingStability", "skillReleaseAccuracy"],
  moba: ["resourceAllocation", "decisionEfficiency", "reactionSpeed", "skillReleaseAccuracy"],
};

// A genre pool = its own category plus universal; per ability the first
// catalog row (created order) wins. Tests without an ability attribution
// never enter a pool.
export function arrange(tests: CatalogTest[], genre: Genre): ArrangedStep[] {
  return GENRE_ORDER[genre]
    .map((ability) =>
      tests.find(
        (t) =>
          t.ability === ability &&
          (t.category === genre || t.category === "universal") &&
          t.code != null,
      ),
    )
    .filter((t): t is CatalogTest => t !== undefined)
    .map((t) => ({ ability: t.ability as string, code: t.code as string, name: t.name ?? "" }));
}

// Default genre = majority vote of the member's game preferences mapped
// through the catalog's category; tie or no classified picks → null (the
// pick page owns the manual choice).
export function defaultGenre(
  preferenceCodes: readonly string[],
  games: readonly { id?: string | null; category?: string | null }[],
): Genre | null {
  const categoryOf = new Map(games.map((g) => [g.id ?? "", g.category ?? ""]));
  let fps = 0;
  let moba = 0;
  for (const code of preferenceCodes) {
    const category = categoryOf.get(code);
    if (category === "fps") fps++;
    if (category === "moba") moba++;
  }
  if (fps === 0 && moba === 0) return null;
  return fps === moba ? null : fps > moba ? "fps" : "moba";
}
