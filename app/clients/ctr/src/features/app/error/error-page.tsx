import { Link, useRouter } from "@tanstack/react-router";
import clsx from "clsx";
import type { ReactNode } from "react";
import { view } from "#/domains/error";
import { m } from "#/paraglide/messages";

interface ErrorPageProps {
  appName: string;
  homePath: string;
  headerTrailing?: ReactNode;
  error: unknown;
  extraBtn?: ReactNode;
}

export function ErrorPage({
  appName,
  homePath,
  headerTrailing,
  error,
  extraBtn = null,
}: Readonly<ErrorPageProps>) {
  const router = useRouter();
  const resolved = view(error);

  return (
    <div className="text-ink flex min-h-svh flex-col">
      <header className="safe-pt safe-px text-ink-tertiary relative z-20 flex items-center justify-between py-4 font-mono text-[10px] tracking-[0.22em]">
        <span className="text-ink font-semibold">{appName}</span>
        {headerTrailing ? <div>{headerTrailing}</div> : null}
      </header>

      <main className="safe-px relative z-10 flex flex-1 flex-col justify-center pb-12">
        <div className="w-full max-w-3xl">
          {resolved.silent ? null : (
            <span
              className="border-danger/40 bg-danger/10 text-danger-text inline-flex items-center gap-2.5 rounded-full border px-3.5 py-1.75 font-mono text-[11px] tracking-[0.22em] motion-safe:animate-pulse"
              dir="ltr"
            >
              <span className="bg-danger-text shadow-danger-text size-1.5 rounded-full shadow-[0_0_8px]" />
              {`ERR · ${resolved.kind.toUpperCase()}`}
            </span>
          )}

          {resolved.silent ? null : (
            <h1
              className="bg-cta-primary mt-7 bg-clip-text font-mono text-[clamp(72px,19vw,160px)] leading-none font-bold tracking-[-0.02em] text-transparent"
              dir="ltr"
            >
              {resolved.http}
            </h1>
          )}

          <h2
            className={clsx(
              "text-[clamp(21px,3.2vw,30px)] font-bold tracking-[0.02em] text-balance rtl:tracking-normal",
              !resolved.silent && "mt-5",
            )}
            role={resolved.silent ? "status" : undefined}
          >
            {resolved.title}
          </h2>

          <div className="mt-9 flex flex-wrap gap-3.5">
            <button
              className="rounded-control bg-cta-primary shadow-cta-primary ease-out-expo inline-flex cursor-pointer items-center gap-2 px-7 py-3.25 text-sm font-semibold tracking-[0.04em] text-white transition-transform duration-200 active:scale-95 rtl:tracking-normal"
              onClick={() => router.invalidate()}
              type="button"
            >
              {m.common_retry()}
            </button>
            <Link
              className="rounded-control border-line bg-surface-content text-ink hover:border-ink-tertiary inline-flex items-center gap-2 border px-7 py-3.25 text-sm font-semibold tracking-[0.04em] backdrop-blur-[6px] transition-colors rtl:tracking-normal"
              to={homePath as never}
            >
              {m.error_home()}
            </Link>

            {extraBtn}
          </div>
        </div>
      </main>

      {resolved.silent ? null : (
        <footer className="safe-px safe-pb border-line text-ink-tertiary relative z-10 flex flex-wrap items-center gap-x-7 gap-y-2.5 border-t py-3.75 font-mono text-[10px] tracking-[0.14em]">
          <span dir="ltr">{`ERR · ${resolved.http} · ${resolved.kind.toUpperCase()}`}</span>
        </footer>
      )}
    </div>
  );
}
