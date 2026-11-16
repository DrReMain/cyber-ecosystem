import { m } from "#/paraglide/messages";

// The four control captions every Filter card shares — a function so
// the paraglide calls resolve per render, never frozen at module load.
export function filterLabels() {
  return {
    search: m.common_filter_search(),
    reset: m.common_filter_reset(),
    fold: m.common_filter_fold(),
    expand: m.common_filter_expand(),
  };
}
