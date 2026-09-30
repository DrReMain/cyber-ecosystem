import handler from "@tanstack/react-start/server-entry";
import { resolveLocaleSource } from "#/domains/i18n/server";
import { cookieMaxAge, cookieName, getLocale } from "#/paraglide/runtime.js";
import { paraglideMiddleware } from "#/paraglide/server.js";

export default {
  fetch(req: Request): Promise<Response> {
    const isHttps =
      new URL(req.url).protocol === "https:" || req.headers.get("x-forwarded-proto") === "https";
    const secureFlag = isHttps ? "; Secure" : "";

    return paraglideMiddleware(req, async () => {
      const response = await handler.fetch(req);
      if (resolveLocaleSource(req) !== "none") {
        response.headers.append(
          "Set-Cookie",
          `${cookieName}=${getLocale()}; Path=/; Max-Age=${cookieMaxAge}; SameSite=Lax${secureFlag}`,
        );
      }
      response.headers.append("Accept-CH", "Sec-CH-Prefers-Color-Scheme");
      return response;
    });
  },
};
