import type { ReactNode } from "react";
import defaultSkin from "../skin-presets/default";
import lumaSkin from "../skin-presets/luma";
import lyraSkin from "../skin-presets/lyra";
import type { SkinPlugin } from "./types";

const SKINS: Record<string, SkinPlugin> = {
  [defaultSkin.id]: defaultSkin,
  [lumaSkin.id]: lumaSkin,
  [lyraSkin.id]: lyraSkin,
};

export function getSkin(id: string): SkinPlugin | undefined {
  return SKINS[id];
}

export function getAllSkins(): SkinPlugin[] {
  return Object.values(SKINS);
}

export function getSkinIds(): string[] {
  return Object.keys(SKINS);
}

export function SkinSwitcher({
  skinId,
  isDark,
  compact,
  children,
}: {
  skinId: string;
  isDark: boolean;
  compact: boolean;
  children: ReactNode;
}) {
  const skin = getSkin(skinId);
  const Provider = skin?.Provider;
  if (!Provider) return <>{children}</>;
  return (
    <Provider compact={compact} isDark={isDark}>
      {children}
    </Provider>
  );
}
