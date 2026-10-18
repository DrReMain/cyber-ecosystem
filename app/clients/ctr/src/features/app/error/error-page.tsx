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
    <div className="flex min-h-svh flex-col text-ink">
      <header className="safe-pt safe-px relative z-20 flex items-center justify-between py-4 font-mono text-[10px] text-ink-tertiary tracking-[0.22em]">
        <span className="font-semibold text-ink">{appName}</span>
        {headerTrailing ? <div>{headerTrailing}</div> : null}
      </header>

      <main className="safe-px relative z-10 flex flex-1 flex-col justify-center pb-12">
        <div className="w-full max-w-3xl">
          {resolved.silent ? null : (
            <span
              className="inline-flex items-center gap-2.5 rounded-full border border-danger/40 bg-danger/10 px-3.5 py-1.75 font-mono text-[11px] text-danger-text tracking-[0.22em] motion-safe:animate-pulse"
              dir="ltr"
            >
              <span className="size-1.5 rounded-full bg-danger-text shadow-[0_0_8px] shadow-danger-text" />
              {`ERR · ${resolved.kind.toUpperCase()}`}
            </span>
          )}

          {resolved.silent ? null : (
            <h1
              className="mt-7 bg-cta-primary bg-clip-text font-bold font-mono text-[clamp(72px,19vw,160px)] text-transparent leading-none tracking-[-0.02em]"
              dir="ltr"
            >
              {resolved.http}
            </h1>
          )}

          <h2
            className={clsx(
              "text-balance font-bold text-[clamp(21px,3.2vw,30px)] tracking-[0.02em] rtl:tracking-normal",
              !resolved.silent && "mt-5",
            )}
            role={resolved.silent ? "status" : undefined}
          >
            {resolved.title}
          </h2>

          <div className="mt-9 flex flex-wrap gap-3.5">
            <button
              className="inline-flex cursor-pointer items-center gap-2 rounded-control bg-cta-primary px-7 py-3.25 font-semibold text-sm text-white tracking-[0.04em] shadow-cta-primary transition-transform duration-200 ease-out-expo active:scale-95 rtl:tracking-normal"
              onClick={() => router.invalidate()}
              type="button"
            >
              {m.common_retry()}
            </button>
            <Link
              className="inline-flex items-center gap-2 rounded-control border border-line bg-surface-content px-7 py-3.25 font-semibold text-ink text-sm tracking-[0.04em] backdrop-blur-[6px] transition-colors hover:border-ink-tertiary rtl:tracking-normal"
              to={homePath as never}
            >
              {m.error_home()}
            </Link>

            {extraBtn}
          </div>
        </div>
      </main>

      {resolved.silent ? null : (
        <footer className="safe-px safe-pb relative z-10 flex flex-wrap items-center gap-x-7 gap-y-2.5 border-line border-t py-3.75 font-mono text-[10px] text-ink-tertiary tracking-[0.14em]">
          <span dir="ltr">{`ERR · ${resolved.http} · ${resolved.kind.toUpperCase()}`}</span>
        </footer>
      )}
    </div>
  );
}
