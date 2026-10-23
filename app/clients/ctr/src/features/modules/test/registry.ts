import type { GameModule } from "./contract";
import { dualZoneModule } from "./games/dual-zone/dual-zone";
import { flashAimingModule } from "./games/flash-aiming/flash-aiming";
import { movingTargetModule } from "./games/moving-target/moving-target";
import { multiTargetModule } from "./games/multi-target/multi-target";
import { redDotModule } from "./games/red-dot/red-dot";
import { skillTrajectoryModule } from "./games/skill-trajectory/skill-trajectory";
import { targetRecognitionModule } from "./games/target-recognition/target-recognition";
import { trackingMovingModule } from "./games/tracking-moving/tracking-moving";

const modules: ReadonlyMap<string, GameModule> = new Map([
  [redDotModule.code, redDotModule],
  [flashAimingModule.code, flashAimingModule],
  [movingTargetModule.code, movingTargetModule],
  [multiTargetModule.code, multiTargetModule],
  [targetRecognitionModule.code, targetRecognitionModule],
  [trackingMovingModule.code, trackingMovingModule],
  [dualZoneModule.code, dualZoneModule],
  [skillTrajectoryModule.code, skillTrajectoryModule],
]);

export function getGameModule(code: string): GameModule | null {
  return modules.get(code) ?? null;
}
