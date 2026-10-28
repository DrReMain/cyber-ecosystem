import { expect, test } from "@playwright/test";

// The H5 skeleton's reachability anchor: the login face answers the bare
// dev server with a rendered document — the deeper session and touch laws
// are the business lane's face and land with its slices.
test("login face is reachable", async ({ page }) => {
  const res = await page.goto("/login");
  expect(res?.status()).toBe(200);
});
