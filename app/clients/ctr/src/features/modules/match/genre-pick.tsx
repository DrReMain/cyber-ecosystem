import { useNavigate, useSearch } from "@tanstack/react-router";
import { ChevronRight, FolderOpen } from "lucide-react";
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
          <div className="border-line bg-surface-card rounded-lg border p-5 text-center">
            <FolderOpen
              aria-hidden
              className="text-ink-tertiary mx-auto"
              size={30}
              strokeWidth={1.5}
            />
            <div className="text-ink mt-2 text-sm font-extrabold">{m.match_pick_empty_title()}</div>
            <p className="text-ink-tertiary mt-2 text-[12.5px] leading-relaxed">
              {m.match_pick_empty_body()}
            </p>
          </div>
          <button
            className="bg-cta-primary w-full rounded-sm px-4 py-3 font-semibold text-white active:opacity-80"
            onClick={() => void navigate({ to: "/match/pick" })}
            type="button"
          >
            {m.match_pick_empty_retry()}
          </button>
          <button
            className="border-line bg-surface-content text-ink w-full rounded-sm border px-4 py-3 font-semibold active:opacity-60"
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
          <div className="border-status-pending/45 bg-surface-deep text-status-pending rounded-lg border p-3.5 text-[12px] leading-relaxed">
            {m.match_pick_tie_head()}
          </div>
        ) : (
          <div className="border-line bg-surface-deep text-ink-tertiary rounded-lg border p-3.5 text-[12px] leading-relaxed">
            {m.match_pick_normal_head()}
          </div>
        )}
        {(["fps", "moba"] as const).map((genre) => (
          <button
            className="border-line bg-surface-card rounded-lg border p-4 text-left active:opacity-60"
            key={genre}
            onClick={() => choose(genre)}
            type="button"
          >
            <div className="flex items-center">
              <div className="flex-1">
                <div className="text-ink text-[15px] font-extrabold">
                  {genre === "fps" ? m.match_pick_fps() : m.match_pick_moba()}
                </div>
                <div className="text-ink-tertiary mt-0.5 text-[11px]">{m.match_pick_sub()}</div>
              </div>
              <ChevronRight aria-hidden className="text-ink-tertiary" size={16} />
            </div>
            <div className="mt-1.5">
              {GENRE_ORDER[genre].map((ability) => {
                const word = abilityWord(ability);
                return (
                  <div className="mt-1 flex items-center gap-2" key={ability}>
                    <span
                      className={`${word.gradient} flex h-5 w-5 items-center justify-center rounded-sm text-[9px] font-bold text-white`}
                    >
                      {word.axis}
                    </span>
                    <span className="text-ink-secondary text-[12px]">{word.cn}</span>
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
