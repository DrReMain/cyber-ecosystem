// The engine registry: code → GameModule. H-A statically registers the
// first module; the E7 batches extend this map (dynamic import per game is
// the natural growth point once the set grows past a handful).
import type { GameModule } from "./contract";
import { redDotModule } from "./games/red-dot/red-dot";

const modules: ReadonlyMap<string, GameModule> = new Map([[redDotModule.code, redDotModule]]);

export function getGameModule(code: string): GameModule | null {
  return modules.get(code) ?? null;
}
