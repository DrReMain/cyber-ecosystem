import type { PropsWithChildren } from "react";
import { BrandCanvas } from "./ui/brand-canvas";
import { ConsolePanel } from "./ui/console-panel";

const KEYFRAMES = `
@keyframes nc-drift { from { transform: scale(1) translateX(0); } to { transform: scale(1.06) translateX(-14px); } }
@keyframes nc-vscan { from { top: -2px; } to { top: 100%; } }
@keyframes nc-pulse { 0%, 100% { opacity: 1; } 50% { opacity: 0.55; } }`;

export function LoginShell({ children }: Readonly<PropsWithChildren>) {
  return (
    <div className="flex min-h-svh bg-[#05070d] text-[#e8ecf4] max-[899px]:flex-col">
      <style>{KEYFRAMES}</style>
      <BrandCanvas />
      <ConsolePanel>{children}</ConsolePanel>
    </div>
  );
}
