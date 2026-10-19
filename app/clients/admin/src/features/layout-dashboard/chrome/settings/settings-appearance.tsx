import { getAllSkins } from "@cyber-ecosystem/shared-antd/skins";
import { type ThemeMode, useTheme } from "@cyber-ecosystem/shared-theme";
import { Switch } from "antd";
import clsx from "clsx";
import { Monitor, Moon, Sun } from "lucide-react";
import { m } from "#/paraglide/messages";
import { SettingsSkinCard } from "./settings-skin-card";

const THEME_OPTIONS = [
  { icon: Sun, label: m.layout_dashboard_settings_theme_light, value: "light" },
  { icon: Moon, label: m.layout_dashboard_settings_theme_dark, value: "dark" },
  { icon: Monitor, label: m.layout_dashboard_settings_theme_system, value: "system" },
] as const satisfies ReadonlyArray<{
  icon: typeof Sun;
  label: () => string;
  value: ThemeMode;
}>;

export function SettingsAppearance() {
  const { compact, mode, setCompact, setMode, skinId, setSkinId } = useTheme();

  return (
    <div className="flex flex-col gap-5">
      <div className="flex flex-col gap-2">
        <span className="font-medium text-[13px]">{m.layout_dashboard_settings_theme()}</span>
        <div className="grid grid-cols-3 gap-2">
          {THEME_OPTIONS.map(({ icon: Icon, label, value }) => {
            const active = mode === value;
            return (
              <button
                aria-pressed={active}
                className={clsx(
                  "flex flex-col items-center gap-1.5 rounded-lg border px-2 py-3 text-[12px] transition-colors",
                  active
                    ? "border-primary bg-primary/6 font-medium text-primary"
                    : "border-line text-ink-secondary hover:border-line-hover",
                )}
                key={value}
                onClick={() => setMode(value)}
                type="button"
              >
                <Icon aria-hidden className="size-4" />
                {label()}
              </button>
            );
          })}
        </div>
      </div>
      <div className="flex flex-col gap-2.5 border-line-soft border-t pt-4">
        <span className="font-medium text-[13px]">{m.layout_dashboard_settings_skin()}</span>
        <div className="flex items-center justify-between gap-3">
          <span className="text-[13px]">{m.layout_dashboard_settings_compact()}</span>
          <Switch
            aria-label={m.layout_dashboard_settings_compact()}
            checked={compact}
            onChange={setCompact}
          />
        </div>
        <div className="grid grid-cols-2 gap-2.5">
          {getAllSkins().map((skin) => (
            <SettingsSkinCard
              key={skin.id}
              onSelect={() => setSkinId(skin.id)}
              selected={skin.id === skinId}
              skin={skin}
            />
          ))}
        </div>
      </div>
    </div>
  );
}
