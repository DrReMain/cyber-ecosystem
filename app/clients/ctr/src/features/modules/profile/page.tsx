import { useQuery } from "@connectrpc/connect-query";
import { getProfile } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/profile-ProfileService_connectquery";
import { useNavigate } from "@tanstack/react-router";
import { useEffect } from "react";
import { normalize } from "#/domains/error";
import { m } from "#/paraglide/messages";

// The browser↔esports leg rides this page: the query fires from the browser
// with the HttpOnly cookie attached same-origin, so a rendered member id
// proves the whole chain (cookie → introspection → subject → response).
export function ProfilePage() {
  const navigate = useNavigate();
  const profileQuery = useQuery(getProfile, {});

  const unauthenticated =
    profileQuery.error !== null &&
    normalize(profileQuery.error).reason === "GENERAL_ERROR_UNAUTHENTICATED";

  useEffect(() => {
    if (unauthenticated) {
      void navigate({ to: "/login", search: { next: "/profile" } });
    }
  }, [unauthenticated, navigate]);

  if (profileQuery.isPending) {
    return (
      <main className="safe-px safe-pt flex min-h-app items-center justify-center">
        <p className="text-ink-tertiary">{m.common_loading()}</p>
      </main>
    );
  }

  if (profileQuery.error !== null && !unauthenticated) {
    return (
      <main className="safe-px safe-pt flex min-h-app flex-col items-center justify-center gap-3">
        <p className="text-danger-text">{m.profile_error_generic()}</p>
      </main>
    );
  }

  const profile = profileQuery.data?.profile ?? undefined;
  const unset = m.profile_unset();
  const birth =
    profile?.birthYear != null && profile?.birthMonth != null
      ? `${profile.birthYear} · ${profile.birthMonth}`
      : unset;
  const gamesCount =
    profile?.games !== undefined && profile.games.length > 0 ? String(profile.games.length) : unset;

  return (
    <main className="safe-px safe-pt flex min-h-app flex-col gap-4 py-6">
      <h1 className="font-bold text-ink text-xl">{m.profile_title()}</h1>
      <dl className="w-full max-w-sm rounded-lg border border-line bg-surface-card p-5">
        <div className="flex justify-between border-line-soft border-b py-2.5">
          <dt className="text-ink-secondary text-sm">{m.profile_id_label()}</dt>
          <dd className="font-mono text-ink text-sm">{profile?.id ?? unset}</dd>
        </div>
        <div className="flex justify-between border-line-soft border-b py-2.5">
          <dt className="text-ink-secondary text-sm">{m.profile_nickname_label()}</dt>
          <dd className="text-ink text-sm">{profile?.nickname ?? unset}</dd>
        </div>
        <div className="flex justify-between border-line-soft border-b py-2.5">
          <dt className="text-ink-secondary text-sm">{m.profile_birth_label()}</dt>
          <dd className="text-ink text-sm">{birth}</dd>
        </div>
        <div className="flex justify-between border-line-soft border-b py-2.5">
          <dt className="text-ink-secondary text-sm">{m.profile_band_label()}</dt>
          <dd className="text-ink text-sm">{profile?.ageBand ?? unset}</dd>
        </div>
        <div className="flex justify-between py-2.5">
          <dt className="text-ink-secondary text-sm">{m.profile_games_label()}</dt>
          <dd className="text-ink text-sm">{gamesCount}</dd>
        </div>
      </dl>
    </main>
  );
}
