export interface RobotsConfig {
  mode: "public" | "private";
  /** Absolute site origin without a trailing slash. */
  siteUrl: string;
  disallowPrefixes: readonly string[];
  /** All supported locales; every localized URL form of a path is disallowed. */
  locales: readonly string[];
  /** Maps an absolute URL to its localized counterpart. */
  localize: (absoluteUrl: string, locale: string) => string;
}

export function generateRobotsTxt(config: RobotsConfig): string {
  if (config.mode === "private") {
    return "User-agent: *\nDisallow: /\n";
  }

  const lines = ["User-agent: *", "Allow: /"];
  for (const prefix of config.disallowPrefixes) {
    const normalized = prefix.startsWith("/") ? prefix : `/${prefix}`;
    for (const locale of config.locales) {
      const path = new URL(config.localize(`${config.siteUrl}${normalized}`, locale)).pathname;
      lines.push(`Disallow: ${path}`);
    }
  }
  lines.push("", `Sitemap: ${config.siteUrl}/sitemap.xml`, "");
  return lines.join("\n");
}
