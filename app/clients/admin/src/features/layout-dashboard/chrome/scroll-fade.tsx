import clsx from "clsx";

const SIDE = {
  "x-end":
    "inset-e-0 inset-y-0 w-5 bg-linear-to-l from-container/85 via-container/40 rtl:bg-linear-to-r",
  "x-start":
    "inset-s-0 inset-y-0 w-5 bg-linear-to-r from-container/85 via-container/40 rtl:bg-linear-to-l",
  "y-end": "bottom-0 inset-x-0 h-10 bg-linear-to-t from-container/85 via-container/40",
  "y-start": "top-0 inset-x-0 h-10 bg-linear-to-b from-container/85 via-container/40",
} as const;

interface ScrollFadeProps {
  axis: "x" | "y";
  show: boolean;
  side: "end" | "start";
}

export function ScrollFade({ axis, show, side }: Readonly<ScrollFadeProps>) {
  return (
    <span
      aria-hidden
      className={clsx(
        "pointer-events-none absolute opacity-0 transition-opacity",
        SIDE[`${axis}-${side}`],
        show && "opacity-100",
      )}
    />
  );
}
