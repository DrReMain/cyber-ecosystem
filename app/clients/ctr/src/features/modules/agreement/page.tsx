import { ChevronLeft } from "lucide-react";
import { AmbientLayout } from "#/features/app/ambient/ambient-layout";
import { m } from "#/paraglide/messages";
import { AGREEMENT_DOC_META, type AgreementDoc } from "./versions";

interface AgreementPageProps {
  doc: AgreementDoc;
}

/**
 * The legal face (prototype agreement.html): one page component, two static
 * routes — /agreement/user and /agreement/privacy each pass their document.
 * Chapter copy is the prototype's demo placeholder until ljr's final legal
 * text lands; the version line reads versions.ts.
 */
export function AgreementPage({ doc }: Readonly<AgreementPageProps>) {
  const meta = AGREEMENT_DOC_META[doc];
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

  return (
    <AmbientLayout>
      <div className="flex w-full flex-col gap-3 pb-6">
        {/* Inline header — match's FlowHeader is another module's territory;
            same shape (the §4.7 fab: 40px, borderless) declared here. */}
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
          {m.agreement_version_line({ date: meta.effectiveDate, no: meta.no })}
          {" · "}
          {m.agreement_demo_note()}
        </div>

        {chapters.map((chapter) => (
          <div className="border-line bg-surface-card rounded-card border p-4" key={chapter.key}>
            <div className="text-ink mb-1.5 text-[13.5px] font-extrabold">{chapter.title}</div>
            <div className="text-ink-tertiary text-[12.5px] leading-[1.8]">{chapter.body}</div>
          </div>
        ))}

        <div className="border-line bg-surface-deep text-ink-tertiary rounded-control border p-3.5 text-[11.5px] leading-relaxed">
          {m.agreement_consent_note()}
        </div>
      </div>
    </AmbientLayout>
  );
}
