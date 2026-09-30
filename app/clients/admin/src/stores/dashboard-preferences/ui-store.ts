import { defineStore } from "@cyber-ecosystem/shared-store";

export const dashboardPreferencesUiStore = defineStore(
  "store_dashboard_preferences_ui",
  { drawerOpen: false, maximized: false },
  { debugLabel: "DashboardPreferencesUi" },
);
