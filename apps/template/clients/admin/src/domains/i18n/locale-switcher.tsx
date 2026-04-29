import type { KeyboardEvent, ReactNode } from "react"
import { useCallback, useEffect, useRef, useState } from "react"
import { m } from "#/paraglide/messages"
import { getLocale, locales, setLocale } from "#/paraglide/runtime"

interface DefaultProps {
  className?: string
}

interface ChildrenProps {
  children: (props: {
    current: string
    label: string
    open: boolean
    setOpen: (open: boolean) => void
  }) => ReactNode
}

type IProps = DefaultProps | ChildrenProps

function hasChildren(props: IProps): props is ChildrenProps {
  return "children" in props && typeof props.children === "function"
}

export function LocaleSwitcher(props: Readonly<IProps>) {
  const current = getLocale()
  const label = m.locale_name({ locale: current })
  const [open, setOpen] = useState(false)
  const containerRef = useRef<HTMLDivElement>(null)
  const itemRefs = useRef<(HTMLButtonElement | null)[]>([])

  useEffect(() => {
    function onClick(e: MouseEvent) {
      if (
        containerRef.current &&
        !containerRef.current.contains(e.target as Node)
      )
        setOpen(false)
    }
    document.addEventListener("click", onClick)
    return () => document.removeEventListener("click", onClick)
  }, [])

  const focusItem = useCallback((index: number) => {
    const clamped = Math.max(0, Math.min(index, locales.length - 1))
    itemRefs.current[clamped]?.focus()
  }, [])

  function onKeyDown(e: KeyboardEvent) {
    if (!open) return
    const active = document.activeElement as HTMLButtonElement | null
    const idx = itemRefs.current.indexOf(active)

    switch (e.key) {
      case "Escape":
        e.preventDefault()
        setOpen(false)
        break
      case "ArrowDown":
        e.preventDefault()
        focusItem(idx < 0 ? 0 : idx + 1)
        break
      case "ArrowUp":
        e.preventDefault()
        focusItem(idx < 0 ? 0 : idx - 1)
        break
      case "Home":
        e.preventDefault()
        focusItem(0)
        break
      case "End":
        e.preventDefault()
        focusItem(locales.length - 1)
        break
    }
  }

  const dropdown = open ? (
    <div
      className="absolute right-0 z-50 mt-1 min-w-full overflow-hidden whitespace-nowrap rounded-md border border-gray-200 bg-white py-1 shadow-lg dark:border-white/15 dark:bg-neutral-900"
      onKeyDown={onKeyDown}
      role="listbox"
    >
      {locales.map((locale, i) => (
        <button
          aria-selected={locale === current}
          className={`flex w-full items-center px-3 py-1.5 text-left text-sm transition-colors ${
            locale === current
              ? "bg-gray-100 font-medium text-gray-900 dark:bg-white/10 dark:text-white"
              : "text-gray-600 hover:bg-gray-50 hover:text-gray-900 dark:text-white/70 dark:hover:bg-white/5 dark:hover:text-white"
          }`}
          key={locale}
          onClick={() => {
            setOpen(false)
            setLocale(locale)
          }}
          ref={(el) => {
            itemRefs.current[i] = el
          }}
          role="option"
          type="button"
        >
          {m.locale_name({ locale })}
        </button>
      ))}
    </div>
  ) : null

  if (hasChildren(props)) {
    return (
      <div className="relative inline-block" ref={containerRef}>
        {props.children({
          current,
          label,
          open,
          setOpen,
        })}
        {dropdown}
      </div>
    )
  }

  const { className } = props as DefaultProps
  return (
    <div
      className={`relative inline-block ${className ?? ""}`}
      ref={containerRef}
    >
      <button
        aria-expanded={open}
        aria-haspopup="listbox"
        className="inline-flex min-w-34 items-center justify-between gap-1.5 rounded-md border border-gray-300 px-3 py-1.5 text-gray-600 text-sm transition-colors hover:border-gray-400 hover:text-gray-900 dark:border-white/15 dark:text-white/70 dark:hover:border-white/30 dark:hover:text-white"
        onClick={() => setOpen((v) => !v)}
        type="button"
      >
        {label}
        <svg
          aria-hidden="true"
          className={`h-3.5 w-3.5 transition-transform ${open ? "rotate-180" : ""}`}
          fill="none"
          stroke="currentColor"
          strokeWidth={2}
          viewBox="0 0 24 24"
        >
          <path d="M6 9l6 6 6-6" />
        </svg>
      </button>
      {dropdown}
    </div>
  )
}
