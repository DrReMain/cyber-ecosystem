import { m } from "#/paraglide/messages";

// The eight-band vocabulary mirrors the server's derivation (Profile.age_band
// is server-written); this table only renders codes. Seamlessness per the
// product ruling: every age belongs to exactly one band.
const BAND_LABELS: Record<string, () => string> = {
  under12: () => m.profile_band_under12(),
  "12-14": () => m.profile_band_12_14(),
  "15-17": () => m.profile_band_15_17(),
  "18-22": () => m.profile_band_18_22(),
  "23-25": () => m.profile_band_23_25(),
  "26-29": () => m.profile_band_26_29(),
  "30-40": () => m.profile_band_30_40(),
  over40: () => m.profile_band_over40(),
};

export function bandLabel(band: string | undefined): string | undefined {
  if (band === undefined) {
    return undefined;
  }
  return (BAND_LABELS[band] ?? (() => band))();
}

// Client-side derivation mirror for the live hint under the birth selects —
// a UX affordance only; the server re-derives and stores the authoritative
// band on every write.
export function bandOfAge(age: number): string {
  if (age < 12) {
    return "under12";
  }
  if (age <= 14) {
    return "12-14";
  }
  if (age <= 17) {
    return "15-17";
  }
  if (age <= 22) {
    return "18-22";
  }
  if (age <= 25) {
    return "23-25";
  }
  if (age <= 29) {
    return "26-29";
  }
  if (age <= 40) {
    return "30-40";
  }
  return "over40";
}
