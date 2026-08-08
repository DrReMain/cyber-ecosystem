import type { ReactNode } from "react";
import { errorMessage } from "#/domains/i18n/error-message";
import { normalize } from "./app-error";

interface ErrorFallbackProps {
  error: unknown;
  title?: ReactNode;
  message?: ReactNode;
  fullScreen?: boolean;
  resetErrorBoundary?: () => void;
  showDetails?: boolean;
}

export function ErrorFallback({
  error,
  title = "500",
  message,
  fullScreen = false,
  resetErrorBoundary,
  showDetails = false,
}: Readonly<ErrorFallbackProps>) {
  const appError = normalize(error);
  return (
    <div
      className={`flex flex-col items-center justify-center gap-2 ${fullScreen ? "min-h-screen" : "py-8"}`}
    >
      <h1 className="font-bold text-4xl">{title}</h1>
      <p>{message ?? errorMessage(appError)}</p>
      {appError.message ? <p className="text-sm opacity-70">{appError.message}</p> : null}
      <div className="flex gap-2">
        <a className="border px-2 py-1" href="/">
          Back Home
        </a>
        {resetErrorBoundary ? (
          <button className="border px-2 py-1" onClick={resetErrorBoundary} type="button">
            Retry
          </button>
        ) : null}
      </div>
      {showDetails ? (
        <pre>{error instanceof Error ? String(error) : JSON.stringify(error, null, 2)}</pre>
      ) : null}
    </div>
  );
}
