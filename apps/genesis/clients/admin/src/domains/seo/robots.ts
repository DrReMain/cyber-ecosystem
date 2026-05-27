import { getSiteUrl } from "#/env"

export function generateRobotsTxt(): string {
  const host = getSiteUrl()
  return [
    "User-agent: *",
    "Allow: /",
    "",
    `Sitemap: ${host}/sitemap.xml`,
    "",
  ].join("\n")
}
