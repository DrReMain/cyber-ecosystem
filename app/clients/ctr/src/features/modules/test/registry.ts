import type { GameModule } from "./contract";
import { flashAimingModule } from "./games/flash-aiming/flash-aiming";
import { multiTargetModule } from "./games/multi-target/multi-target";
import { redDotModule } from "./games/red-dot/red-dot";
import { targetRecognitionModule } from "./games/target-recognition/target-recognition";

const modules: ReadonlyMap<string, GameModule> = new Map([
  [redDotModule.code, redDotModule],
  [flashAimingModule.code, flashAimingModule],
  [multiTargetModule.code, multiTargetModule],
  [targetRecognitionModule.code, targetRecognitionModule],
]);

export function getGameModule(code: string): GameModule | null {
  return modules.get(code) ?? null;
}
