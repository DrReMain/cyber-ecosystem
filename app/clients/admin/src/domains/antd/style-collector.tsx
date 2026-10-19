import type { createCache } from "@ant-design/cssinjs";
import { extractStaticStyle } from "@cyber-ecosystem/shared-antd/ssr";
import { useState } from "react";

interface StyleCollectorProps {
  cache: ReturnType<typeof createCache>;
}

export function StyleCollector({ cache }: Readonly<StyleCollectorProps>) {
  const [css] = useState(() => {
    if (typeof document !== "undefined") return "";

    const styles = extractStaticStyle("", { antdCache: cache });
    const allCss = styles
      .map((s) => s.css)
      .filter(Boolean)
      .join("\n");

    return allCss ? `@layer ssr{${allCss}}` : "";
  });
  if (!css) return null;
  return (
    <style
      // biome-ignore lint/security/noDangerouslySetInnerHtml: antd cssinjs SSR style extraction
      dangerouslySetInnerHTML={{ __html: css }}
      href="antd-cssinjs"
      precedence="high"
      suppressHydrationWarning
    />
  );
}
