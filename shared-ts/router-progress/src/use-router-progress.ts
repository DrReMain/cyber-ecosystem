import { useRouterState } from "@tanstack/react-router";
import type { RouterProgressConfig } from "./config";
import { useProgressController } from "./use-progress-controller";

/** Connects the progress controller to TanStack Router's loading state. */
export function useRouterProgress(config: RouterProgressConfig) {
  const isLoading = useRouterState({ select: (state) => state.isLoading });
  return useProgressController(config, isLoading);
}
