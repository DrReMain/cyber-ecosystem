import handler from "@tanstack/react-start/server-entry"
import { generateRobotsTxt, generateSitemap } from "#/domains/seo"
import {
  defineCustomServerStrategy,
  extractLocaleFromHeader,
  getLocale,
  toLocale,
} from "#/paraglide/runtime.js"
import { paraglideMiddleware } from "#/paraglide/server.js"

const COOKIE_NAME = "PARAGLIDE_LOCALE"

defineCustomServerStrategy("custom-smart-preferred", {
  getLocale: (request) => {
    if (!request) return undefined

    // Already visited — let URL strategy be authoritative
    if (request.headers.get("cookie")?.includes(COOKIE_NAME)) return undefined

    const url = new URL(request.url)
    const firstSegment = url.pathname.split("/").filter(Boolean)[0]
    if (firstSegment && toLocale(firstSegment)) return undefined

    // First visit, no cookie — detect from Accept-Language
    return extractLocaleFromHeader(request)
  },
})

function handleSeoRequest(req: Request): Response | null {
  const { pathname } = new URL(req.url)

  if (req.method !== "GET") return null

  if (pathname === "/robots.txt") {
    return new Response(generateRobotsTxt(), {
      headers: { "Content-Type": "text/plain; charset=utf-8" },
    })
  }

  if (pathname === "/sitemap.xml") {
    return new Response(generateSitemap(), {
      headers: { "Content-Type": "application/xml; charset=utf-8" },
    })
  }

  return null
}

export default {
  fetch(req: Request): Promise<Response> {
    const seoResponse = handleSeoRequest(req)
    if (seoResponse) return Promise.resolve(seoResponse)

    return paraglideMiddleware(req, async () => {
      const response = await handler.fetch(req)
      response.headers.append(
        "Set-Cookie",
        `${COOKIE_NAME}=${getLocale()}; Path=/; Max-Age=31536000; SameSite=Lax`,
      )
      response.headers.append("Accept-CH", "Sec-CH-Prefers-Color-Scheme")
      return response
    })
  },
}
