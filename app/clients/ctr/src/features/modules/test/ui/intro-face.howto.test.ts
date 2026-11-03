import { describe, expect, it } from "vitest";
import { howToNodes } from "./intro-face";

// The seeded how-to copy carries the prototype's <b> markers (its own
// authoring format). The renderer turns exactly those tokens into bold
// spans and leaves every other character — stray tags included — literal:
// nothing the server sends is ever parsed as markup.
describe("howToNodes", () => {
  it("renders marker-free copy as one plain text node", () => {
    expect(howToNodes("按屏幕提示完成操作。")).toEqual(["按屏幕提示完成操作。"]);
  });

  it("bolds the marked phrase and keeps the sentence order", () => {
    const nodes = howToNodes("屏幕会出现目标，出现后<b>尽快点击</b>。越快越好。");
    expect(nodes).toHaveLength(3);
    expect(nodes[0]).toBe("屏幕会出现目标，出现后");
    expect((nodes[1] as { props: { children: string } }).props.children).toBe("尽快点击");
    expect(nodes[2]).toBe("。越快越好。");
  });

  it("leaves stray tags as literal text", () => {
    expect(howToNodes("<i>斜体</i>")).toEqual(["<i>斜体</i>"]);
  });

  it("tolerates an unmatched opener by bolding the tail", () => {
    const nodes = howToNodes("前缀<b>尾巴");
    expect(nodes[0]).toBe("前缀");
    expect((nodes[1] as { props: { children: string } }).props.children).toBe("尾巴");
  });
});
