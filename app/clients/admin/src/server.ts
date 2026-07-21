import handler from "@tanstack/react-start/server-entry";
import { resolveLocaleSource } from "#/domains/i18n/server";
import { generateRobotsTxt, generateSitemap } from "#/domains/seo";
import { cookieMaxAge, cookieName, getLocale } from "#/paraglide/runtime.js";
import { paraglideMiddleware } from "#/paraglide/server.js";

function handleSeoRequest(req: Request): Response | null {
  if (req.method !== "GET") return null;

  const { pathname } = new URL(req.url);

  if (pathname === "/robots.txt") {
    return new Response(generateRobotsTxt(), {
      headers: {
        "Content-Type": "text/plain; charset=utf-8",
        "Cache-Control": "public, max-age=86400",
      },
    });
  }

  if (pathname === "/sitemap.xml") {
    return new Response(generateSitemap(), {
      headers: {
        "Content-Type": "application/xml; charset=utf-8",
        "Cache-Control": "public, max-age=3600",
      },
    });
  }

  return null;
}

export default {
  fetch(req: Request): Promise<Response> {
    const seo = handleSeoRequest(req);
    if (seo) return Promise.resolve(seo);

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
