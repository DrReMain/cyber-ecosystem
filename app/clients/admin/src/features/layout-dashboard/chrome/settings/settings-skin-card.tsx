import type { SkinPlugin } from "@cyber-ecosystem/shared-antd/skins";
import { useTheme } from "@cyber-ecosystem/shared-theme";
import { theme } from "antd";
import { Check } from "lucide-react";
import { SettingsSkinThumb } from "./settings-skin-thumb";

interface SettingsSkinCardProps {
  onSelect: () => void;
  selected: boolean;
  skin: SkinPlugin;
}

export function SettingsSkinCard({ onSelect, selected, skin }: Readonly<SettingsSkinCardProps>) {
  const { preference, compact } = useTheme();
  const { token } = theme.useToken();
  const SkinProvider = skin.Provider;

  return (
    <button
      aria-pressed={selected}
      className="flex flex-col gap-2 p-2 text-start"
      onClick={onSelect}
      style={{
        border: `1px solid ${selected ? token.colorPrimary : token.colorBorder}`,
        borderRadius: token.borderRadius,
        boxShadow: selected ? `0 0 0 1px ${token.colorPrimary}` : undefined,
      }}
      type="button"
    >
      <span className="relative block">
        <SkinProvider compact={compact} isDark={preference === "dark"}>
          <SettingsSkinThumb />
        </SkinProvider>
        {selected ? (
          <span
            className="absolute inset-e-1.5 top-1.5 flex size-4 items-center justify-center rounded-full text-ink-inverse"
            style={{ background: token.colorPrimary }}
          >
            <Check aria-hidden className="size-2.5" strokeWidth={3} />
          </span>
        ) : null}
      </span>
      <span className="text-[12px]">{skin.name}</span>
    </button>
  );
}
