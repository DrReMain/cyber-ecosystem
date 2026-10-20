import { useNavigate, useSearch } from "@tanstack/react-router";
import { FolderOpen } from "lucide-react";
import { toast } from "sonner";
import { AmbientLayout } from "#/features/app/ambient/ambient-layout";
import { m } from "#/paraglide/messages";
import { GENRE_ORDER, type Genre } from "./arranger";
import { FlowHeader } from "./ui/flow-header";
import { abilityWord, genreLabel } from "./vocabulary";

/**
 * The manual genre pick (prototype match-genre-pick): tie face, plain
 * face, or the empty-pool face (?why=empty). Choosing rebuilds the
 * lineup via the genre URL param and returns to the guide.
 */
export function GenrePickPage() {
  const navigate = useNavigate();
  const search = useSearch({ from: "/match/pick" }) as { why?: "tie" | "empty" };
  const empty = search.why === "empty";

  function choose(genre: Genre) {
    toast.success(m.match_pick_toast({ type: genreLabel(genre) }));
    void navigate({ to: "/match", search: { genre } });
  }

  if (empty) {
    return (
      <AmbientLayout>
        <div className="flex w-full flex-col gap-3 pb-6">
          <FlowHeader
            onBack={() => void navigate({ to: "/" })}
            title={m.match_pick_empty_title()}
          />
          <div className="rounded-lg border border-line bg-surface-card p-5 text-center">
            <FolderOpen
              aria-hidden
              className="mx-auto text-ink-tertiary"
              size={30}
              strokeWidth={1.5}
            />
            <div className="mt-2 font-extrabold text-ink text-sm">{m.match_pick_empty_title()}</div>
            <p className="mt-2 text-[12.5px] text-ink-tertiary leading-relaxed">
              {m.match_pick_empty_body()}
            </p>
          </div>
          <button
            className="w-full rounded-sm bg-cta-primary px-4 py-3 font-semibold text-white active:opacity-80"
            onClick={() => void navigate({ to: "/match/pick" })}
            type="button"
          >
            {m.match_pick_empty_retry()}
          </button>
          <button
            className="w-full rounded-sm border border-line bg-surface-content px-4 py-3 font-semibold text-ink active:opacity-60"
            onClick={() => void navigate({ to: "/" })}
            type="button"
          >
            {m.match_pick_empty_home()}
          </button>
        </div>
      </AmbientLayout>
    );
  }

  return (
    <AmbientLayout>
      <div className="flex w-full flex-col gap-3 pb-6">
        <FlowHeader onBack={() => void navigate({ to: "/" })} title={m.match_pick_title()} />
        {search.why === "tie" ? (
          <div className="rounded-lg border border-status-pending/45 bg-surface-deep p-3.5 text-[12px] text-status-pending leading-relaxed">
            {m.match_pick_tie_head()}
          </div>
        ) : (
          <div className="rounded-lg border border-line bg-surface-deep p-3.5 text-[12px] text-ink-tertiary leading-relaxed">
            {m.match_pick_normal_head()}
          </div>
        )}
        {(["fps", "moba"] as const).map((genre) => (
          <button
            className="rounded-lg border border-line bg-surface-card p-4 text-left active:opacity-60"
            key={genre}
            onClick={() => choose(genre)}
            type="button"
          >
            <div className="flex items-center">
              <div className="flex-1">
                <div className="font-extrabold text-[15px] text-ink">
                  {genre === "fps" ? m.match_pick_fps() : m.match_pick_moba()}
                </div>
                <div className="mt-0.5 text-[11px] text-ink-tertiary">{m.match_pick_sub()}</div>
              </div>
              <span aria-hidden className="text-ink-tertiary">
                ›
              </span>
            </div>
            <div className="mt-1.5">
              {GENRE_ORDER[genre].map((ability) => {
                const word = abilityWord(ability);
                return (
                  <div className="mt-1 flex items-center gap-2" key={ability}>
                    <span
                      className={`${word.gradient} flex h-5 w-5 items-center justify-center rounded-sm font-bold text-[9px] text-white`}
                    >
                      {word.axis}
                    </span>
                    <span className="text-[12px] text-ink-secondary">{word.cn}</span>
                  </div>
                );
              })}
            </div>
          </button>
        ))}
      </div>
    </AmbientLayout>
  );
}
