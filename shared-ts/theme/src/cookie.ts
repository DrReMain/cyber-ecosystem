import { readCookie, writeCookie } from "@cyber-ecosystem/shared-cookie";
import { z } from "zod";

export type ThemePreference = "light" | "dark";
export type ThemeMode = "light" | "dark" | "system";

const ThemeCookieSchema = z.object({
  skinId: z.string().default("default"),
  mode: z.enum(["light", "dark", "system"]).default("light"),
  preference: z.enum(["light", "dark"]).default("light"),
  compact: z.boolean().default(false),
});

export type ThemeCookieData = z.infer<typeof ThemeCookieSchema>;

export const THEME_COOKIE_KEY = "theme";

const THEME_DEFAULT: ThemeCookieData = Object.freeze({
  skinId: "default",
  mode: "light",
  preference: "light",
  compact: false,
});

// Pure theme resolution — the consuming app owns the server fn (composition
// root): it reads the cookie + sec-ch header and calls this. Keeps the
// package free of createServerFn / framework server glue.
export function resolveThemeData(
  raw: string | undefined,
  secChPrefersColorScheme: string | undefined,
): ThemeCookieData {
  const data = readCookie(raw, ThemeCookieSchema, THEME_DEFAULT);

  if (data.mode === "system") {
    const headerPref =
      secChPrefersColorScheme === "dark" || secChPrefersColorScheme === "light"
        ? secChPrefersColorScheme
        : undefined;
    data.preference = headerPref ?? data.preference;
  } else {
    data.preference = data.mode;
  }

  return data;
}

export function setThemeCookie(
  skinId: string,
  mode: ThemeMode,
  preference: ThemePreference,
  compact: boolean,
): void {
  writeCookie(THEME_COOKIE_KEY, { skinId, mode, preference, compact });
}
