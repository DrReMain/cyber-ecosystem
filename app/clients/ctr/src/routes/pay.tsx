import { createFileRoute, Outlet } from "@tanstack/react-router";
import { sessionGuard } from "#/features/auth/guard";

// Layout parent for the /pay tree: both the wall and the result page sit
// behind the session guard once, here; each child owns its head and URL
// shape.
export const Route = createFileRoute("/pay")({
  beforeLoad: sessionGuard,
  component: () => <Outlet />,
});
