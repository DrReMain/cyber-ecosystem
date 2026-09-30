import type { ErrorDomain } from "@cyber-ecosystem/shared-error";

let uninstall: (() => void) | undefined;

export function installWindowCapture(domain: ErrorDomain): () => void {
  if (typeof window === "undefined" || uninstall !== undefined) {
    return () => {};
  }
  const onError = (event: ErrorEvent) => {
    if (event.error) {
      domain.capture(event.error);
    }
  };
  const onRejection = (event: PromiseRejectionEvent) => {
    domain.capture(event.reason);
  };
  window.addEventListener("error", onError);
  window.addEventListener("unhandledrejection", onRejection);
  uninstall = () => {
    window.removeEventListener("error", onError);
    window.removeEventListener("unhandledrejection", onRejection);
  };
  return uninstall;
}
