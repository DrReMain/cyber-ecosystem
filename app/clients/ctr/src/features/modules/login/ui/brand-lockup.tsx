import { Zap } from "lucide-react";
import { m } from "#/paraglide/messages";

// The brand lockup plus the prototype's .gap8 breathing room between it
// and the segment.
export function BrandLockup() {
  return (
    <>
      <div className="flex items-center gap-3">
        <div className="bg-logo shadow-inset-ring flex h-11 w-11 items-center justify-center rounded-xl">
          <Zap aria-hidden className="text-white" size={22} />
        </div>
        <div>
          <div className="text-ink text-xl font-extrabold">{m.login_brand_name()}</div>
          <div className="text-ink-tertiary mt-0.5 text-[11.5px]">{m.login_brand_tagline()}</div>
        </div>
      </div>
      <div aria-hidden className="h-2 shrink-0" />
    </>
  );
}
