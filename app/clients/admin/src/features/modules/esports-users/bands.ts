import { m } from "#/paraglide/messages";

const BAND_LABELS: Record<string, () => string> = {
  under12: () => m.esports_users_band_under12(),
  "12-14": () => m.esports_users_band_12_14(),
  "15-17": () => m.esports_users_band_15_17(),
  "18-22": () => m.esports_users_band_18_22(),
  "23-25": () => m.esports_users_band_23_25(),
  "26-29": () => m.esports_users_band_26_29(),
  "30-40": () => m.esports_users_band_30_40(),
  over40: () => m.esports_users_band_over40(),
};

export function bandLabel(band: string | undefined): string | undefined {
  if (band === undefined) {
    return undefined;
  }
  return (BAND_LABELS[band] ?? (() => band))();
}
