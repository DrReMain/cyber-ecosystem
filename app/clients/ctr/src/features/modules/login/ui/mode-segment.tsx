import { m } from "#/paraglide/messages";

export type LoginMode = "code" | "password";

// The prototype's two-mode segment, DOM order kept — code first and
// default now that the system SMS contract exists; the on-state is the
// ambient blue pill.
export function ModeSegment({
  mode,
  onSwitch,
}: Readonly<{ mode: LoginMode; onSwitch: (target: LoginMode) => void }>) {
  const tab = (target: LoginMode, label: string) => (
    <button
      className={
        mode === target
          ? "bg-ambient-blue shadow-seg-on text-ink flex-1 rounded-[9px] py-2.75 text-center text-[13.5px] font-bold"
          : "text-ink-muted flex-1 rounded-[9px] py-2.75 text-center text-[13.5px] active:opacity-60"
      }
      onClick={() => onSwitch(target)}
      type="button"
    >
      {label}
    </button>
  );
  return (
    <div className="border-line-soft bg-surface-input/90 rounded-control flex gap-1 border p-1">
      {tab("code", m.login_tab_code())}
      {tab("password", m.login_tab_password())}
    </div>
  );
}
