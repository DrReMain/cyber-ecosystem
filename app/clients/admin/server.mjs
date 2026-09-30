import { createReadStream } from "node:fs";
import { stat } from "node:fs/promises";
import { createServer } from "node:http";
import { extname, join, normalize } from "node:path";
import { Readable } from "node:stream";
import { fileURLToPath } from "node:url";
import handler from "./dist/server/server.js";

const port = Number(process.env.PORT ?? 3000);
const clientRoot = fileURLToPath(new URL("./dist/client/", import.meta.url));

const MIME_TYPES = {
  ".css": "text/css; charset=utf-8",
  ".ico": "image/x-icon",
  ".js": "text/javascript; charset=utf-8",
  ".json": "application/json; charset=utf-8",
  ".map": "application/json",
  ".mjs": "text/javascript; charset=utf-8",
  ".png": "image/png",
  ".svg": "image/svg+xml",
  ".webp": "image/webp",
  ".woff": "font/woff",
  ".woff2": "font/woff2",
};

// Vite emits content-hashed filenames under assets/, so those are immutable.
// Everything else (favicon, icons) revalidates on every request.
function cacheControl(pathname) {
  return pathname.startsWith("/assets/") ? "public, max-age=31536000, immutable" : "no-cache";
}

function safeDecode(pathname) {
  try {
    return decodeURIComponent(pathname);
  } catch {
    return pathname; // malformed sequences fall through to the SSR handler
  }
}

async function tryServeStatic(pathname, res) {
  const target = join(clientRoot, normalize(`.${safeDecode(pathname)}`));
  if (!target.startsWith(clientRoot)) return false;
  const file = await stat(target).catch(() => null);
  if (!file?.isFile()) return false;
  res.writeHead(200, {
    "content-type": MIME_TYPES[extname(target)] ?? "application/octet-stream",
    "content-length": file.size,
    "cache-control": cacheControl(pathname),
  });
  createReadStream(target).pipe(res);
  return true;
}

function requestBaseUrl(req) {
  const proto = req.headers["x-forwarded-proto"] ?? "http";
  const host = req.headers.host ?? "localhost";
  return `${proto}://${host}`;
}

const server = createServer(async (req, res) => {
  try {
    const url = new URL(req.url ?? "/", requestBaseUrl(req));
    if (
      (req.method === "GET" || req.method === "HEAD") &&
      (await tryServeStatic(url.pathname, res))
    ) {
      return;
    }
    const hasBody = req.method !== "GET" && req.method !== "HEAD";
    const request = new Request(url, {
      method: req.method,
      headers: req.headers,
      body: hasBody ? Readable.toWeb(req) : undefined,
      duplex: hasBody ? "half" : undefined,
    });
    const response = await handler.fetch(request);
    const headers = Object.fromEntries(response.headers);
    const setCookies = response.headers.getSetCookie();
    if (setCookies.length > 0) headers["set-cookie"] = setCookies;
    res.writeHead(response.status, headers);
    if (!response.body) {
      res.end();
      return;
    }
    Readable.fromWeb(response.body).pipe(res);
  } catch (error) {
    console.error("[admin] request failed:", error);
    if (!res.headersSent) res.writeHead(500, { "content-type": "text/plain" });
    res.end("Internal Server Error");
  }
});

server.listen(port, () => {
  console.log(`[admin] listening on :${port}`);
});
