import { Link, useRouter } from "@tanstack/react-router";
import type { ReactNode } from "react";
import { m } from "#/paraglide/messages";

interface NotFoundPageProps {
  appName: string;
  homePath: string;
  headerTrailing?: ReactNode;
}

export function NotFoundPage({ appName, homePath, headerTrailing }: Readonly<NotFoundPageProps>) {
  const router = useRouter();

  return (
    <div className="flex min-h-svh flex-col text-ink">
      <header className="safe-pt safe-px relative z-20 flex items-center justify-between py-4 font-mono text-[10px] text-ink-tertiary tracking-[0.22em]">
        <span className="font-semibold text-ink">{appName}</span>
        {headerTrailing ? <div>{headerTrailing}</div> : null}
      </header>

      <main className="safe-px relative z-10 flex flex-1 flex-col justify-center pb-12">
        <div className="w-full max-w-3xl">
          <span
            className="inline-flex items-center gap-2.5 rounded-full border border-grade-b/40 bg-grade-b/10 px-3.5 py-1.75 font-mono text-[11px] text-grade-b tracking-[0.22em] motion-safe:animate-pulse"
            dir="ltr"
          >
            <span className="size-1.5 rounded-full bg-grade-b shadow-[0_0_8px] shadow-grade-b" />
            PAGE NOT FOUND
          </span>

          <h1
            className="mt-7 bg-cta-primary bg-clip-text font-bold font-mono text-[clamp(72px,19vw,160px)] text-transparent leading-none tracking-[-0.02em]"
            dir="ltr"
          >
            404
          </h1>

          <h2 className="text-balance font-bold text-[clamp(21px,3.2vw,30px)] tracking-[0.02em] rtl:tracking-normal">
            {m.not_found_title()}
          </h2>
          <p className="mt-3.5 max-w-136 text-pretty text-[14.5px] text-ink-secondary leading-[1.95]">
            {m.not_found_description()}
          </p>

          <div className="mt-9 flex flex-wrap gap-3.5">
            <Link
              className="inline-flex items-center gap-2 rounded-control bg-cta-primary px-7 py-3.25 font-semibold text-sm text-white tracking-[0.04em] shadow-cta-primary transition-transform duration-200 ease-out-expo active:scale-95 rtl:tracking-normal"
              to={homePath as never}
            >
              {m.not_found_back_home()}
            </Link>
            <button
              className="inline-flex cursor-pointer items-center gap-2 rounded-control border border-line bg-surface-content px-7 py-3.25 font-semibold text-ink text-sm tracking-[0.04em] backdrop-blur-[6px] transition-colors hover:border-ink-tertiary rtl:tracking-normal"
              onClick={() => router.history.back()}
              type="button"
            >
              {m.not_found_back_previous()}
            </button>
          </div>
        </div>
      </main>
    </div>
  );
}
