import type { GameModule } from "./contract";
import { characterCounterModule } from "./games/character-counter/character-counter";
import { dualZoneModule } from "./games/dual-zone/dual-zone";
import { economicSimModule } from "./games/economic-sim/economic-sim";
import { flashAimingModule } from "./games/flash-aiming/flash-aiming";
import { movingTargetModule } from "./games/moving-target/moving-target";
import { multiTargetModule } from "./games/multi-target/multi-target";
import { numberStarsModule } from "./games/number-stars/number-stars";
import { offenseDefenseModule } from "./games/offense-defense/offense-defense";
import { redDotModule } from "./games/red-dot/red-dot";
import { resourceMatchModule } from "./games/resource-match/resource-match";
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
  [offenseDefenseModule.code, offenseDefenseModule],
  [resourceMatchModule.code, resourceMatchModule],
  [numberStarsModule.code, numberStarsModule],
  [characterCounterModule.code, characterCounterModule],
  [economicSimModule.code, economicSimModule],
]);

export function getGameModule(code: string): GameModule | null {
  return modules.get(code) ?? null;
}
