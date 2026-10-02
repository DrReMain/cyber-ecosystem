import { Select, Switch } from "antd";
import { useAtomValue, useSetAtom } from "jotai";
import { m } from "#/paraglide/messages";
import { getLocale, isLocale, locales, setLocale } from "#/paraglide/runtime";
import {
  dashboardPreferencesStore,
  watermarkEnabledAtom,
} from "#/stores/dashboard-preferences/store";

export function SettingsGeneral() {
  const setPreferences = useSetAtom(dashboardPreferencesStore.atom);
  const watermarkEnabled = useAtomValue(watermarkEnabledAtom);

  const changeLocale = (value: string) => {
    if (isLocale(value)) void setLocale(value);
  };

  return (
    <div className="flex flex-col gap-4">
      <div className="flex items-center justify-between gap-3">
        <span className="text-[13px]">{m.layout_dashboard_settings_language()}</span>
        <Select
          className="min-w-40"
          onChange={changeLocale}
          options={locales.map((locale) => ({
            value: locale,
            label: <span dir="auto">{m.common_locale_name({ locale })}</span>,
          }))}
          popupMatchSelectWidth={false}
          value={getLocale()}
        />
      </div>
      <div className="flex items-center justify-between gap-3">
        <span className="text-[13px]">{m.layout_dashboard_settings_watermark()}</span>
        <Switch
          aria-label={m.layout_dashboard_settings_watermark()}
          checked={watermarkEnabled}
          onChange={(v) =>
            setPreferences((d) => {
              d.watermark.enabled = v;
            })
          }
        />
      </div>
    </div>
  );
}
