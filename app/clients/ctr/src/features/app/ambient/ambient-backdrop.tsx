/*
 * Ambient backdrop — three drifting blurred orbs on the page canvas.
 * Decorative only (aria-hidden): sizes, corners, drift phases, and rest
 * opacities are the legacy client's calibrated values; motion-safe keeps
 * the orbs static (at their rest opacity) under reduced motion.
 */
export function AmbientBackdrop() {
  return (
    <div aria-hidden className="pointer-events-none absolute inset-0 z-0 overflow-hidden">
      <div className="bg-ambient-a motion-safe:animate-ambient-a absolute top-[-14%] left-[-18%] size-[min(78vw,440px)] rounded-full opacity-[0.42] blur-[68px]" />
      <div className="bg-ambient-b motion-safe:animate-ambient-b absolute right-[-16%] bottom-[-22%] size-[min(85vw,480px)] rounded-full opacity-[0.36] blur-[68px]" />
      <div className="bg-ambient-c motion-safe:animate-ambient-c absolute top-[38%] left-[32%] size-[min(55vw,320px)] rounded-full opacity-[0.32] blur-[68px]" />
    </div>
  );
}
