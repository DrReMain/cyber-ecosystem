import { useNavigate } from "@tanstack/react-router";
import { Check } from "lucide-react";
import { AmbientLayout } from "#/features/app/ambient/ambient-layout";
import { m } from "#/paraglide/messages";

// The account-merge closing screen: it says what moved and where to check
// it — static copy, no state of its own.
export function MergeSuccessPage() {
  const navigate = useNavigate();
  return (
    <AmbientLayout>
      <div className="flex w-full flex-col items-center gap-4 pt-14">
        <span className="bg-cta-success shadow-cta-success flex h-[72px] w-[72px] items-center justify-center rounded-full">
          <Check aria-hidden className="text-white" size={34} strokeWidth={3} />
        </span>
        <div className="text-center">
          <div className="text-ink text-[21px] font-extrabold">{m.security_merge_success_title()}</div>
          <p className="text-ink-secondary mt-1.5 text-center text-[12px] leading-[1.7]">
            {m.security_merge_success_body()}
          </p>
        </div>
        <div className="border-line-accent-blue bg-identity rounded-card w-full border px-4 py-2">
          <CheckRow text={m.security_merge_success_line_credential()} />
          <CheckRow text={m.security_merge_success_line_history()} />
        </div>
        <button
          className="bg-cta-primary shadow-cta-primary rounded-control w-full px-4 py-3.5 text-[15px] font-bold text-white active:opacity-60"
          onClick={() => void navigate({ to: "/profile", viewTransition: false })}
          type="button"
        >
          {m.security_merge_success_cta()}
        </button>
        <div className="text-ink-tertiary text-center text-[11px]">{m.security_merge_success_note()}</div>
      </div>
    </AmbientLayout>
  );
}

function CheckRow({ text }: { text: string }) {
  return (
    <div className="flex items-center gap-2.5 py-[11px]">
      <span className="flex h-[22px] w-[22px] shrink-0 items-center justify-center rounded-full bg-[rgba(34,197,94,0.2)]">
        <Check aria-hidden className="text-[#4ADE80]" size={12} strokeWidth={3.5} />
      </span>
      <span className="text-ink-secondary text-[13px]">{text}</span>
    </div>
  );
}
