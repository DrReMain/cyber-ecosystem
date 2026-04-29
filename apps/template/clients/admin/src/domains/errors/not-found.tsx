import { useRouter } from "@tanstack/react-router"
import { LocaleSwitcher } from "#/domains/i18n"
import { useTheme } from "#/domains/theme"
import { m } from "#/paraglide/messages"

export function NotFound() {
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
        404
      </h1>
      <p
        className={`max-w-md px-6 text-center font-medium text-base sm:text-lg ${isDark ? "text-white/60" : "text-gray-500"}`}
      >
        {m.error_not_found_subtitle()}
      </p>
      <button
        className={`mt-2 rounded-lg px-6 py-3 font-semibold text-sm transition-all duration-300 hover:scale-105 active:scale-95 sm:px-8 sm:text-base ${
          isDark
            ? "bg-white/10 text-white/70 hover:bg-white/20"
            : "bg-gray-100 text-gray-700 hover:bg-gray-200"
        }`}
        onClick={() => router.navigate({ to: "/" })}
        type="button"
      >
        {m.error_back_home()}
      </button>
    </div>
  )
}
