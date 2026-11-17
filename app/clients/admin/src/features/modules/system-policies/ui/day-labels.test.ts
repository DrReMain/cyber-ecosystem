import { describe, expect, it } from "vitest";

import { dayLabels } from "./day-labels";

// The picker's contract with the wire `days` field: index = Go time.Weekday
// (0 = Sunday), seven labels, stable order across calls. The copy itself is
// walkthrough territory.
describe("dayLabels", () => {
  it("answers seven non-empty labels in stable Sunday-first order", () => {
    const first = dayLabels();
    expect(first).toHaveLength(7);
    expect(first.every((label) => label.length > 0)).toBe(true);
    expect(dayLabels()).toEqual(first);
  });
});
