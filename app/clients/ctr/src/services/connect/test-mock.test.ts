import { createClient } from "@connectrpc/connect";
import { TestService } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/test_pb";
import { describe, expect, it } from "vitest";
import { createTestMockTransport } from "./test-mock";

// The mock must serve the consumer runtime shape exactly: wrapper fields
// as scalars, engine_params as a plain record, duration as a number — the
// same fields the real (toJson-wrapped) transport delivers.
describe("createTestMockTransport", () => {
  const client = createClient(TestService, createTestMockTransport());

  it("lists the full seed with 20-char ids and scalar fields", async () => {
    const res = await client.listTests({});
    const redDot = res.tests.find((t) => t.code === "red_dot");
    expect(redDot?.id).toHaveLength(20);
    expect(redDot?.name).toBe("随机红点触摸测试");
    expect(redDot?.durationSeconds).toBe(30);
    expect(res.tests.length).toBe(13);
  });

  it("serves the red_dot introduction on getTest", async () => {
    const list = await client.listTests({});
    const redDotId = list.tests.find((t) => t.code === "red_dot")?.id ?? "";
    const res = await client.getTest({ id: redDotId });
    expect(res.test?.code).toBe("red_dot");
    expect(res.introduction).toContain("红点");
  });

  it("serves red_dot engine params as a plain record on getTestConfig", async () => {
    const list = await client.listTests({});
    const redDotId = list.tests.find((t) => t.code === "red_dot")?.id ?? "";
    const res = await client.getTestConfig({ id: redDotId });
    expect(res.config?.durationSeconds).toBe(30);
    expect(res.config?.backgroundInvalidSeconds).toBe(3);
    expect(res.config?.engineParams.target_size_px).toBe("20");
    expect(res.config?.engineParams.target_lifetime_ms).toBe("1500");
  });

  it("rejects unknown ids and the deferred submit seam", async () => {
    await expect(client.getTest({ id: "x".repeat(20) })).rejects.toThrow();
    await expect(client.submitTestAttempt({ testCode: "red_dot" })).rejects.toThrow();
  });
});
