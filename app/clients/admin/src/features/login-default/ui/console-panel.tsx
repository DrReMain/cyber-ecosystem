import { ThemeToggle } from "@cyber-ecosystem/shared-theme";
import { Moon, Sun } from "lucide-react";
import type { ReactNode } from "react";
import { LocaleSelect } from "#/domains/i18n";
import { m } from "#/paraglide/messages";

export function ConsolePanel({ children }: { children?: ReactNode }) {
  return (
    <section className="relative flex min-w-0 flex-1 flex-col border-s border-black/10 bg-white text-black/88 max-[899px]:mt-auto max-[899px]:flex-none max-[899px]:rounded-t-2xl max-[899px]:border max-[899px]:backdrop-blur-[14px] dark:border-white/10 dark:bg-[#070b15] dark:text-[#e8ecf4]">
      <div
        aria-hidden
        className="from-primary/5 pointer-events-none absolute inset-0 hidden bg-radial-[120%_50%_at_100%_0%] from-0% to-transparent to-55% rtl:bg-radial-[120%_50%_at_0%_0%] min-[900px]:dark:block"
      />
      <div
        aria-hidden
        className="pointer-events-none absolute inset-0 hidden bg-linear-to-b from-[#e8ecf4]/3 to-[#e8ecf4]/[0.012] min-[900px]:dark:block"
      />
      <div
        aria-hidden
        className="from-primary/4 pointer-events-none absolute inset-0 hidden bg-radial-[120%_50%_at_100%_0%] from-0% to-transparent to-55% min-[900px]:block rtl:bg-radial-[120%_50%_at_0%_0%] min-[900px]:dark:hidden"
      />
      <div
        aria-hidden
        className="pointer-events-none absolute inset-0 hidden bg-[linear-gradient(180deg,rgba(10,15,26,0.92),rgba(7,11,21,0.97))] max-[899px]:dark:block"
      />
      <div
        aria-hidden
        className="pointer-events-none absolute inset-0 hidden bg-[linear-gradient(180deg,rgba(255,255,255,0.94),rgba(248,250,252,0.98))] max-[899px]:block max-[899px]:dark:hidden"
      />

      <header className="relative z-20 flex items-center justify-end gap-2 px-[clamp(24px,3.5vw,44px)] pt-5.5 font-mono text-[10px] tracking-[0.22em] text-black/35 dark:text-white/35">
        <ThemeToggle>
          {({ isDark, toggle }) => (
            <button
              aria-label={isDark ? "Switch to light mode" : "Switch to dark mode"}
              className="focus-visible:border-primary inline-flex cursor-pointer items-center justify-center rounded-full border border-black/10 bg-black/3 p-2 text-black/55 backdrop-blur-[6px] transition-colors outline-none hover:border-black/40 hover:text-black focus-visible:text-black dark:border-white/10 dark:bg-white/4 dark:text-white/55 dark:hover:border-white/40 dark:hover:text-white dark:focus-visible:text-white"
              onClick={toggle}
              type="button"
            >
              {isDark ? (
                <Sun aria-hidden className="size-3" />
              ) : (
                <Moon aria-hidden className="size-3" />
              )}
            </button>
          )}
        </ThemeToggle>
        <LocaleSelect />
      </header>

      <main className="relative z-10 mx-auto flex w-[min(100%,400px)] flex-1 flex-col justify-center px-[clamp(24px,3.5vw,44px)] py-8 max-[400px]:px-5">
        <h2 className="text-[27px] font-bold tracking-[0.01em] rtl:tracking-normal">
          {m.login_default_form_title()}
        </h2>
        <p className="mt-2 mb-7.5 text-[13.5px] leading-[1.8] text-black/55 dark:text-white/55">
          {m.login_default_form_sub()}
        </p>
        {children}
      </main>

      <footer className="relative z-10 flex flex-wrap justify-between gap-x-3 gap-y-2.5 border-t border-black/6 px-[clamp(24px,3.5vw,44px)] pt-5 pb-5.5 font-mono text-[9.5px] tracking-[0.16em] text-black/35 max-[400px]:px-5 dark:border-white/6 dark:text-white/35">
        <span>{`© 2026`}</span>
        <span>V1.0</span>
      </footer>
    </section>
  );
}
