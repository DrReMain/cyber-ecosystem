import { Switch } from "antd";
import { useAtomValue, useSetAtom } from "jotai";
import { Minus, Plus } from "lucide-react";
import { m } from "#/paraglide/messages";
import {
  dashboardPreferencesStore,
  SIDEBAR_WIDTH_MAX,
  SIDEBAR_WIDTH_MIN,
  SIDEBAR_WIDTH_STEP,
  sidebarWidthAtom,
} from "#/stores/dashboard-preferences/store";

export function SettingsLayout() {
  const { sidebar } = useAtomValue(dashboardPreferencesStore.atom);
  const setPreferences = useSetAtom(dashboardPreferencesStore.atom);
  const width = useAtomValue(sidebarWidthAtom);
  const setWidth = useSetAtom(sidebarWidthAtom);

  return (
    <div className="flex flex-col gap-4">
      <span className="text-[12px] text-ink-tertiary">
        {m.layout_dashboard_settings_group_sidebar()}
      </span>
      <div className="flex items-center justify-between gap-3">
        <span className="text-[13px]">{m.layout_dashboard_settings_sidebar_visible()}</span>
        <Switch
          aria-label={m.layout_dashboard_settings_sidebar_visible()}
          checked={sidebar.visible}
          onChange={(v) =>
            setPreferences((d) => {
              d.sidebar.visible = v;
            })
          }
        />
      </div>
      <div className="flex items-center justify-between gap-3">
        <span className="text-[13px]">{m.layout_dashboard_settings_sidebar_collapsed()}</span>
        <Switch
          aria-label={m.layout_dashboard_settings_sidebar_collapsed()}
          checked={sidebar.collapsed}
          disabled={!sidebar.visible}
          onChange={(v) =>
            setPreferences((d) => {
              d.sidebar.collapsed = v;
            })
          }
        />
      </div>
      <div className="flex items-center justify-between gap-3">
        <span className="text-[13px]">{m.layout_dashboard_settings_sidebar_width()}</span>
        <div className="flex h-8 items-center overflow-hidden rounded-md border border-line">
          <button
            aria-label={m.layout_dashboard_settings_sidebar_width_decrease()}
            className="flex h-full w-8 items-center justify-center text-ink-secondary hover:bg-fill-tertiary disabled:cursor-not-allowed disabled:opacity-40"
            disabled={width <= SIDEBAR_WIDTH_MIN}
            onClick={() => setWidth(width - SIDEBAR_WIDTH_STEP)}
            type="button"
          >
            <Minus aria-hidden className="size-3" />
          </button>
          <span className="w-12 border-line-soft border-x text-center font-mono text-[13px] leading-8">
            {width}
          </span>
          <button
            aria-label={m.layout_dashboard_settings_sidebar_width_increase()}
            className="flex h-full w-8 items-center justify-center text-ink-secondary hover:bg-fill-tertiary disabled:cursor-not-allowed disabled:opacity-40"
            disabled={width >= SIDEBAR_WIDTH_MAX}
            onClick={() => setWidth(width + SIDEBAR_WIDTH_STEP)}
            type="button"
          >
            <Plus aria-hidden className="size-3" />
          </button>
        </div>
      </div>
    </div>
  );
}
