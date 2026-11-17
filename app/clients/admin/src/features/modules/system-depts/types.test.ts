import type { Dept } from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/dept_pb";
import fc from "fast-check";
import { describe, expect, it } from "vitest";
import { buildDeptTree, buildParentTree, collectRowKeys, type DeptNode } from "./types";

const idsOf = (nodes: DeptNode[]): string[] => nodes.flatMap((n) => [n.id, ...idsOf(n.children)]);

// Wire rows arrive protojson-shaped; the cast is the transport family's
// standing reality, not a shortcut.
const bare = (id: string, parentId?: string) => ({ id, name: id, parentId }) as Dept;

describe("buildDeptTree", () => {
  it("conserves the node set: every declared id appears exactly once in the forest", () => {
    fc.assert(
      fc.property(
        fc.uniqueArray(fc.string({ minLength: 1, maxLength: 4 }), { minLength: 1, maxLength: 10 }),
        (ids) => {
          const list = ids.map((id, i) => ({
            id,
            name: `n${i}`,
            parentId: undefined as string | undefined,
          }));
          const roots = buildDeptTree(list);
          const seen: string[] = [];
          const walk = (nodes: ReturnType<typeof buildDeptTree>) => {
            for (const n of nodes) {
              seen.push(n.id);
              walk(n.children);
            }
          };
          walk(roots);
          expect(seen.sort()).toEqual([...ids].sort());
        },
      ),
    );
  });

  it("nests by parentId; orphans and self-parents root themselves", () => {
    const tree = buildDeptTree([
      bare("root"),
      bare("child", "root"),
      bare("orphan"),
      bare("self", "self"),
    ]);
    expect(tree.map((n) => n.id).sort()).toEqual(["orphan", "root", "self"]);
    expect(tree.find((n) => n.id === "root")?.children.map((c) => c.id)).toEqual(["child"]);
  });
});

describe("buildParentTree", () => {
  const chain = (n: number): Dept[] =>
    Array.from({ length: n }, (_, i) => bare(`d${i}`, i > 0 ? `d${i - 1}` : undefined));

  it("without an exclusion it is the plain tree", () => {
    expect(buildParentTree(chain(3))).toEqual(buildDeptTree(chain(3)));
  });

  it("drops the excluded node and everything below it, keeps the ancestors", () => {
    const out = buildParentTree(chain(4), "d2");
    expect(idsOf(out)).toEqual(["d0", "d1"]);
  });

  it("survives a parent cycle without losing the unrelated branches", () => {
    const cyclic: Dept[] = [bare("a", "b"), bare("b", "a"), bare("c")];
    const out = buildParentTree(cyclic, "a");
    expect(out.map((n) => n.id)).toEqual(["c"]);
  });
});

describe("collectRowKeys", () => {
  it("walks depth-first over the row forest", () => {
    const rows = [
      {
        dept: { id: "1" },
        children: [
          { dept: { id: "1.1" } },
          { dept: { id: "1.2" }, children: [{ dept: { id: "1.2.1" } }] },
        ],
      },
      { dept: { id: "2" } },
    ];
    expect(collectRowKeys(rows as never)).toEqual(["1", "1.1", "1.2", "1.2.1", "2"]);
  });
});
