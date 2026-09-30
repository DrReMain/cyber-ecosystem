import clsx from "clsx";
import { Check, ChevronDown, Globe } from "lucide-react";
import { Button, I18nProvider, ListBox, ListBoxItem, Popover, Select } from "react-aria-components";
import { m } from "#/paraglide/messages";
import { getLocale, isLocale, locales, setLocale } from "#/paraglide/runtime";

const codeOf = (locale: string) => (locale.split("-")[0] ?? locale).toUpperCase();

export function LocaleSelect() {
  return (
    <I18nProvider locale={getLocale()}>
      <span className="inline-flex font-normal font-sans text-[13px] text-black normal-case not-italic leading-none tracking-normal dark:text-white">
        <Select
          aria-label="Language"
          onChange={(key) => {
            if (typeof key === "string" && isLocale(key)) {
              void setLocale(key);
            }
          }}
        >
          <Button className="inline-flex cursor-pointer items-center gap-2 rounded-full border border-black/10 bg-black/3 px-3.5 py-2 font-mono text-[11px] text-black/55 leading-none tracking-[0.16em] outline-none backdrop-blur-[6px] transition-colors hover:border-black/40 hover:text-black focus-visible:border-primary focus-visible:text-black dark:border-white/10 dark:bg-white/4 dark:text-white/55 dark:focus-visible:text-white dark:hover:border-white/40 dark:hover:text-white">
            <Globe aria-hidden className="size-3" />
            {codeOf(getLocale())}
            <ChevronDown aria-hidden className="size-3" />
          </Button>
          <Popover
            className="z-50 min-w-50 rounded-xl border border-black/10 bg-white p-1.5 shadow-[0_12px_40px_rgba(0,0,0,0.15)] backdrop-blur-xl dark:border-white/10 dark:bg-[#0a0f1a]/95 dark:shadow-[0_12px_40px_rgba(0,0,0,0.5)]"
            offset={8}
            placement="bottom end"
          >
            <ListBox className="outline-none">
              {locales.map((locale) => (
                <ListBoxItem
                  className={({ isFocused, isSelected }) =>
                    clsx(
                      "flex cursor-pointer items-center justify-between gap-6 rounded-lg px-3 py-2 text-[13px] leading-normal outline-none transition-colors",
                      isFocused && "bg-black/6 text-black dark:bg-white/6 dark:text-white",
                      !isFocused && isSelected && "text-black dark:text-white",
                      !(isFocused || isSelected) && "text-black/70 dark:text-white/70",
                    )
                  }
                  id={locale}
                  key={locale}
                  textValue={m.common_locale_name({ locale })}
                >
                  {({ isSelected }) => (
                    <>
                      <span className="truncate" dir="auto">
                        {m.common_locale_name({ locale })}
                      </span>
                      <span className="inline-flex items-center gap-1.5 font-mono text-[10px] text-black/35 tracking-[0.16em] dark:text-white/35">
                        {locale}
                        {isSelected ? <Check aria-hidden className="size-3 text-primary" /> : null}
                      </span>
                    </>
                  )}
                </ListBoxItem>
              ))}
            </ListBox>
          </Popover>
        </Select>
      </span>
    </I18nProvider>
  );
}
