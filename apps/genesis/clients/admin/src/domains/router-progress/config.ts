export interface RouterProgressConfig {
  /** Bar height in pixels */
  height: number
  /** Starting width (%) when navigation begins */
  initialMin: number
  initialMax: number
  /** Maximum width (%) during trickle phase */
  trickleCeiling: number
  /** Trickle increment as fraction of remaining distance */
  trickleRate: number
  /** Interval (ms) between trickle increments */
  trickleInterval: number
  /** Time (ms) to wait after reaching 100% before hiding */
  finishDelay: number
  /** CSS transition duration (ms) for width changes */
  transitionDuration: number
}

export const defaultConfig: RouterProgressConfig = {
  height: 2,
  initialMin: 5,
  initialMax: 12,
  trickleCeiling: 90,
  trickleRate: 0.1,
  trickleInterval: 200,
  finishDelay: 300,
  transitionDuration: 200,
}
