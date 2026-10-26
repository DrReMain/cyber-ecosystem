import { m } from "#/paraglide/messages";

/** The start CTA card: the desc, the primary button, and the hint line. */
export function CtaCard({
  authed,
  hasGames,
  onStart,
}: Readonly<{ authed: boolean; hasGames: boolean; onStart: () => void }>) {
  return (
    <div className="border-line bg-surface-card shadow-card rounded-card border px-4 py-4.5">
      <div className="text-ink text-base font-extrabold">{m.home_cta_title()}</div>
      <p className="text-ink-tertiary mt-1.25 text-[11.5px] leading-[1.7]">
        {/* The desc follows the games state alone (prototype law): a
            guest has no preferences yet, so it reads the pick-games
            line — the login-first ordering rides on the button and
            hint, not the desc. */}
        {!hasGames ? m.home_cta_desc_need_games() : m.home_cta_desc_ready()}
      </p>
      <button
        className="bg-cta-primary shadow-cta-primary rounded-control mt-3.5 w-full px-4 py-3.5 text-[15px] font-bold text-white active:opacity-80"
        onClick={onStart}
        type="button"
      >
        {!authed ? m.home_cta_login() : !hasGames ? m.home_cta_pick() : m.home_cta_start()}
      </button>
      {!(authed && hasGames) && (
        <p className="text-ink-tertiary mt-2.25 text-center text-[11px]">
          {!authed ? m.home_hint_login() : m.home_hint_games()}
        </p>
      )}
    </div>
  );
}
