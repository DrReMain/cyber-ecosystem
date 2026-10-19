import clsx from "clsx";

interface SkeletonProps {
  /** Size and radius overrides; defaults give a tile-proportioned block. */
  className?: string;
}

/**
 * Decorative placeholder block for pending content shapes. Purely visual
 * (aria-hidden): the host composes the a11y story — a container with
 * role="status" and an sr-only "loading" label next to the skeleton blocks.
 */
export function Skeleton({ className }: Readonly<SkeletonProps>) {
  return (
    <div
      aria-hidden
      className={clsx("rounded-tile bg-ink/8 motion-safe:animate-pulse", className)}
    />
  );
}
