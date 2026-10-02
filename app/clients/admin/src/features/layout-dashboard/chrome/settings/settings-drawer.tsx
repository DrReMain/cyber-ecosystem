import { Drawer, Segmented } from "antd";
import { useState } from "react";
import { m } from "#/paraglide/messages";
import { getTextDirection } from "#/paraglide/runtime";
import { SettingsAppearance } from "./settings-appearance";
import { SettingsGeneral } from "./settings-general";
import { SettingsLayout } from "./settings-layout";

type SettingsSection = "appearance" | "layout" | "general";

const SECTION_LABELS: Record<SettingsSection, () => string> = {
  appearance: m.layout_dashboard_settings_section_appearance,
  layout: m.layout_dashboard_settings_section_layout,
  general: m.layout_dashboard_settings_section_general,
};

const sectionOptions = () =>
  (Object.keys(SECTION_LABELS) as SettingsSection[]).map((value) => ({
    label: SECTION_LABELS[value](),
    value,
  }));

interface SettingsDrawerProps {
  onClose: () => void;
  open: boolean;
}

export function SettingsDrawer({ onClose, open }: Readonly<SettingsDrawerProps>) {
  const [section, setSection] = useState<SettingsSection>("appearance");

  return (
    <Drawer
      onClose={onClose}
      open={open}
      placement={getTextDirection() === "rtl" ? "left" : "right"}
      styles={{ wrapper: { maxWidth: "100vw" } }}
      title={m.layout_dashboard_topbar_settings()}
    >
      <div className="flex flex-col gap-4">
        <Segmented
          block
          onChange={(v) => setSection(v as SettingsSection)}
          options={sectionOptions()}
          value={section}
        />
        {section === "appearance" ? (
          <SettingsAppearance />
        ) : section === "layout" ? (
          <SettingsLayout />
        ) : (
          <SettingsGeneral />
        )}
      </div>
    </Drawer>
  );
}
