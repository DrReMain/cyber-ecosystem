import { StyleProvider } from "@ant-design/cssinjs"
import { getAllSkins, getSkin, SkinSwitcher } from "@shared/antd/skins"
import type { Preview } from "@storybook/react-vite"
import { theme } from "antd"

const skinItems = getAllSkins().map((s) => ({
  value: s.id,
  title: s.name,
}))

function ThemedBody({
  children,
  bodyBg,
  isDark,
  layout,
}: {
  children: React.ReactNode
  bodyBg: string
  isDark: boolean
  layout?: string
}) {
  const { token } = theme.useToken()

  return (
    <div
      className={isDark ? "dark" : undefined}
      style={{
        background: bodyBg,
        color: token.colorText,
        minHeight: "100vh",
        padding: layout === "fullscreen" || layout === "centered" ? 0 : 24,
        transition: "background 0.3s, color 0.3s",
      }}
    >
      {children}
    </div>
  )
}

const preview: Preview = {
  parameters: {
    options: {
      storySort: {
        order: [
          "Antd",
          [
            "Foundation",
            "General",
            "Data Entry",
            "Data Display",
            "Feedback",
            "Navigation",
            "Layout",
            "Other",
            "Combinations",
          ],
          "Antd Pro",
          ["Components"],
        ],
      },
    },
    backgrounds: { disable: true },
    controls: { expanded: true },
    viewport: {
      options: {
        mobile: { name: "Mobile", styles: { width: "375px", height: "812px" } },
        tablet: {
          name: "Tablet",
          styles: { width: "768px", height: "1024px" },
        },
      },
    },
  },
  globalTypes: {
    skin: {
      name: "Skin",
      description: "Active skin",
      defaultValue: "default",
      toolbar: {
        items: skinItems,
        dynamicTitle: true,
      },
    },
    mode: {
      name: "Mode",
      description: "Light or Dark",
      defaultValue: "light",
      toolbar: {
        items: [
          { value: "light", icon: "sun", title: "Light" },
          { value: "dark", icon: "moon", title: "Dark" },
        ],
        dynamicTitle: true,
      },
    },
    compact: {
      name: "Compact",
      description: "Compact mode",
      defaultValue: "false",
      toolbar: {
        items: [
          { value: "false", title: "Normal" },
          { value: "true", title: "Compact" },
        ],
        dynamicTitle: true,
      },
    },
  },
  decorators: [
    (Story, context) => {
      const skinId = context.globals.skin ?? "default"
      const isDark = context.globals.mode === "dark"
      const compact = context.globals.compact === "true"

      const skin = getSkin(skinId)
      const bodyBg =
        skin?.bodyBg[isDark ? "dark" : "light"] ?? (isDark ? "#000" : "#fff")

      return (
        <StyleProvider layer>
          <SkinSwitcher skinId={skinId} isDark={isDark} compact={compact}>
            <ThemedBody
              bodyBg={bodyBg}
              isDark={isDark}
              layout={context.parameters.layout}
            >
              <Story />
            </ThemedBody>
          </SkinSwitcher>
        </StyleProvider>
      )
    },
  ],
}

export default preview
