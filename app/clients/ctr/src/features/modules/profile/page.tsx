import { useMutation, useQuery } from "@connectrpc/connect-query";
import type { Game } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/game_pb";
import { listGames } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/game-GameService_connectquery";
import type { Profile as ProfilePB } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/profile_pb";
import {
  getProfile,
  updateProfile,
} from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/profile-ProfileService_connectquery";
import { useNavigate } from "@tanstack/react-router";
import { useEffect, useState } from "react";
import {
  Button,
  Checkbox,
  CheckboxGroup,
  Form,
  Input,
  Label,
  TextField,
} from "react-aria-components";
import { normalize } from "#/domains/error";
import { m } from "#/paraglide/messages";
import { logoutFn } from "./logout.fn";

const YEARS = Array.from({ length: 87 }, (_, i) => 2026 - i);
const MONTHS = Array.from({ length: 12 }, (_, i) => i + 1);
interface BirthPair {
  month: string;
  year: string;
}
interface CustomName {
  id: string;
  name: string;
}

// Client-side mirror of the server's under-13 gate — a UX affordance only;
// the server re-derives and refuses regardless of what the client shows.
function under13(pair: BirthPair): boolean {
  if (pair.year === "" || pair.month === "") {
    return false;
  }
  const now = new Date();
  let age = now.getFullYear() - Number(pair.year);
  if (now.getMonth() + 1 < Number(pair.month)) {
    age--;
  }
  return age < 13;
}

function boxClass(selected: boolean): string {
  return `h-4 w-4 shrink-0 rounded-sm border ${selected ? "border-focus bg-focus/40" : "border-line"}`;
}

export function ProfilePage() {
  const navigate = useNavigate();
  const profileQuery = useQuery(getProfile, {});
  const gamesQuery = useQuery(listGames, {});
  const updateMutation = useMutation(updateProfile);
  const [editing, setEditing] = useState(false);
  const [loggingOut, setLoggingOut] = useState(false);
  const [pending, setPending] = useState(false);
  const [birthPair, setBirthPair] = useState<BirthPair>({ month: "", year: "" });
  const [customNames, setCustomNames] = useState<CustomName[]>([]);

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

  function startEditing() {
    setBirthPair({
      month: profile?.birthMonth != null ? String(profile.birthMonth) : "",
      year: profile?.birthYear != null ? String(profile.birthYear) : "",
    });
    setCustomNames(
      (profile?.games ?? [])
        .map((g) => g.customName)
        .filter((name): name is string => name != null)
        .map((name) => ({ id: crypto.randomUUID(), name })),
    );
    setEditing(true);
  }

  async function onLogout() {
    setLoggingOut(true);
    try {
      await logoutFn();
      await navigate({ to: "/login" });
    } finally {
      setLoggingOut(false);
    }
  }

  async function onSave(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const form = new FormData(event.currentTarget);
    setPending(true);
    try {
      await updateMutation.mutateAsync({
        nickname: String(form.get("nickname") ?? ""),
        ...(birthPair.year !== "" && birthPair.month !== ""
          ? { birthYear: Number(birthPair.year), birthMonth: Number(birthPair.month) }
          : {}),
        ...(under13(birthPair) ? { guardianAccepted: form.get("guardian") === "on" } : {}),
        gamePreference: {
          customNames: customNames.map((entry) => entry.name),
          gameIds: form.getAll("game_ids").map(String),
        },
      });
      setEditing(false);
      await profileQuery.refetch();
    } finally {
      setPending(false);
    }
  }

  return (
    <main className="safe-px safe-pt flex min-h-app flex-col gap-4 py-6">
      <div className="flex items-center justify-between">
        <h1 className="font-bold text-ink text-xl">{m.profile_title()}</h1>
        <div className="flex items-center gap-3">
          {!editing && (
            <Button
              className="rounded-sm border border-line px-3 py-1.5 text-ink-secondary text-sm"
              onPress={startEditing}
            >
              {m.profile_edit()}
            </Button>
          )}
          <Button
            className="text-ink-tertiary text-sm underline disabled:opacity-50"
            isDisabled={loggingOut}
            onPress={onLogout}
          >
            {loggingOut ? m.common_loading() : m.profile_logout()}
          </Button>
        </div>
      </div>
      {editing ? (
        <ProfileEditForm
          birthPair={birthPair}
          customNames={customNames}
          games={gamesQuery.data?.games ?? []}
          onCancel={() => setEditing(false)}
          onSubmit={onSave}
          pending={pending}
          profile={profile}
          setBirthPair={setBirthPair}
          setCustomNames={setCustomNames}
        />
      ) : (
        <ProfileDisplay profile={profile} />
      )}
    </main>
  );
}

interface ProfileEditFormProps {
  birthPair: BirthPair;
  customNames: CustomName[];
  games: Game[];
  onCancel: () => void;
  onSubmit: (event: React.FormEvent<HTMLFormElement>) => void;
  pending: boolean;
  profile: ProfilePB | undefined;
  setBirthPair: React.Dispatch<React.SetStateAction<BirthPair>>;
  setCustomNames: React.Dispatch<React.SetStateAction<CustomName[]>>;
}

function ProfileEditForm({
  birthPair,
  customNames,
  games,
  onCancel,
  onSubmit,
  pending,
  profile,
  setBirthPair,
  setCustomNames,
}: Readonly<ProfileEditFormProps>) {
  return (
    <Form
      className="flex w-full max-w-sm flex-col gap-4 rounded-lg border border-line bg-surface-card p-5"
      onSubmit={onSubmit}
    >
      <TextField defaultValue={profile?.nickname ?? ""} isRequired name="nickname">
        <Label className="mb-1 block text-ink-secondary text-sm">
          {m.profile_nickname_label()}
        </Label>
        <Input className="w-full rounded-sm border border-line bg-surface-input px-3 py-2.5 text-ink outline-none focus:border-focus" />
      </TextField>
      <div className="flex gap-3">
        <label className="flex-1">
          <span className="mb-1 block text-ink-secondary text-sm">{m.profile_birth_year()}</span>
          <select
            className="w-full rounded-sm border border-line bg-surface-input px-3 py-2.5 text-ink outline-none focus:border-focus"
            name="birth_year"
            onChange={(event) => setBirthPair((prev) => ({ ...prev, year: event.target.value }))}
            value={birthPair.year}
          >
            <option value="">--</option>
            {YEARS.map((y) => (
              <option key={y} value={y}>
                {y}
              </option>
            ))}
          </select>
        </label>
        <label className="w-28">
          <span className="mb-1 block text-ink-secondary text-sm">{m.profile_birth_month()}</span>
          <select
            className="w-full rounded-sm border border-line bg-surface-input px-3 py-2.5 text-ink outline-none focus:border-focus"
            name="birth_month"
            onChange={(event) => setBirthPair((prev) => ({ ...prev, month: event.target.value }))}
            value={birthPair.month}
          >
            <option value="">--</option>
            {MONTHS.map((mo) => (
              <option key={mo} value={mo}>
                {mo}
              </option>
            ))}
          </select>
        </label>
      </div>
      {under13(birthPair) && (
        <Checkbox name="guardian">
          {({ isSelected }) => (
            <div className="flex items-center gap-2">
              <span className={boxClass(isSelected)} />
              <span className="text-ink-secondary text-sm">{m.profile_guardian()}</span>
            </div>
          )}
        </Checkbox>
      )}
      <div>
        <span className="mb-1 block text-ink-secondary text-sm">{m.profile_games_label()}</span>
        <CheckboxGroup
          defaultValue={(profile?.games ?? [])
            .map((g) => g.gameId)
            .filter((id): id is string => id != null)}
          name="game_ids"
        >
          <div className="grid grid-cols-2 gap-2">
            {games.map((g) => (
              <Checkbox key={g.id} value={g.id ?? ""}>
                {({ isSelected }) => (
                  <div className="flex items-center gap-2">
                    <span className={boxClass(isSelected)} />
                    <span className="text-ink-secondary text-sm">{g.name}</span>
                  </div>
                )}
              </Checkbox>
            ))}
          </div>
        </CheckboxGroup>
      </div>
      <div className="flex flex-col gap-2">
        <input
          className="w-full rounded-sm border border-line bg-surface-input px-3 py-2 text-ink text-sm outline-none focus:border-focus"
          onKeyDown={(event) => {
            if (event.key !== "Enter") {
              return;
            }
            event.preventDefault();
            const input = event.currentTarget;
            const value = input.value.trim();
            if (value !== "" && customNames.length < 20) {
              setCustomNames((prev) => [...prev, { id: crypto.randomUUID(), name: value }]);
              input.value = "";
            }
          }}
          placeholder={m.profile_custom_placeholder()}
          type="text"
        />
        {customNames.length > 0 && (
          <div className="flex flex-wrap gap-2">
            {customNames.map((entry) => (
              <button
                className="rounded-full border border-line px-3 py-1 text-ink-secondary text-xs"
                key={entry.id}
                onClick={() => setCustomNames((prev) => prev.filter((e) => e.id !== entry.id))}
                type="button"
              >
                {entry.name} ×
              </button>
            ))}
          </div>
        )}
      </div>
      <div className="flex gap-3">
        <Button
          className="flex-1 rounded-sm bg-linear-to-r from-[#2563eb] to-[#06b6d4] px-4 py-3 font-semibold text-white disabled:opacity-50"
          isDisabled={pending}
          type="submit"
        >
          {pending ? m.common_loading() : m.profile_save()}
        </Button>
        <Button
          className="rounded-sm border border-line px-4 py-3 text-ink-secondary"
          onPress={onCancel}
        >
          {m.profile_cancel()}
        </Button>
      </div>
    </Form>
  );
}

function ProfileDisplay({ profile }: Readonly<{ profile: ProfilePB | undefined }>) {
  const unset = m.profile_unset();
  const birth =
    profile?.birthYear != null && profile?.birthMonth != null
      ? `${profile.birthYear} · ${profile.birthMonth}`
      : unset;
  const gamesCount =
    profile?.games !== undefined && profile.games.length > 0 ? String(profile.games.length) : unset;
  const rows: [string, string][] = [
    [m.profile_id_label(), profile?.id ?? unset],
    [m.profile_nickname_label(), profile?.nickname ?? unset],
    [m.profile_birth_label(), birth],
    [m.profile_band_label(), profile?.ageBand ?? unset],
    [m.profile_games_label(), gamesCount],
  ];
  return (
    <dl className="w-full max-w-sm rounded-lg border border-line bg-surface-card p-5">
      {rows.map(([label, value], index) => (
        <div
          className={`flex justify-between py-2.5 ${index < rows.length - 1 ? "border-line-soft border-b" : ""}`}
          key={label}
        >
          <dt className="text-ink-secondary text-sm">{label}</dt>
          <dd
            className={`text-sm ${label === m.profile_id_label() ? "font-mono text-ink" : "text-ink"}`}
          >
            {value}
          </dd>
        </div>
      ))}
    </dl>
  );
}
