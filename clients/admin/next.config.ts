import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  /* config options here */
  reactCompiler: true,
  typedRoutes: true,
  poweredByHeader: false,
  output: "standalone",
};

export default nextConfig;
