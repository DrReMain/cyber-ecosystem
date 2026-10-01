import clsx from "clsx";
import { ChevronDown, Globe } from "lucide-react";
import { useEffect, useId, useRef, useState } from "react";
import { m } from "#/paraglide/messages";
import { getLocale, isLocale, locales, setLocale } from "#/paraglide/runtime";

const codeOf = (locale: string) => (locale.split("-")[0] ?? locale).toUpperCase();

// Next highlighted index for a movement key, null for everything else.
const moveActive = (key: string, active: number): number | null => {
  if (key === "ArrowDown") return (active + 1) % locales.length;
  if (key === "ArrowUp") return (active - 1 + locales.length) % locales.length;
  if (key === "Home") return 0;
  if (key === "End") return locales.length - 1;
  return null;
};

const isCommitKey = (key: string) => key === "Enter" || key === " ";
const isOpenKey = (key: string) => key === "ArrowDown" || isCommitKey(key);

// Self-contained dropdown (APG listbox pattern: generic elements + explicit
// roles, container-managed focus). The previous react-aria-components Select
// pulled a ~180kB headless-primitive closure into every first-paint graph —
// this widget renders in __root error faces and both chrome surfaces. The
// hand-rolled version keeps the observable contract: trigger keys open,
// arrows wrap, Enter/Space/click commit, Escape/outside-pointer closes and
// returns focus, focus moves into the list with aria-activedescendant while
// open, and the panel anchors to the logical end so RTL layouts mirror.
export function LocaleSelect() {
  const current = getLocale();
  const [open, setOpen] = useState(false);
  const [active, setActive] = useState(() => Math.max(locales.indexOf(current), 0));
  const rootRef = useRef<HTMLSpanElement>(null);
  const triggerRef = useRef<HTMLButtonElement>(null);
  const listRef = useRef<HTMLDivElement>(null);
  const listId = useId();

  useEffect(() => {
    if (!open) return;
    const onPointerDown = (event: PointerEvent) => {
      if (!rootRef.current?.contains(event.target as Node)) setOpen(false);
    };
    document.addEventListener("pointerdown", onPointerDown);
    return () => document.removeEventListener("pointerdown", onPointerDown);
  }, [open]);

  useEffect(() => {
    if (open) listRef.current?.focus();
  }, [open]);

  const commit = (locale: string) => {
    setOpen(false);
    triggerRef.current?.focus();
    if (isLocale(locale) && locale !== current) void setLocale(locale);
  };

  const onTriggerKeyDown = (event: React.KeyboardEvent) => {
    const key: string = event.key;
    if (isOpenKey(key)) {
      event.preventDefault();
      setOpen(true);
    }
  };

  const onListKeyDown = (event: React.KeyboardEvent) => {
    const key: string = event.key;
    if (key === "Escape" || key === "Tab") {
      setOpen(false);
      if (key === "Escape") {
        event.preventDefault();
        triggerRef.current?.focus();
      }
      return;
    }
    const moved = moveActive(key, active);
    if (moved !== null) {
      event.preventDefault();
      setActive(moved);
      return;
    }
    if (isCommitKey(key) && locales[active]) {
      event.preventDefault();
      commit(locales[active]);
    }
  };

  return (
    <span
      className="relative inline-flex font-normal font-sans text-[13px] text-black normal-case not-italic leading-none tracking-normal dark:text-white"
      ref={rootRef}
    >
      <button
        aria-controls={open ? listId : undefined}
        aria-expanded={open}
        aria-haspopup="listbox"
        aria-label="Language"
        className="inline-flex cursor-pointer items-center gap-2 rounded-full border border-black/10 bg-black/3 px-3.5 py-2 font-mono text-[11px] text-black/55 leading-none tracking-[0.16em] outline-none backdrop-blur-[6px] transition-colors hover:border-black/40 hover:text-black focus-visible:border-primary focus-visible:text-black dark:border-white/10 dark:bg-white/4 dark:text-white/55 dark:focus-visible:text-white dark:hover:border-white/40 dark:hover:text-white"
        onClick={() => setOpen((value) => !value)}
        onKeyDown={onTriggerKeyDown}
        ref={triggerRef}
        type="button"
      >
        <Globe aria-hidden className="size-3" />
        {codeOf(current)}
        <ChevronDown aria-hidden className="size-3" />
      </button>
      {open ? (
        <div
          aria-activedescendant={`${listId}-${locales[active]}`}
          aria-label="Language"
          className="absolute inset-e-0 top-full z-50 mt-2 min-w-50 rounded-xl border border-black/10 bg-white p-1.5 shadow-[0_12px_40px_rgba(0,0,0,0.15)] outline-none backdrop-blur-xl dark:border-white/10 dark:bg-[#0a0f1a]/95 dark:shadow-[0_12px_40px_rgba(0,0,0,0.5)]"
          id={listId}
          onKeyDown={onListKeyDown}
          ref={listRef}
          role="listbox"
          tabIndex={-1}
        >
          {locales.map((locale, index) => {
            const isSelected = locale === current;
            const isFocused = index === active;
            return (
              <div
                aria-selected={isSelected}
                className={clsx(
                  "flex cursor-pointer items-center justify-between gap-6 rounded-lg px-3 py-2 text-[13px] leading-normal outline-none transition-colors",
                  isFocused && "bg-black/6 text-black dark:bg-white/6 dark:text-white",
                  !isFocused && isSelected && "text-black dark:text-white",
                  !(isFocused || isSelected) && "text-black/70 dark:text-white/70",
                )}
                id={`${listId}-${locale}`}
                key={locale}
                onClick={() => commit(locale)}
                onKeyDown={(event) => {
                  if (isCommitKey(event.key) && locale !== current) {
                    event.preventDefault();
                    commit(locale);
                  }
                }}
                onPointerEnter={() => setActive(index)}
                role="option"
                tabIndex={-1}
              >
                <span className="truncate" dir="auto">
                  {m.common_locale_name({ locale })}
                </span>
                <span className="inline-flex items-center gap-1.5 font-mono text-[10px] text-black/35 tracking-[0.16em] dark:text-white/35">
                  {locale}
                </span>
              </div>
            );
          })}
        </div>
      ) : null}
    </span>
  );
}
