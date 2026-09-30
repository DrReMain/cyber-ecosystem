import { createCsrfMiddleware, createMiddleware, createStart } from "@tanstack/react-start";
import { setResponseHeader } from "@tanstack/react-start/server";
import { setupErrorDomain } from "#/domains/error";

const csrfMiddleware = createCsrfMiddleware({
  filter: (ctx) => ctx.handlerType === "serverFn",
});

const securityHeadersMiddleware = createMiddleware().server(async (ctx) => {
  setResponseHeader(
    "Content-Security-Policy",
    "frame-ancestors 'none'; object-src 'none'; base-uri 'self'",
  );
  setResponseHeader("X-Content-Type-Options", "nosniff");
  setResponseHeader("Referrer-Policy", "same-origin");
  return ctx.next();
});

export const startInstance = createStart(() => {
  setupErrorDomain();
  return {
    requestMiddleware: [securityHeadersMiddleware, csrfMiddleware],
  };
});
