import { useTheme } from "#/domains/theme"
import { m } from "#/paraglide/messages"

export function Pending() {
  const { preference } = useTheme()
  const isDark = preference === "dark"

  return (
    <div
      className={`flex min-h-screen flex-col items-center justify-center gap-4 ${isDark ? "bg-black" : "bg-white"}`}
    >
      <div
        aria-label={m.pending_subtitle()}
        className={`h-8 w-8 animate-spin rounded-full border-[3px] border-t-transparent ${
          isDark
            ? "border-white/30 border-t-white"
            : "border-gray-300 border-t-gray-900"
        }`}
        role="status"
      />
      <p
        className={`font-medium text-sm ${isDark ? "text-white/50" : "text-gray-400"}`}
      >
        {m.pending_subtitle()}
      </p>
    </div>
  )
}
