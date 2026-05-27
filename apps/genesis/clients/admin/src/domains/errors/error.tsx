import { useRouter } from "@tanstack/react-router"
import { useState } from "react"
import { LocaleSwitcher } from "#/domains/i18n"
import { useTheme } from "#/domains/theme"
import { m } from "#/paraglide/messages"

interface IProps {
  error: Error
  showDetails?: boolean
}

export function ErrorPage({ error, showDetails }: Readonly<IProps>) {
  const router = useRouter()
  const { preference } = useTheme()
  const isDark = preference === "dark"

  return (
    <div
      className={`relative flex min-h-screen flex-col items-center justify-center gap-6 ${isDark ? "bg-black" : "bg-white"}`}
    >
      <div className="absolute top-4 right-4">
        <LocaleSwitcher />
      </div>
      <h1
        className={`font-black text-[4rem] leading-none tracking-tighter sm:text-[8rem] ${isDark ? "text-white" : "text-gray-900"}`}
      >
        500
      </h1>
      <p
        className={`max-w-md px-6 text-center font-medium text-base sm:text-lg ${isDark ? "text-white/60" : "text-gray-500"}`}
      >
        {m.error_error_subtitle()}
      </p>
      <div className="flex gap-3">
        <button
          className={`rounded-lg px-6 py-3 font-semibold text-sm transition-all duration-300 hover:scale-105 active:scale-95 sm:px-8 sm:text-base ${
            isDark
              ? "bg-white/10 text-white/70 hover:bg-white/20"
              : "bg-gray-100 text-gray-700 hover:bg-gray-200"
          }`}
          onClick={() => router.navigate({ to: "/" })}
          type="button"
        >
          {m.error_back_home()}
        </button>
        <button
          className={`rounded-lg px-6 py-3 font-semibold text-sm transition-all duration-300 hover:scale-105 active:scale-95 sm:px-8 sm:text-base ${
            isDark
              ? "bg-white/10 text-white/70 hover:bg-white/20"
              : "bg-gray-100 text-gray-700 hover:bg-gray-200"
          }`}
          onClick={() => router.invalidate()}
          type="button"
        >
          {m.error_retry()}
        </button>
      </div>
      {showDetails ? <ErrorDetails error={error} isDark={isDark} /> : null}
    </div>
  )
}

function ErrorDetails({ error, isDark }: { error: Error; isDark: boolean }) {
  const [expanded, setExpanded] = useState(false)
  const [copied, setCopied] = useState(false)

  const text = error.stack
    ? `${error.message}\n\n${error.stack}`
    : error.message

  async function handleCopy() {
    try {
      await navigator.clipboard.writeText(text)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    } catch {
      // clipboard API unavailable or permission denied
    }
  }

  return (
    <div
      className={`mt-6 w-full max-w-2xl rounded-lg border px-4 sm:px-0 ${isDark ? "border-white/10" : "border-gray-200"}`}
    >
      <button
        className={`flex w-full items-center justify-between px-4 py-2.5 text-left font-medium text-sm transition-colors ${
          isDark
            ? "text-white/70 hover:text-white"
            : "text-gray-600 hover:text-gray-900"
        }`}
        onClick={() => setExpanded((v) => !v)}
        type="button"
      >
        <span className="flex items-center gap-2">
          <svg
            aria-hidden="true"
            className={`h-3.5 w-3.5 transition-transform ${expanded ? "rotate-90" : ""}`}
            fill="none"
            stroke="currentColor"
            strokeWidth={2}
            viewBox="0 0 24 24"
          >
            <path d="M9 5l7 7-7 7" />
          </svg>
          Error Details
        </span>
      </button>
      <div
        className={`grid transition-[grid-template-rows] duration-200 ${expanded ? "grid-rows-[1fr]" : "grid-rows-[0fr]"}`}
      >
        <div className="overflow-hidden">
          <div
            className={`relative mx-4 mb-3 rounded-md p-4 font-mono text-xs leading-relaxed ${
              isDark ? "bg-white/5 text-red-400" : "bg-gray-50 text-red-600"
            }`}
          >
            <button
              aria-label={copied ? "Copied" : "Copy"}
              className={`absolute top-2 right-2 rounded px-1.5 py-0.5 text-xs transition-colors ${
                isDark
                  ? "text-white/40 hover:text-white/70"
                  : "text-gray-400 hover:text-gray-700"
              }`}
              onClick={handleCopy}
              type="button"
            >
              {copied ? "Copied" : "Copy"}
            </button>
            <pre className="whitespace-pre-wrap">{text}</pre>
          </div>
        </div>
      </div>
    </div>
  )
}
