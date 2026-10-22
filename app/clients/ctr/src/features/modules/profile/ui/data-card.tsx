import type { Profile as ProfilePB } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/profile_pb";
import { useNavigate } from "@tanstack/react-router";
import { ChevronRight } from "lucide-react";
import { m } from "#/paraglide/messages";

const CARD_CLASS = "border-line-accent-blue bg-identity rounded-card border px-4 py-[15px]";

/**
 * The editable profile facts: one row per setup step, each hopping into the
 * wizard's matching step (edit path).
 */
export function DataCard({ profile }: Readonly<{ profile: ProfilePB | undefined }>) {
  const navigate = useNavigate();
  const games = profile?.games ?? [];
  // A year without a month renders the year alone — the pair is only shown
  // as a pair (the birth gate in after-login.ts treats it as incomplete).
  const birth =
    profile?.birthYear == null
      ? m.profile_unset()
      : profile.birthMonth == null
        ? String(profile.birthYear)
        : `${profile.birthYear} · ${profile.birthMonth}`;
  const rows: { key: string; label: string; value: string; to: string }[] = [
    {
      key: "nickname",
      label: m.profile_nickname_label(),
      to: "/profile/setup/nickname",
      value: profile?.nickname ?? m.profile_unset(),
    },
    {
      key: "birth",
      label: m.profile_birth_label(),
      to: "/profile/setup/age",
      value: birth,
    },
    {
      key: "games",
      label: m.profile_games_label(),
      to: "/profile/setup/games",
      value: games.length > 0 ? m.profile_games_count({ n: games.length }) : m.profile_unset(),
    },
  ];
  return (
    <div className={CARD_CLASS}>
      <div className="text-ink mb-1 text-[14.5px] font-extrabold">{m.profile_card_data()}</div>
      {rows.map((row, index) => (
        <button
          className={`flex w-full items-center justify-between py-[9px] text-left active:opacity-60 ${
            index < rows.length - 1 ? "border-b border-[rgba(55,65,81,0.4)]" : ""
          }`}
          key={row.key}
          onClick={() => void navigate({ to: row.to })}
          type="button"
        >
          <span className="text-ink-secondary text-[13px] font-bold">{row.label}</span>
          <span className="ml-3 flex min-w-0 items-center gap-1.5">
            <span className="text-ink-muted truncate text-[12px]">{row.value}</span>
            <ChevronRight aria-hidden className="text-ink-tertiary shrink-0" size={14} />
          </span>
        </button>
      ))}
    </div>
  );
}
