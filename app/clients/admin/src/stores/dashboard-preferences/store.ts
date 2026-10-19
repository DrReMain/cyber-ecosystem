import { defineStore } from "@cyber-ecosystem/shared-store";
import { atom } from "jotai";
import { z } from "zod";

export const SIDEBAR_WIDTH_MIN = 160;
export const SIDEBAR_WIDTH_MAX = 320;
export const SIDEBAR_WIDTH_STEP = 10;
export const SIDEBAR_COLLAPSED_WIDTH = 64;
const SIDEBAR_WIDTH_DEFAULT = 220;

const preferencesSchema = z.object({
  sidebar: z
    .object({
      accordion: z.boolean().default(false),
      collapsed: z.boolean().default(false),
      fixed: z.boolean().default(true),
      visible: z.boolean().default(true),
      width: z
        .number()
        .min(SIDEBAR_WIDTH_MIN)
        .max(SIDEBAR_WIDTH_MAX)
        .default(SIDEBAR_WIDTH_DEFAULT)
        .catch(SIDEBAR_WIDTH_DEFAULT),
    })
    .prefault({}),
  watermark: z
    .object({
      enabled: z.boolean().default(true),
    })
    .prefault({}),
});

export type Preferences = z.infer<typeof preferencesSchema>;

const defaults = preferencesSchema.parse({});

export const DEFAULT_PREFERENCES: Preferences = Object.freeze({
  sidebar: Object.freeze(defaults.sidebar),
  watermark: Object.freeze(defaults.watermark),
});

export const dashboardPreferencesStore = defineStore(
  "store_dashboard_preferences",
  DEFAULT_PREFERENCES,
  {
    persist: true,
    debugLabel: "DashboardPreferences",
    schema: preferencesSchema,
  },
);

export const sidebarWidthAtom = atom(
  (get) => get(dashboardPreferencesStore.atom).sidebar.width,
  (_get, set, width: number) =>
    set(dashboardPreferencesStore.atom, (d) => {
      d.sidebar.width = Math.min(SIDEBAR_WIDTH_MAX, Math.max(SIDEBAR_WIDTH_MIN, Math.round(width)));
    }),
);

export const watermarkEnabledAtom = atom(
  (get) => get(dashboardPreferencesStore.atom).watermark.enabled,
);
