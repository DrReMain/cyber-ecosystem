import { Moon, Sun } from "lucide-react";
import type { MouseEvent, ReactNode } from "react";
import { useEffect, useRef, useState } from "react";
import { useTheme, useToggleTheme } from "./hooks";
import "./toggle.css";
import "./view-transition.css";

interface DefaultProps {
  className?: string;
}

interface ChildrenProps {
  children: (props: { isDark: boolean; toggle: (e: MouseEvent) => void }) => ReactNode;
}

type IProps = DefaultProps | ChildrenProps;

function hasChildren(props: IProps): props is ChildrenProps {
  return "children" in props && typeof props.children === "function";
}

export function ThemeToggle(props: Readonly<IProps>) {
  const { preference } = useTheme();
  const toggleTheme = useToggleTheme();
  const isDark = preference === "dark";
  const [iconDark, setIconDark] = useState(isDark);
  const mountedRef = useRef(false);

  useEffect(() => {
    if (!mountedRef.current) {
      mountedRef.current = true;
      setIconDark(isDark);
      return;
    }

    const hasVT = typeof document.startViewTransition === "function";
    if (!hasVT) {
      setIconDark(isDark);
      return;
    }

    const id = setTimeout(() => setIconDark(isDark), 500);
    return () => clearTimeout(id);
  }, [isDark]);

  if (hasChildren(props)) {
    return props.children({ isDark, toggle: toggleTheme });
  }

  const { className } = props as DefaultProps;
  return (
    <button
      aria-label={isDark ? "Switch to light mode" : "Switch to dark mode"}
      className={`theme-toggle ${isDark ? "theme-dark" : ""} ${iconDark ? "icon-dark" : ""} ${className ?? ""}`}
      onClick={toggleTheme}
      type="button"
    >
      <Sun className="theme-toggle-icon sun" />
      <Moon className="theme-toggle-icon moon" />
    </button>
  );
}
