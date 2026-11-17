import { describe, expect, it } from "vitest";
import { reactDevToolsMcp } from "./index";

type TransformFn = (
  this: { environment?: { name?: string } },
  code: string,
  id: string,
) => string | undefined;

const transformOf = () => reactDevToolsMcp().transform as unknown as TransformFn;

describe("reactDevToolsMcp", () => {
  it("prepends the register import to the dev client entry in the client environment", () => {
    const out = transformOf().call(
      { environment: { name: "client" } },
      "await import(entry)",
      "virtual:tanstack-start-dev-client-entry",
    );
    expect(out?.startsWith('import "react-devtools-cdt-mcp/register";')).toBe(true);
    expect(out?.endsWith("await import(entry)")).toBe(true);
  });

  it("leaves the module untouched in the ssr environment", () => {
    const out = transformOf().call(
      { environment: { name: "ssr" } },
      "await import(entry)",
      "virtual:tanstack-start-dev-client-entry",
    );
    expect(out).toBeUndefined();
  });

  it("leaves unrelated modules untouched", () => {
    const out = transformOf().call(
      { environment: { name: "client" } },
      "console.log(1)",
      "/src/router.tsx",
    );
    expect(out).toBeUndefined();
  });

  it("applies only when serving (dev), never on build", () => {
    const plugin = reactDevToolsMcp();
    const apply = plugin.apply as unknown as (config: never, env: { command: string }) => boolean;
    expect(apply({} as never, { command: "serve" })).toBe(true);
    expect(apply({} as never, { command: "build" })).toBe(false);
  });
});
