import { getRequest, getRequestHeader, setResponseHeader } from "@tanstack/react-start/server";

// Mirrors shared-go/kratos/security/auth.SessionCookieName — the server-side
// extractor recognizes exactly that spelling; rename the two together.
const SESSION_COOKIE = "session_token";

function isSecureRequest(): boolean {
  const forwarded = getRequestHeader("x-forwarded-proto");
  if (forwarded !== undefined) {
    return forwarded === "https";
  }
  return new URL(getRequest().url).protocol === "https:";
}

function cookieAttributes(secure: boolean): string {
  return ["HttpOnly", "Path=/", "SameSite=Lax", secure ? "Secure" : ""].filter(Boolean).join("; ");
}

function cookieValue(name: string): string | null {
  const header = getRequestHeader("cookie");
  if (header === undefined) {
    return null;
  }
  for (const part of header.split(/;\s*/)) {
    const eq = part.indexOf("=");
    if (eq === -1) {
      continue;
    }
    if (part.slice(0, eq) === name) {
      const value = part.slice(eq + 1);
      return value === "" ? null : value;
    }
  }
  return null;
}

export function save(token: string): void {
  setResponseHeader(
    "Set-Cookie",
    `${SESSION_COOKIE}=${token}; ${cookieAttributes(isSecureRequest())}`,
  );
}

export function clear(): void {
  setResponseHeader(
    "Set-Cookie",
    `${SESSION_COOKIE}=; ${cookieAttributes(isSecureRequest())}; Max-Age=0`,
  );
}

export function loadSession(): string | null {
  return cookieValue(SESSION_COOKIE);
}
