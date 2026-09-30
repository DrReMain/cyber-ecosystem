export interface RouterProgressConfig {
  /** Height of the progress indicator in CSS pixels. */
  height: number;
  /** Minimum percentage used when a navigation starts. */
  initialMin: number;
  /** Maximum percentage used when a navigation starts. */
  initialMax: number;
  /** Maximum percentage reached before the navigation resolves. */
  trickleCeiling: number;
  /**
   * Fraction of the remaining distance added per 60 Hz frame.
   * The animation normalizes this value against elapsed time.
   */
  trickleRate: number;
  /** Minimum time in milliseconds from completion until the indicator hides. */
  finishDelay: number;
  /** Duration in milliseconds of the transition to 100%. */
  transitionDuration: number;
  /** CSS color or gradient used for the indicator. */
  color: string;
  /** Duration in milliseconds of the color-cycle animation. Set to 0 to disable it. */
  cycleDuration: number;
  /** Viewport edge where the indicator is rendered. */
  position: "top" | "bottom";
  /** CSS stacking order for the indicator container. */
  zIndex: number;
}

export const DEFAULT_ROUTER_PROGRESS_CONFIG = Object.freeze({
  height: 2,
  initialMin: 5,
  initialMax: 12,
  trickleCeiling: 90,
  trickleRate: 0.015,
  finishDelay: 300,
  transitionDuration: 200,
  color:
    "linear-gradient(90deg, #f87171, #fb923c, #fbbf24, #a3e635, #34d399, #38bdf8, #818cf8, #c084fc, #f87171)",
  cycleDuration: 6000,
  position: "top",
  zIndex: 9999,
} satisfies RouterProgressConfig);

function finiteNumber(value: number | undefined, fallback: number) {
  return typeof value === "number" && Number.isFinite(value) ? value : fallback;
}

function clamp(value: number, min: number, max: number) {
  return Math.min(Math.max(value, min), max);
}

/**
 * Merges consumer options with defaults and normalizes values that can affect
 * layout or animation timing. This keeps JavaScript consumers as safe as TypeScript consumers.
 */
export function resolveRouterProgressConfig(
  overrides: Partial<RouterProgressConfig> = {},
): RouterProgressConfig {
  const initialMin = clamp(
    finiteNumber(overrides.initialMin, DEFAULT_ROUTER_PROGRESS_CONFIG.initialMin),
    0,
    99,
  );
  const initialMax = clamp(
    finiteNumber(overrides.initialMax, DEFAULT_ROUTER_PROGRESS_CONFIG.initialMax),
    initialMin,
    99,
  );

  return {
    height: Math.max(0, finiteNumber(overrides.height, DEFAULT_ROUTER_PROGRESS_CONFIG.height)),
    initialMin,
    initialMax,
    trickleCeiling: clamp(
      finiteNumber(overrides.trickleCeiling, DEFAULT_ROUTER_PROGRESS_CONFIG.trickleCeiling),
      initialMax,
      99.9,
    ),
    trickleRate: clamp(
      finiteNumber(overrides.trickleRate, DEFAULT_ROUTER_PROGRESS_CONFIG.trickleRate),
      0,
      1,
    ),
    finishDelay: Math.max(
      0,
      finiteNumber(overrides.finishDelay, DEFAULT_ROUTER_PROGRESS_CONFIG.finishDelay),
    ),
    transitionDuration: Math.max(
      0,
      finiteNumber(overrides.transitionDuration, DEFAULT_ROUTER_PROGRESS_CONFIG.transitionDuration),
    ),
    color:
      typeof overrides.color === "string" && overrides.color.trim().length > 0
        ? overrides.color
        : DEFAULT_ROUTER_PROGRESS_CONFIG.color,
    cycleDuration: Math.max(
      0,
      finiteNumber(overrides.cycleDuration, DEFAULT_ROUTER_PROGRESS_CONFIG.cycleDuration),
    ),
    position: overrides.position === "bottom" ? "bottom" : "top",
    zIndex: Math.trunc(finiteNumber(overrides.zIndex, DEFAULT_ROUTER_PROGRESS_CONFIG.zIndex)),
  };
}
