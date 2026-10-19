import { AntdProvider } from "#/domains/antd";
import { useSessionWatcher } from "./auth/use-session-watcher";
import { DashboardLayoutBody } from "./chrome/dashboard-layout-body";

export function DashboardLayout() {
  useSessionWatcher();
  return (
    <AntdProvider>
      <DashboardLayoutBody />
    </AntdProvider>
  );
}
