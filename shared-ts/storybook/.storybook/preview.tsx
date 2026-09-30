import { getAllSkins, getSkin, SkinSwitcher } from "@cyber-ecosystem/shared-antd/skins";
import type { Preview } from "@storybook/react-vite";
import { theme } from "antd";
import type { ReactNode } from "react";

const skinItems = getAllSkins().map((s) => ({
  value: s.id,
  title: s.name,
}));

/**
 * Sits INSIDE SkinSwitcher so its useToken() reads the active skin's resolved
 * tokens — text color and background track whatever the toolbar selected.
 */
function ThemedBody({
  children,
  bodyBg,
  layout,
}: {
  children: ReactNode;
  bodyBg: string;
  layout?: string;
}) {
  const { token } = theme.useToken();

  return (
    <div
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
  );
}

const preview: Preview = {
  parameters: {
    options: {
      storySort: {
        order: ["Foundation", "Components", "Combinations"],
      },
    },
    backgrounds: { disable: true },
    controls: { expanded: true },
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
      const skinId = context.globals.skin ?? "default";
      const isDark = context.globals.mode === "dark";
      const compact = context.globals.compact === "true";

      const skin = getSkin(skinId);
      const bodyBg = skin?.bodyBg[isDark ? "dark" : "light"] ?? (isDark ? "#000" : "#fff");

      return (
        <SkinSwitcher compact={compact} isDark={isDark} skinId={skinId}>
          <ThemedBody bodyBg={bodyBg} layout={context.parameters.layout}>
            <Story />
          </ThemedBody>
        </SkinSwitcher>
      );
    },
  ],
};

export default preview;
