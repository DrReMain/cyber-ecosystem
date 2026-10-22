import clsx from "clsx";

/*
 * Ambient backdrop — blurred solid-tone orbs bleeding off the page edges.
 * Decorative only (aria-hidden). The prototype art-directs orbs per page (two
 * or three, from a five-tone palette); the shared default is the login trio.
 * Rest state (opacity .26, blur 70px, solid color) is the prototype's .blob
 * form; motion-safe drift keeps that rest as its 0%/100% keyframe so any
 * still frame stays pixel-honest. Position/size classes are dot-free —
 * Tailwind v4's extractor splits arbitrary values containing decimal points
 * into dead classes.
 */

export type AmbientTone = "purple" | "blue" | "cyan" | "indigo" | "teal";

export interface AmbientOrb {
  /** Position + size utilities (e.g. "top-[-70px] left-[-60px] size-[min(67vw,260px)]"). */
  className: string;
  tone: AmbientTone;
}

/* Literal classes only — interpolated class names are invisible to Tailwind's
 * extractor, so tones and drift phases resolve through static maps. */
const toneClass: Record<AmbientTone, string> = {
  purple: "bg-ambient-purple",
  blue: "bg-ambient-blue",
  cyan: "bg-ambient-cyan",
  indigo: "bg-ambient-indigo",
  teal: "bg-ambient-teal",
};

const driftClass = [
  "motion-safe:animate-ambient-a",
  "motion-safe:animate-ambient-b",
  "motion-safe:animate-ambient-c",
] as const;

/** The login page's trio — also the default for pages not yet art-directed. */
export const loginOrbs: readonly AmbientOrb[] = [
  { className: "top-[-70px] left-[-60px] size-[min(67vw,260px)]", tone: "purple" },
  { className: "top-[320px] right-[-80px] size-[min(57vw,220px)]", tone: "blue" },
  { className: "bottom-[-40px] left-[40px] size-[min(52vw,200px)]", tone: "cyan" },
];

export interface AmbientBackdropProps {
  orbs?: readonly AmbientOrb[];
}

export function AmbientBackdrop({ orbs = loginOrbs }: Readonly<AmbientBackdropProps>) {
  return (
    <div aria-hidden className="pointer-events-none absolute inset-0 z-0 overflow-hidden">
      {orbs.map((orb, index) => (
        <div
          className={clsx(
            "absolute rounded-full opacity-26 blur-[70px]",
            toneClass[orb.tone],
            driftClass[index % driftClass.length],
            orb.className,
          )}
          key={orb.className}
        />
      ))}
    </div>
  );
}
