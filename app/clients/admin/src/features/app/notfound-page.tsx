import { Link, useRouter } from "@tanstack/react-router";
import { Image } from "@unpic/react";
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
    <>
      <div aria-hidden className="fixed inset-0 bg-[#05070d]" />
      <div className="relative mx-auto flex min-h-svh w-full max-w-[1920px] flex-col overflow-x-clip border-white/10 border-x text-[#e8ecf4]">
        <style>{`
@keyframes nc-scan {
  from { transform: translateX(0); }
  to { transform: translateX(350%); }
}
@keyframes nc-scan-rev {
  from { transform: translateX(0); }
  to { transform: translateX(-350%); }
}`}</style>

        <div aria-hidden className="absolute inset-0 rtl:-scale-x-100">
          <div className="absolute inset-0 bg-[url('/assets/nebula.png')] bg-cover bg-position-[72%_center]" />
        </div>
        <div
          aria-hidden
          className="absolute inset-0 bg-linear-to-r from-[#05070d]/90 via-[#05070d]/55 to-[#05070d]/45 rtl:bg-linear-to-l"
        />
        <div
          aria-hidden
          className="absolute inset-0 bg-linear-to-t from-[#05070d]/85 via-[#05070d]/20 to-transparent"
        />

        <header className="relative z-10 flex items-center justify-between px-[clamp(20px,5vw,64px)] py-4.5 font-mono text-[10px] text-white/35 tracking-[0.22em]">
          <span className="inline-flex items-center gap-2.5 font-semibold text-[#e8ecf4]">
            <Image
              alt=""
              aria-hidden
              className="size-4.5"
              height={18}
              src="/mark-accent.svg"
              width={18}
            />
            {appName}
          </span>
          {headerTrailing ? <div className="dark">{headerTrailing}</div> : null}
        </header>

        <main className="relative z-10 flex flex-1 flex-col justify-center px-[clamp(20px,5vw,64px)] pb-12 max-[640px]:justify-start max-[640px]:pt-[16vh]">
          <div className="w-full max-w-3xl">
            <span className="inline-flex items-center gap-2.5 rounded-full border border-primary/40 bg-primary/10 px-3.5 py-1.75 font-mono text-[11px] text-primary tracking-[0.22em] motion-safe:animate-pulse">
              <span className="h-1.5 w-1.5 rounded-full bg-primary shadow-[0_0_8px] shadow-primary" />
              {"PAGE NOT FOUND"}
            </span>

            <h1 className="-ms-1 mt-7 bg-linear-to-br from-4% from-white via-45% via-primary to-90% to-[#8b5cf6] bg-clip-text font-bold font-mono text-[clamp(96px,19vw,224px)] text-transparent leading-none tracking-[-0.02em] drop-shadow-[0_8px_48px] drop-shadow-primary/22 rtl:bg-linear-to-bl">
              404
            </h1>

            <h2 className="text-balance font-bold text-[clamp(21px,3.2vw,30px)] tracking-[0.02em] rtl:tracking-normal">
              {m.not_found_title()}
            </h2>
            <p className="mt-3.5 max-w-136 text-pretty text-[14.5px] text-white/55 leading-[1.95]">
              {m.not_found_description()}
            </p>

            <div className="mt-9 flex flex-wrap gap-3.5">
              <Link
                className="inline-flex items-center gap-2 rounded-xl bg-linear-to-r from-primary to-primary-bright px-7 py-3.25 font-semibold text-sm text-white tracking-[0.04em] shadow-[0_4px_28px] shadow-primary/35 transition-all duration-250 hover:-translate-y-0.5 hover:shadow-[0_8px_36px] hover:shadow-primary/50 rtl:bg-linear-to-l rtl:tracking-normal"
                to={homePath as never}
              >
                {m.not_found_back_home()}
              </Link>
              <button
                className="inline-flex cursor-pointer items-center gap-2 rounded-xl border border-white/10 bg-white/4 px-7 py-3.25 font-semibold text-[#e8ecf4] text-sm tracking-[0.04em] backdrop-blur-[6px] transition-all hover:border-white/40 hover:bg-white/8 rtl:tracking-normal"
                onClick={() => router.history.back()}
                type="button"
              >
                {m.not_found_back_previous()}
              </button>
            </div>

            <div
              aria-hidden
              className="mt-11.5 w-[min(300px,60vw)] font-mono text-[11px] text-white/35 tracking-[0.18em]"
            >
              SEARCHING
              <span className="relative mt-2 block h-0.5 overflow-hidden rounded bg-white/10">
                <span className="absolute inset-s-[-40%] inset-y-0 w-2/5 bg-linear-to-r from-transparent via-primary to-transparent motion-safe:animate-[nc-scan_2.6s_ease-in-out_infinite] rtl:motion-safe:animate-[nc-scan-rev_2.6s_ease-in-out_infinite]" />
              </span>
            </div>
          </div>
        </main>

        <footer className="relative z-10 flex flex-wrap items-center justify-between gap-x-7 gap-y-2.5 border-white/10 border-t bg-[#05070d]/35 px-[clamp(20px,5vw,64px)] py-3.75 font-mono text-[10px] text-white/35 tracking-[0.14em] backdrop-blur" />
      </div>
    </>
  );
}
