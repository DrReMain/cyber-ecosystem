export interface SitemapRoute {
  path: string
  changefreq?:
    | "always"
    | "hourly"
    | "daily"
    | "weekly"
    | "monthly"
    | "yearly"
    | "never"
  priority?: number
}

export const sitemapRoutes: SitemapRoute[] = [
  {
    path: "/",
    changefreq: "weekly",
    priority: 1.0,
  },
  {
    path: "/playground",
    changefreq: "monthly",
    priority: 0.5,
  },
]
