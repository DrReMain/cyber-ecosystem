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
    <div className="text-ink flex min-h-svh flex-col">
      <header className="safe-pt safe-px text-ink-tertiary relative z-20 flex items-center justify-between py-4 font-mono text-[10px] tracking-[0.22em]">
        <span className="text-ink font-semibold">{appName}</span>
        {headerTrailing ? <div>{headerTrailing}</div> : null}
      </header>

      <main className="safe-px relative z-10 flex flex-1 flex-col justify-center pb-12">
        <div className="w-full max-w-3xl">
          <span
            className="border-grade-b/40 bg-grade-b/10 text-grade-b inline-flex items-center gap-2.5 rounded-full border px-3.5 py-1.75 font-mono text-[11px] tracking-[0.22em] motion-safe:animate-pulse"
            dir="ltr"
          >
            <span className="bg-grade-b shadow-grade-b size-1.5 rounded-full shadow-[0_0_8px]" />
            PAGE NOT FOUND
          </span>

          <h1
            className="bg-cta-primary mt-7 bg-clip-text font-mono text-[clamp(72px,19vw,160px)] leading-none font-bold tracking-[-0.02em] text-transparent"
            dir="ltr"
          >
            404
          </h1>

          <h2 className="text-[clamp(21px,3.2vw,30px)] font-bold tracking-[0.02em] text-balance rtl:tracking-normal">
            {m.not_found_title()}
          </h2>
          <p className="text-ink-secondary mt-3.5 max-w-136 text-[14.5px] leading-[1.95] text-pretty">
            {m.not_found_description()}
          </p>

          <div className="mt-9 flex flex-wrap gap-3.5">
            <Link
              className="rounded-control bg-cta-primary shadow-cta-primary ease-out-expo inline-flex items-center gap-2 px-7 py-3.25 text-sm font-semibold tracking-[0.04em] text-white transition-transform duration-200 active:scale-95 rtl:tracking-normal"
              to={homePath as never}
            >
              {m.not_found_back_home()}
            </Link>
            <button
              className="rounded-control border-line bg-surface-content text-ink hover:border-ink-tertiary inline-flex cursor-pointer items-center gap-2 border px-7 py-3.25 text-sm font-semibold tracking-[0.04em] backdrop-blur-[6px] transition-colors rtl:tracking-normal"
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
