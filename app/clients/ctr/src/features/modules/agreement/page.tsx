import { useQuery } from "@connectrpc/connect-query";
import { getAgreementCurrent } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/agreement-AgreementService_connectquery";
import { ChevronLeft } from "lucide-react";
import { AmbientLayout } from "#/features/app/ambient/ambient-layout";
import { localDateOf } from "#/features/app/date";
import { m } from "#/paraglide/messages";

export type AgreementDoc = "user" | "privacy";

interface AgreementPageProps {
  doc: AgreementDoc;
}

// The wire carries full value names while the generated type says enum —
// String() bridges the two (the transport JSON family's standing shape).
const wireDoc = (doc: AgreementDoc): string =>
  doc === "privacy" ? "AGREEMENT_DOC_PRIVACY" : "AGREEMENT_DOC_USER";

// Timestamps arrive JSON-shaped and the transport family renders them
// differently ({seconds} objects or RFC3339 strings — the accepted WKT
// variance); normalize every shape, the line shows the date only.
const wireDate = (at: unknown): string => {
  const ms = dateLikeMs(at);
  return ms === null ? "" : localDateOf(ms);
};

const dateLikeMs = (v: unknown): number | null => {
  if (typeof v === "string") {
    const parsed = new Date(v).getTime();
    // NaN check, not falsiness: the epoch itself parses to 0.
    return Number.isNaN(parsed) ? null : parsed;
  }
  if (v != null && typeof v === "object") {
    const seconds = Number((v as { seconds?: unknown }).seconds ?? Number.NaN);
    return Number.isNaN(seconds) ? null : seconds * 1000;
  }
  return null;
};

/**
 * The legal face (prototype agreement.html): one page component, two static
 * routes — /agreement/user and /agreement/privacy each pass their document.
 * Chapter copy is the prototype's demo placeholder until ljr's final legal
 * text lands; the version line reads the server's current version — the
 * operator's publish is the only truth, a local versions file is not.
 */
export function AgreementPage({ doc }: Readonly<AgreementPageProps>) {
  const title = doc === "privacy" ? m.agreement_privacy_title() : m.agreement_user_title();
  const chapters =
    doc === "privacy"
      ? [
          { body: m.agreement_privacy_c1_d(), key: "p1", title: m.agreement_privacy_c1_t() },
          { body: m.agreement_privacy_c2_d(), key: "p2", title: m.agreement_privacy_c2_t() },
          { body: m.agreement_privacy_c3_d(), key: "p3", title: m.agreement_privacy_c3_t() },
          { body: m.agreement_privacy_c4_d(), key: "p4", title: m.agreement_privacy_c4_t() },
          { body: m.agreement_privacy_c5_d(), key: "p5", title: m.agreement_privacy_c5_t() },
          { body: m.agreement_privacy_c6_d(), key: "p6", title: m.agreement_privacy_c6_t() },
        ]
      : [
          { body: m.agreement_user_c1_d(), key: "u1", title: m.agreement_user_c1_t() },
          { body: m.agreement_user_c2_d(), key: "u2", title: m.agreement_user_c2_t() },
          { body: m.agreement_user_c3_d(), key: "u3", title: m.agreement_user_c3_t() },
          { body: m.agreement_user_c4_d(), key: "u4", title: m.agreement_user_c4_t() },
          { body: m.agreement_user_c5_d(), key: "u5", title: m.agreement_user_c5_t() },
          { body: m.agreement_user_c6_d(), key: "u6", title: m.agreement_user_c6_t() },
        ];

  // The route loader prefetched this; no current version, no version line —
  // a legal page never guesses its own version.
  const current = useQuery(getAgreementCurrent, {});
  const meta = (current.data?.docs ?? []).find((d) => String(d.doc) === wireDoc(doc));

  return (
    <AmbientLayout>
      <div className="flex w-full flex-col gap-3 pb-6">
        {/* Inline header — match's FlowHeader is another module's territory;
            same shape (the prototype's fab: 40px, borderless) declared here. */}
        <div className="flex items-center gap-3">
          <button
            aria-label={m.common_back()}
            className="bg-surface-input text-ink-secondary flex h-10 w-10 shrink-0 items-center justify-center rounded-full active:opacity-60"
            onClick={() => void history.back()}
            type="button"
          >
            <ChevronLeft aria-hidden size={20} />
          </button>
          <h1 className="text-ink text-[19px] font-bold">{title}</h1>
        </div>

        <div className="text-ink-tertiary text-[11.5px]">
          {meta !== undefined && (
            <>
              {m.agreement_version_line({
                date: wireDate(meta.effectiveAt),
                no: meta.no ?? "",
              })}
              {" · "}
            </>
          )}
          {m.agreement_demo_note()}
        </div>

        {chapters.map((chapter) => (
          <div
            className="border-line bg-surface-card shadow-card rounded-card border px-4 py-3.5 backdrop-blur-md"
            key={chapter.key}
          >
            <div className="text-ink mb-1.5 text-[13.5px] font-extrabold">{chapter.title}</div>
            <div className="text-ink-muted text-[12.5px] leading-[1.8]">{chapter.body}</div>
          </div>
        ))}

        <div className="border-line bg-surface-content text-ink-muted rounded-control border px-3.5 py-3 text-[11.5px] leading-[1.7]">
          {m.agreement_consent_note()}
        </div>
      </div>
    </AmbientLayout>
  );
}
