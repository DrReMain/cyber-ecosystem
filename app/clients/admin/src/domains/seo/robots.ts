import { getSiteUrl } from "#/env";

export function generateRobotsTxt(options?: { disallow?: string[] }): string {
  const host = getSiteUrl();
  const lines = ["User-agent: *", "Allow: /"];
  for (const path of options?.disallow ?? []) {
    lines.push(`Disallow: ${path}`);
  }
  lines.push("", `Sitemap: ${host}/sitemap.xml`, "");
  return lines.join("\n");
}
