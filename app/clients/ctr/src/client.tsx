import { StartClient } from "@tanstack/react-start/client";
import { startTransition } from "react";
import { hydrateRoot } from "react-dom/client";

// Overrides the plugin's default client entry, whose only addition is a
// StrictMode wrapper. Dev StrictMode's simulated unmount destroys the
// first mount's query observers, and query-core cancels any signal-
// consuming fetch on last-observer removal — so every page load aborts
// its first RPC batch (three red AbortErrors per load with the
// ImageAssistant fetch wrapper logging them). The dev double-check is
// not worth that; the client chain gates carry the correctness burden.
startTransition(() => {
  hydrateRoot(document, <StartClient />);
});
