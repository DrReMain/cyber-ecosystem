import { useEffect } from "react";
import { errorHandler } from "./error-handler";

export function ErrorListeners() {
  useEffect(() => {
    const onError = (e: ErrorEvent) => {
      errorHandler.handle(e.error ?? e.message, "window");
    };
    const onRejection = (e: PromiseRejectionEvent) => {
      errorHandler.handle(e.reason, "rejection");
    };
    window.addEventListener("error", onError);
    window.addEventListener("unhandledrejection", onRejection);
    return () => {
      window.removeEventListener("error", onError);
      window.removeEventListener("unhandledrejection", onRejection);
    };
  }, []);

  return null;
}
