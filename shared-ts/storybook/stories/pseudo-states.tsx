import { Flex } from "antd"
import { type ReactNode, useEffect, useRef } from "react"
import { Label } from "./helpers"

const PSEUDO_MAP = {
  hover: "pseudo-hover",
  focus: "pseudo-focus",
  active: "pseudo-active",
  "focus-within": "pseudo-focus-within",
  "focus-visible": "pseudo-focus-visible",
} as const

const PSEUDO_REGEX =
  /:(hover|focus-within|focus-visible|focus|active|visited|link|target)\b/g

const TAG_ID = "pseudo-states-injected"

function extractPseudoRules(rules: CSSRuleList, layer?: string): string {
  let css = ""
  for (const rule of rules) {
    if (rule instanceof CSSLayerBlockRule) {
      css += extractPseudoRules(rule.cssRules, rule.name ?? undefined)
    } else if (rule instanceof CSSStyleRule) {
      const sel = rule.selectorText
      if (!sel || !PSEUDO_REGEX.test(sel)) continue

      PSEUDO_REGEX.lastIndex = 0
      let m: RegExpExecArray | null
      // biome-ignore lint/suspicious/noAssignInExpressions: idiomatic regex exec loop
      while ((m = PSEUDO_REGEX.exec(sel))) {
        const pseudo = m[1] as keyof typeof PSEUDO_MAP
        const cls = PSEUDO_MAP[pseudo]
        if (!cls) continue

        // Split comma-separated selectors so each part gets its own pseudo
        // prefix. Otherwise `.a:hover, .b:focus` stripped becomes `.a, .b`
        // and only the first part gets the `.pseudo-hover` prefix.
        const parts = sel.split(",")
        const prefixedParts = parts
          .map((part) => {
            const trimmed = part.trim()
            if (!trimmed.includes(`:${pseudo}`)) return null
            // Use replaceAll with a string to avoid resetting the global
            // PSEUDO_REGEX.lastIndex, which would break the outer while loop.
            return `.${cls} ${trimmed.replaceAll(`:${pseudo}`, "")}`
          })
          .filter(Boolean)

        if (prefixedParts.length === 0) continue

        const newSel = prefixedParts.join(", ")
        const body = `{ ${rule.style.cssText} }`
        css += layer
          ? `@layer ${layer} { ${newSel} ${body} }\n`
          : `${newSel} ${body}\n`
      }
    }
  }
  return css
}

function rewriteAll() {
  const existing = document.getElementById(TAG_ID)
  if (existing) existing.remove()

  let css = ""
  for (const sheet of document.styleSheets) {
    try {
      const rules = sheet.cssRules
      if (!rules) continue
      css += extractPseudoRules(rules)
    } catch {
      // cross-origin
    }
  }

  if (!css) return

  const tag = document.createElement("style")
  tag.id = TAG_ID
  tag.textContent = css
  document.head.appendChild(tag)
}

const STATES = ["default", "hover", "focus", "active"] as const
const LABELS: Record<string, string> = {
  default: "Default",
  hover: "Hover",
  focus: "Focus",
  active: "Active",
}

export function PseudoStates({ children }: { children: ReactNode }) {
  const rewritten = useRef(false)

  useEffect(() => {
    if (rewritten.current) return
    const t = setTimeout(() => {
      rewriteAll()
      rewritten.current = true
    }, 200)
    return () => clearTimeout(t)
  }, [])

  return (
    <Flex gap={16} align="start">
      {STATES.map((state) => (
        <Flex key={state} vertical gap={4} align="center" flex={1}>
          <Label>{LABELS[state]}</Label>
          <div
            className={
              state !== "default" ? `pseudo-${state} pseudo-${state}-all` : ""
            }
          >
            {children}
          </div>
        </Flex>
      ))}
    </Flex>
  )
}
