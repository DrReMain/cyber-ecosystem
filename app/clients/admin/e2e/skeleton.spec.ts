import { expect, test } from "@playwright/test";

// The admin skeleton's three load-bearing links, driven the way a user
// arrives: the login face renders its credential form, the session guard
// redirects a cookieless visitor from the protected tree back to login,
// and an unknown route answers the root not-found face with a 404.
test("login face renders the credential form", async ({ page }) => {
  await page.goto("/login");
  await expect(page.locator('input[name="account"]')).toBeVisible();
  await expect(page.locator('input[name="password"]')).toBeVisible();
});

test("session guard redirects cookieless visitors to login", async ({ page }) => {
  await page.goto("/dashboard");
  await expect(page).toHaveURL(/\/login/);
  await expect(page.locator('input[name="account"]')).toBeVisible();
});

test("unknown routes answer the root not-found face", async ({ page }) => {
  const res = await page.goto("/ghost-route-e2e");
  expect(res?.status()).toBe(404);
});
