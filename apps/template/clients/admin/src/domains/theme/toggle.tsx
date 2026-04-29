import { Moon, Sun } from "lucide-react"
import type { MouseEvent, ReactNode } from "react"
import { useEffect, useRef, useState } from "react"
import { useTheme, useToggleTheme } from "./hooks"

interface DefaultProps {
  className?: string
}

interface ChildrenProps {
  children: (props: {
    isDark: boolean
    toggle: (e: MouseEvent) => void
  }) => ReactNode
}

type IProps = DefaultProps | ChildrenProps

function hasChildren(props: IProps): props is ChildrenProps {
  return "children" in props && typeof props.children === "function"
}

export function ThemeToggle(props: Readonly<IProps>) {
  const { preference } = useTheme()
  const toggleTheme = useToggleTheme()
  const isDark = preference === "dark"
  const [iconDark, setIconDark] = useState(isDark)
  const mountedRef = useRef(false)

  useEffect(() => {
    if (!mountedRef.current) {
      mountedRef.current = true
      setIconDark(isDark)
      return
    }

    const hasVT = typeof document.startViewTransition === "function"
    if (!hasVT) {
      setIconDark(isDark)
      return
    }

    const id = setTimeout(() => setIconDark(isDark), 500)
    return () => clearTimeout(id)
  }, [isDark])

  if (hasChildren(props)) {
    return props.children({ isDark, toggle: toggleTheme })
  }

  const { className } = props as DefaultProps
  return (
    <button
      aria-label={isDark ? "Switch to light mode" : "Switch to dark mode"}
      className={`relative inline-flex h-9 w-9 items-center justify-center overflow-hidden rounded-md border transition-colors ${
        isDark
          ? "border-white/15 text-white/70 hover:border-white/30 hover:text-white"
          : "border-gray-300 text-gray-600 hover:border-gray-400 hover:text-gray-900"
      } ${className ?? ""}`}
      onClick={toggleTheme}
      type="button"
    >
      <Sun
        className={`absolute h-4 w-4 transition-all duration-300 ${
          iconDark ? "rotate-0 scale-100" : "-rotate-90 scale-0"
        }`}
      />
      <Moon
        className={`absolute h-4 w-4 transition-all duration-300 ${
          iconDark ? "rotate-90 scale-0" : "rotate-0 scale-100"
        }`}
      />
    </button>
  )
}
