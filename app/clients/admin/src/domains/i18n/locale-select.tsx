import clsx from "clsx";
import { ChevronDown, Globe } from "lucide-react";
import { type KeyboardEvent, useEffect, useId, useRef, useState } from "react";
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
// roles, container-managed focus). Deliberately not a react-aria-components
// Select — that primitive pulls a ~180kB headless-primitive closure into
// every first-paint graph; this widget renders in __root error faces and
// both chrome surfaces. The observable contract: trigger keys open,
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

  const onTriggerKeyDown = (event: KeyboardEvent) => {
    const key: string = event.key;
    if (isOpenKey(key)) {
      event.preventDefault();
      setOpen(true);
    }
  };

  const onListKeyDown = (event: KeyboardEvent) => {
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
      className="text-ink relative inline-flex font-sans text-[13px] leading-none font-normal tracking-normal normal-case not-italic"
      ref={rootRef}
    >
      <button
        aria-controls={open ? listId : undefined}
        aria-expanded={open}
        aria-haspopup="listbox"
        aria-label="Language"
        className="border-line bg-fill-quaternary text-ink-secondary hover:border-line-hover hover:text-ink focus-visible:border-primary focus-visible:text-ink inline-flex cursor-pointer items-center gap-2 rounded-full border px-3.5 py-2 font-mono text-[11px] leading-none tracking-[0.16em] backdrop-blur-[6px] transition-colors outline-none"
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
          className="border-line bg-elevated absolute inset-e-0 top-full z-50 mt-2 min-w-50 rounded-xl border p-1.5 shadow-[0_12px_40px_rgba(0,0,0,0.15)] backdrop-blur-xl outline-none dark:shadow-[0_12px_40px_rgba(0,0,0,0.5)]"
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
                  "flex cursor-pointer items-center justify-between gap-6 rounded-lg px-3 py-2 text-[13px] leading-normal transition-colors outline-none",
                  isFocused && "bg-fill-tertiary text-ink",
                  !isFocused && isSelected && "text-ink",
                  !(isFocused || isSelected) && "text-ink-secondary",
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
                <span className="text-ink-quaternary inline-flex items-center gap-1.5 font-mono text-[10px] tracking-[0.16em]">
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
