import type { GameModule } from "./contract";
import { redDotModule } from "./games/red-dot/red-dot";

const modules: ReadonlyMap<string, GameModule> = new Map([[redDotModule.code, redDotModule]]);

export function getGameModule(code: string): GameModule | null {
  return modules.get(code) ?? null;
}
