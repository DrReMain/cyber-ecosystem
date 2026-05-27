import { createCache, StyleProvider } from "@ant-design/cssinjs"
import { extractStaticStyle } from "@shared/antd/skins/ssr"
import { type PropsWithChildren, useState } from "react"

export function AntdRegistry({ children }: Readonly<PropsWithChildren>) {
  const [cache] = useState(() => createCache())
  return (
    <StyleProvider cache={cache} layer>
      {children}
      <StyleCollector cache={cache} />
    </StyleProvider>
  )
}

interface IPropsStyleCollector {
  cache: ReturnType<typeof createCache>
}

function StyleCollector({ cache }: Readonly<IPropsStyleCollector>) {
  const [css] = useState(() => {
    if (typeof document !== "undefined") return ""

    const styles = extractStaticStyle("", { antdCache: cache })
    const allCss = styles
      .map((s) => s.css)
      .filter(Boolean)
      .join("\n")

    return allCss ? `@layer ssr{${allCss}}` : ""
  })
  if (!css) return null
  return (
    <style
      // biome-ignore lint/security/noDangerouslySetInnerHtml: antd cssinjs SSR style extraction
      dangerouslySetInnerHTML={{ __html: css }}
      href="antd-cssinjs"
      precedence="high"
      suppressHydrationWarning
    />
  )
}
