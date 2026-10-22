import { useQuery } from "@connectrpc/connect-query";
import type { Game } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/game_pb";
import { listGames } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/game-GameService_connectquery";
import { getProfile } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/profile-ProfileService_connectquery";
import { useNavigate } from "@tanstack/react-router";
import { X } from "lucide-react";
import { useEffect, useMemo, useState } from "react";
import { Button } from "react-aria-components";
import { toast } from "sonner";
import { PROFILE_STALE_MS } from "#/features/auth/profile-cache";
import { m } from "#/paraglide/messages";
import { useProfileUpdate } from "../use-profile-update";
import { StepFrame } from "./step-frame";

// The product cap for custom titles. The server cap is config-driven and
// looser (its error owns the truth); the client shows the product number.
const CUSTOM_CAP = 3;

interface Groups {
  order: string[];
  byCategory: Map<string, Game[]>;
}

function groupGames(games: Game[]): Groups {
  const byCategory = new Map<string, Game[]>();
  for (const game of games) {
    const category = game.category ?? "";
    const bucket = byCategory.get(category);
    if (bucket === undefined) {
      byCategory.set(category, [game]);
    } else {
      bucket.push(game);
    }
  }
  const order = [...byCategory.keys()].sort((a, b) => {
    // fps and moba lead (the voting vocabulary); anything the admin adds
    // later follows alphabetically.
    const rank = (c: string): number => (c === "fps" ? 0 : c === "moba" ? 1 : 2);
    return rank(a) - rank(b) || a.localeCompare(b);
  });
  return { byCategory, order };
}

// Only fps/moba get a header (the voting vocabulary); other categories
// render as bare grids until the product says otherwise.
const hasCategoryHeader = (category: string): boolean => category === "fps" || category === "moba";

export function SetupGamesPage() {
  const navigate = useNavigate();
  const profileQuery = useQuery(getProfile, {}, { staleTime: PROFILE_STALE_MS });
  const gamesQuery = useQuery(listGames, {});
  const updateMutation = useProfileUpdate();
  const [picked, setPicked] = useState<string[]>([]);
  const [customs, setCustoms] = useState<string[]>([]);
  const [draft, setDraft] = useState("");
  const [ready, setReady] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const p = profileQuery.data?.profile;
    if (!ready && p !== undefined) {
      setPicked((p.games ?? []).map((g) => g.gameId).filter((id): id is string => id != null));
      setCustoms((p.games ?? []).map((g) => g.customName).filter((n): n is string => n != null));
      setReady(true);
    }
  }, [profileQuery.data?.profile, ready]);

  const groups = useMemo(() => groupGames(gamesQuery.data?.games ?? []), [gamesQuery.data?.games]);

  function toggle(id: string): void {
    setError(null);
    setPicked((prev) => (prev.includes(id) ? prev.filter((x) => x !== id) : [...prev, id]));
  }

  function addCustom(): void {
    const value = draft.trim();
    if (value === "" || customs.includes(value) || picked.includes(value)) {
      if (value !== "") {
        toast.error(m.profile_setup_games_dup());
      }
      return;
    }
    if (customs.length >= CUSTOM_CAP) {
      toast.error(m.profile_setup_games_cap_toast({ cap: CUSTOM_CAP }));
      return;
    }
    setError(null);
    setCustoms((prev) => [...prev, value]);
    setPicked((prev) => [...prev, value]);
    setDraft("");
  }

  function removeCustom(name: string): void {
    setCustoms((prev) => prev.filter((x) => x !== name));
    setPicked((prev) => prev.filter((x) => x !== name));
  }

  const fpsVotes = picked.filter((id) =>
    (groups.byCategory.get("fps") ?? []).some((g) => g.id === id),
  ).length;
  const mobaVotes = picked.filter((id) =>
    (groups.byCategory.get("moba") ?? []).some((g) => g.id === id),
  ).length;
  const totalPicked = picked.length;

  async function onDone() {
    if (totalPicked === 0) {
      setError(m.profile_setup_games_err());
      return;
    }
    const catalogIds = new Set((gamesQuery.data?.games ?? []).map((g) => g.id ?? ""));
    try {
      await updateMutation.mutateAsync({
        gamePreference: {
          customNames: customs,
          gameIds: picked.filter((id) => catalogIds.has(id)),
        },
      });
    } catch {
      // The mutation cache already toasts the reason; swallowing the
      // rejection keeps it out of the window capture (Sentry noise).
      return;
    }
    await navigate({ to: "/" });
  }

  return (
    <StepFrame onBack={() => void navigate({ to: "/profile/setup/age" })} step={3}>
      <div className="rounded-lg border border-line bg-surface-card p-4">
        <div className="mb-1 font-extrabold text-ink text-sm">
          {m.profile_setup_games_title()} <span className="text-danger-text">*</span>
        </div>
        <p className="mb-3 text-[11.5px] text-ink-tertiary leading-relaxed">
          {m.profile_setup_games_desc()}
        </p>
        {groups.order.map((category) => (
          <section key={category}>
            {hasCategoryHeader(category) && (
              <div
                className={`mt-2 mb-1.5 font-bold text-[11.5px] first:mt-0 ${
                  category === "fps" ? "text-accent-fps" : "text-accent-moba"
                }`}
              >
                {category.toUpperCase()}
              </div>
            )}
            <div className="grid grid-cols-2 gap-2">
              {(groups.byCategory.get(category) ?? []).map((game) => {
                const on = picked.includes(game.id ?? "");
                return (
                  <button
                    aria-pressed={on}
                    className={`rounded-lg border px-2 py-3 text-center text-[12.5px] active:opacity-60 ${
                      on
                        ? "border-focus bg-focus/20 font-bold text-accent-info"
                        : "border-line bg-surface-content text-ink-secondary"
                    }`}
                    key={game.id ?? game.name}
                    onClick={() => toggle(game.id ?? "")}
                    type="button"
                  >
                    {game.name}
                  </button>
                );
              })}
            </div>
          </section>
        ))}
        {totalPicked > 0 && (
          <p className="mt-3 text-[11px] text-ink-tertiary">
            {m.profile_setup_games_vote({
              fps: fpsVotes,
              moba: mobaVotes,
              n: totalPicked,
            })}
            {fpsVotes === mobaVotes
              ? m.profile_setup_games_tie()
              : ` ${m.profile_setup_games_default({ type: fpsVotes > mobaVotes ? "FPS" : "MOBA" })}`}
          </p>
        )}
      </div>

      <div className="rounded-lg border border-line bg-surface-card p-4">
        <div className="mb-1 font-extrabold text-ink text-sm">
          {m.profile_setup_games_custom_title()}
        </div>
        <p className="mb-3 text-[11.5px] text-ink-tertiary leading-relaxed">
          {m.profile_setup_games_custom_desc()}
        </p>
        <div className="flex gap-2">
          <input
            aria-label={m.profile_setup_games_custom_title()}
            className="min-w-0 flex-1 rounded-sm border border-line bg-surface-input px-3 py-3 text-ink text-sm outline-none focus:border-focus"
            maxLength={30}
            name="customGame"
            onChange={(event) => setDraft(event.target.value)}
            onKeyDown={(event) => {
              if (event.key === "Enter") {
                event.preventDefault();
                addCustom();
              }
            }}
            placeholder={m.profile_custom_placeholder()}
            type="text"
            value={draft}
          />
          <button
            className="w-20 shrink-0 rounded-sm border border-line px-3 py-3 font-bold text-[13px] text-accent-info disabled:text-ink-tertiary disabled:opacity-45"
            disabled={customs.length >= CUSTOM_CAP}
            onClick={addCustom}
            type="button"
          >
            {m.profile_setup_games_add()}
          </button>
        </div>
        <div
          className={`mt-2 flex justify-between text-[11px] ${
            customs.length >= CUSTOM_CAP ? "text-status-pending" : "text-ink-tertiary"
          }`}
        >
          <span>
            {m.profile_setup_games_cap({
              cap: CUSTOM_CAP,
              n: customs.length,
            })}
          </span>
        </div>
        {customs.length > 0 && (
          <div className="mt-2.5 flex flex-wrap gap-2">
            {customs.map((name) => (
              <button
                aria-label={m.profile_setup_games_remove({ name })}
                className="tap-zone flex items-start gap-1.5 rounded-lg border border-line bg-surface-content px-2.5 py-1.5 text-ink-secondary text-xs active:opacity-60"
                key={name}
                onClick={() => removeCustom(name)}
                type="button"
              >
                <span className="min-w-0 break-all text-left leading-snug">{name}</span>
                <X aria-hidden className="mt-0.5 shrink-0 text-ink-tertiary" size={12} />
              </button>
            ))}
          </div>
        )}
      </div>

      {error !== null && (
        <p className="text-danger-text text-sm" role="alert">
          {error}
        </p>
      )}

      <Button
        className="rounded-sm bg-cta-primary px-4 py-3 font-semibold text-white disabled:opacity-50"
        isDisabled={updateMutation.isPending}
        onPress={() => void onDone()}
      >
        {updateMutation.isPending ? m.common_loading() : m.profile_setup_done()}
      </Button>
    </StepFrame>
  );
}
