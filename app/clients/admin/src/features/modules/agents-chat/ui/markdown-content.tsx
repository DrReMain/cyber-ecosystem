import { CodeHighlighter } from "@ant-design/x";
import { type ComponentProps, XMarkdown } from "@ant-design/x-markdown";
import Latex from "@ant-design/x-markdown/plugins/Latex";
import "@ant-design/x-markdown/themes/light.css";
import "@ant-design/x-markdown/themes/dark.css";
import { useTheme } from "@cyber-ecosystem/shared-theme";
import clsx from "clsx";
import type { ReactNode } from "react";
import { lazy, Suspense } from "react";
import oneDark from "react-syntax-highlighter/dist/esm/styles/prism/one-dark";

const MermaidDiagram = lazy(async () => {
  const { MermaidDiagram: Diagram } = await import("./mermaid-diagram");
  return { default: Diagram };
});

const darkHighlightProps = { style: oneDark };

const markdownComponents = {
  code: ({ block, lang, children }: ComponentProps): ReactNode =>
    block ? (
      lang === "mermaid" ? (
        <Suspense fallback={null}>
          <MermaidDiagram source={flatten(children)} />
        </Suspense>
      ) : (
        <CodeBlock lang={lang}>{flatten(children)}</CodeBlock>
      )
    ) : (
      <code>{children}</code>
    ),
} satisfies Record<string, (props: ComponentProps) => ReactNode>;

const markdownConfig = { extensions: Latex() };

function CodeBlock({ lang, children }: Readonly<{ lang?: string; children: string }>) {
  const isDark = useTheme().preference === "dark";
  return (
    <CodeHighlighter highlightProps={isDark ? darkHighlightProps : undefined} lang={lang ?? "text"}>
      {children}
    </CodeHighlighter>
  );
}

function flatten(children: ReactNode): string {
  return Array.isArray(children) ? children.join("") : String(children ?? "");
}

export function MarkdownContent({
  content,
  streaming,
}: Readonly<{ content: string; streaming: boolean }>) {
  const isDark = useTheme().preference === "dark";
  return (
    <XMarkdown
      className={clsx("min-w-0", isDark ? "x-markdown-dark" : "x-markdown-light")}
      components={markdownComponents}
      config={markdownConfig}
      content={content}
      escapeRawHtml
      openLinksInNewTab
      streaming={{ hasNextChunk: streaming, tail: true }}
    />
  );
}
