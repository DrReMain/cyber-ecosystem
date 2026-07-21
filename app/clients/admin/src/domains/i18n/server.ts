import {
  cookieName,
  defineCustomServerStrategy,
  extractLocaleFromHeader,
  type Locale,
  toLocale,
} from "#/paraglide/runtime.js";

export type LocaleSource = "url" | "cookie" | "accept-language" | "none";

export function readLocaleCookie(request: Request): Locale | undefined {
  const header = request.headers.get("cookie");
  if (!header) return undefined;
  for (const part of header.split(";")) {
    const [name, ...rest] = part.trim().split("=");
    if (name === cookieName) return toLocale(rest.join("="));
  }
  return undefined;
}

export function hasLocalePrefix(request: Request): boolean {
  const segment = new URL(request.url).pathname.split("/").filter(Boolean)[0];
  return !!segment && !!toLocale(segment);
}

export function resolveLocaleSource(request: Request): LocaleSource {
  if (hasLocalePrefix(request)) return "url";
  if (readLocaleCookie(request)) return "cookie";
  if (extractLocaleFromHeader(request)) return "accept-language";
  return "none";
}

// custom-smart-preferred: detection-only on a bare "/".
// - Explicit URL prefix → undefined (url strategy authoritative).
// - Bare "/" with cookie → undefined (already visited; let baseLocale
//   apply so setLocale(baseLocale) isn't reverted by a stale cookie).
// - Bare "/" first visit → Accept-Language detection.
defineCustomServerStrategy("custom-smart-preferred", {
  getLocale: (request) => {
    if (!request) return undefined;
    if (hasLocalePrefix(request)) return undefined;
    if (readLocaleCookie(request)) return undefined;
    return extractLocaleFromHeader(request);
  },
});
