import { Drawer } from "antd";
import { m } from "#/paraglide/messages";
import { getTextDirection } from "#/paraglide/runtime";

interface SettingsDrawerProps {
  onClose: () => void;
  open: boolean;
}

export function SettingsDrawer({ onClose, open }: Readonly<SettingsDrawerProps>) {
  return (
    <Drawer
      onClose={onClose}
      open={open}
      placement={getTextDirection() === "rtl" ? "left" : "right"}
      styles={{ wrapper: { maxWidth: "100vw" } }}
      title={m.layout_dashboard_topbar_settings()}
    >
      {/* TODO */}
    </Drawer>
  );
}
