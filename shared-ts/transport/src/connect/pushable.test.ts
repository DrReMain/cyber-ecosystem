import { describe, expect, it } from "vitest";
import { createPushable } from "./pushable";

async function collect<T>(iterable: AsyncIterable<T>): Promise<T[]> {
  const out: T[] = [];
  for await (const item of iterable) {
    out.push(item);
  }
  return out;
}

describe("createPushable", () => {
  it("yields pushed items in push order", async () => {
    const p = createPushable<number>();
    p.push(1);
    p.push(2);
    p.push(3);
    p.end();
    await expect(collect(p)).resolves.toEqual([1, 2, 3]);
  });

  it("resolves a pull that arrived before the push", async () => {
    const p = createPushable<string>();
    const first = p[Symbol.asyncIterator]().next();
    p.push("late");
    p.end();
    await expect(first).resolves.toMatchObject({ value: "late", done: false });
  });

  it("completes iteration on end", async () => {
    const p = createPushable<number>();
    p.push(1);
    p.end();
    const iterator = p[Symbol.asyncIterator]();
    await iterator.next();
    await expect(iterator.next()).resolves.toMatchObject({ done: true });
  });

  it("drops pushes after end (boundary negative)", async () => {
    const p = createPushable<number>();
    p.push(1);
    p.end();
    p.push(2);
    await expect(collect(p)).resolves.toEqual([1]);
  });

  it("ends are idempotent", async () => {
    const p = createPushable<number>();
    p.end();
    p.end();
    await expect(collect(p)).resolves.toEqual([]);
  });
});
