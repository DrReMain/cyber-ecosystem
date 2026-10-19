import { Mermaid } from "@ant-design/x";
import { useTheme } from "@cyber-ecosystem/shared-theme";
import type { ComponentProps } from "react";

// mermaid.initialize is a global singleton the component re-runs whenever the
// config object identity changes — keep one stable config per skin instead of
// rebuilding an inline object every render.
const themeConfigs = {
  dark: { theme: "dark" },
  light: { theme: "default" },
} satisfies Record<"dark" | "light", NonNullable<ComponentProps<typeof Mermaid>["config"]>>;

// Loaded through React.lazy from markdown-content: the mermaid runtime is a
// megabyte-scale library that only pays off once an answer actually contains
// a fenced mermaid block.
export function MermaidDiagram({ source }: Readonly<{ source: string }>) {
  const isDark = useTheme().preference === "dark";
  return <Mermaid config={isDark ? themeConfigs.dark : themeConfigs.light}>{source}</Mermaid>;
}
