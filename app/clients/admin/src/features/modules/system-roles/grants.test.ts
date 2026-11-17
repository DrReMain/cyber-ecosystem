import { create } from "@bufbuild/protobuf";
import {
  type RoleGrant,
  RoleGrantSchema,
  ScopeKind,
} from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/role_pb";
import fc from "fast-check";
import { describe, expect, it } from "vitest";

import {
  type GrantableView,
  type GrantMap,
  grantMapOf,
  opOf,
  roleGrantsOf,
  summarizeGrantMap,
  summarizeGrants,
  summarizePatterns,
  toScopeKind,
  wildcardOpOf,
} from "./grants";

// The grants face's pure core: wire-name narrowing (the enum reverse-mapping
// trap), the map↔wire round-trip with its ALL→UNSPECIFIED fold, and the
// four-state summary discriminant. Copy faces (formatGrantsSummary,
// scopeKindLabel) are walkthrough territory, not unit territory.

const grantOf = (operation: string, scope: ScopeKind, policyIds: string[] = []): RoleGrant =>
  create(RoleGrantSchema, { operation, scopeKind: scope, policyIds });

describe("toScopeKind", () => {
  it("passes numeric values straight through", () => {
    for (const kind of [
      ScopeKind.UNSPECIFIED,
      ScopeKind.ALL,
      ScopeKind.SELF,
      ScopeKind.DEPT_TREE,
    ]) {
      expect(toScopeKind(kind)).toBe(kind);
    }
  });

  it("strips the wire prefix from full proto names", () => {
    expect(toScopeKind("SCOPE_KIND_SELF")).toBe(ScopeKind.SELF);
    expect(toScopeKind("SCOPE_KIND_DEPT_TREE")).toBe(ScopeKind.DEPT_TREE);
  });

  it("accepts the short name as well (prefix-strip is a no-op on it)", () => {
    expect(toScopeKind("SELF")).toBe(ScopeKind.SELF);
    expect(toScopeKind("DEPT_TREE")).toBe(ScopeKind.DEPT_TREE);
  });

  it("answers UNSPECIFIED for anything unrecognized, never a reverse-mapped number", () => {
    for (const junk of [undefined, "", "SCOPE_KIND_WEIRD", "3"]) {
      expect(toScopeKind(junk)).toBe(ScopeKind.UNSPECIFIED);
    }
  });
});

describe("grantMapOf ∘ roleGrantsOf", () => {
  // The fold: ALL means un-narrowed on the edit face, stored as UNSPECIFIED —
  // the round-trip is the identity exactly on the stored domain.
  const scopeArb = fc.constantFrom(ScopeKind.UNSPECIFIED, ScopeKind.SELF, ScopeKind.DEPT_TREE);
  const mapArb = fc
    .uniqueArray(
      fc.tuple(fc.string({ minLength: 1 }), scopeArb, fc.array(fc.string(), { maxLength: 4 })),
      {
        selector: ([op]) => op,
        maxLength: 8,
      },
    )
    .map((rows) => {
      const map: GrantMap = {};
      for (const [operation, scope, policies] of rows) map[operation] = { scope, policies };
      return map;
    });

  it("round-trips any map over the stored scope domain", () => {
    fc.assert(
      fc.property(mapArb, (map) => {
        const wire = roleGrantsOf(map);
        const grants = wire.map((g) =>
          grantOf(g.operation, toScopeKind(g.scopeKind), [...(g.policyIds ?? [])]),
        );
        expect(grantMapOf(grants)).toEqual(map);
      }),
    );
  });

  it("folds ALL to UNSPECIFIED on intake", () => {
    expect(grantMapOf([grantOf("/svc/Op", ScopeKind.ALL)])).toEqual({
      "/svc/Op": { scope: ScopeKind.UNSPECIFIED, policies: [] },
    });
  });
});

describe("summarizeGrants / summarizePatterns", () => {
  it("classifies by the four-state discriminant for any counts", () => {
    fc.assert(
      fc.property(
        fc.integer({ min: 0, max: 50 }),
        fc.integer({ min: 0, max: 50 }),
        fc.boolean(),
        (ops, wilds, global) => {
          const total = ops + wilds;
          const out = summarizeGrants(global, total, wilds);
          if (global) {
            expect(out.kind).toBe("all");
          } else if (wilds > 0 && ops > 0) {
            expect(out).toMatchObject({ kind: "mixed", ops, wildcards: wilds });
          } else if (wilds > 0) {
            expect(out).toMatchObject({ kind: "wildcard", wildcards: wilds });
          } else {
            expect(out).toMatchObject({ kind: "count", count: ops });
          }
        },
      ),
    );
  });

  it("summarizePatterns counts non-global wildcard tails", () => {
    const out = summarizePatterns(["/system/*", "/agent/*", "/system/GetUser"]);
    expect(out).toMatchObject({ kind: "mixed", ops: 1, wildcards: 2 });
  });

  it("the global wildcard short-circuits everything", () => {
    expect(summarizePatterns(["/*", "/system/GetUser"])).toMatchObject({ kind: "all" });
  });
});

describe("summarizeGrantMap", () => {
  const catalog = [
    {
      name: "System",
      fullName: "system.System",
      comment: "",
      methods: [{ name: "GetUser" }, { name: "ListUsers" }],
    },
    { name: "Agent", fullName: "agent.Agent", comment: "", methods: [{ name: "Chat" }] },
  ] as unknown as GrantableView[];

  it("counts per-service ops and wildcards", () => {
    const grants: GrantMap = {
      [opOf("system.System", "GetUser")]: { scope: ScopeKind.SELF, policies: [] },
      [wildcardOpOf("agent.Agent")]: { scope: ScopeKind.UNSPECIFIED, policies: ["p1"] },
    };
    expect(summarizeGrantMap(catalog, grants)).toEqual({
      global: false,
      ops: 1,
      services: 2,
      wildcards: 1,
      narrowed: 1,
      attached: 1,
    });
  });

  it("the global wildcard answers services=catalog and nothing else", () => {
    const grants: GrantMap = { "/*": { scope: ScopeKind.UNSPECIFIED, policies: [] } };
    expect(summarizeGrantMap(catalog, grants)).toEqual({
      global: true,
      ops: 0,
      services: 2,
      wildcards: 0,
      narrowed: 0,
      attached: 0,
    });
  });
});
