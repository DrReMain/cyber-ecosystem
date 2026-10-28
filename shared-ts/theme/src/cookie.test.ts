import { writeCookie } from "@cyber-ecosystem/shared-cookie";
import { describe, expect, it } from "vitest";
import { resolveThemeData, setThemeCookie, THEME_COOKIE_KEY, THEME_DEFAULT } from "./cookie";

// resolveThemeData is the pure seam the composition root feeds: the schema
// fills defaults on partial cookies, "system" defers to the client hint,
// and an explicit mode wins over the header.

const rawOf = (key: string) => {
  const pair = document.cookie.split("; ").find((c) => c.startsWith(`${key}=`));
  return pair ? decodeURIComponent(pair.slice(key.length + 1)) : undefined;
};

describe("resolveThemeData", () => {
  it("fills every default on an empty cookie", () => {
    expect(resolveThemeData(undefined, undefined)).toEqual(THEME_DEFAULT);
  });

  it("hydrates partial cookies through schema defaults", () => {
    writeCookie(THEME_COOKIE_KEY, { mode: "dark" });
    const data = resolveThemeData(rawOf(THEME_COOKIE_KEY), undefined);
    expect(data).toMatchObject({
      skinId: "default",
      mode: "dark",
      preference: "dark",
      compact: false,
    });
  });

  it("system mode follows the client hint", () => {
    setThemeCookie("nebula", "system", "light", true);
    expect(resolveThemeData(rawOf(THEME_COOKIE_KEY), "dark")).toMatchObject({
      skinId: "nebula",
      preference: "dark",
      compact: true,
    });
  });

  it("system mode without a usable hint keeps the stored preference", () => {
    setThemeCookie("default", "system", "dark", false);
    expect(resolveThemeData(rawOf(THEME_COOKIE_KEY), undefined).preference).toBe("dark");
    expect(resolveThemeData(rawOf(THEME_COOKIE_KEY), "nocturne").preference).toBe("dark");
  });

  it("an explicit mode wins over the header", () => {
    setThemeCookie("default", "light", "light", false);
    expect(resolveThemeData(rawOf(THEME_COOKIE_KEY), "dark").preference).toBe("light");
  });

  it("round-trips setThemeCookie through the reader the server uses", () => {
    setThemeCookie("nebula", "dark", "dark", true);
    expect(resolveThemeData(rawOf(THEME_COOKIE_KEY), undefined)).toEqual({
      skinId: "nebula",
      mode: "dark",
      preference: "dark",
      compact: true,
    });
  });
});
