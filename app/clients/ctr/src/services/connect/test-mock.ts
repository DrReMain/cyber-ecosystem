// Dev-only test-service mock at the transport seam (the roadmap's "mock 落
// transport 缝" — this file is its first landing). Serves the frozen read
// face (ListTests/GetTest/GetTestConfig) from the seed baseline so the game
// engine runs without a backend; the submit seam stays Unimplemented until
// E0 freezes. Never active outside DEV (see transport.ts).
//
// Field-shape note: this service's payloads are int32/string/bool/double
// and map<string,string> only — shapes whose Message runtime equals the
// toJson shape the real transport returns, so consumers see identical
// fields either way. Timestamp fields (createdAt/updatedAt) are not served
// and not consumed by any client face.

import type {
  DescMessage,
  DescMethodStreaming,
  DescMethodUnary,
  MessageInitShape,
} from "@bufbuild/protobuf";
import type { ContextValues, StreamResponse, Transport, UnaryResponse } from "@connectrpc/connect";
import { Code, ConnectError, createRouterTransport } from "@connectrpc/connect";
import { TestService } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/test_pb";
import {
  getTest,
  getTestConfig,
  listTests,
  submitTestAttempt,
} from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/test-TestService_connectquery";

// The seed baseline (esports bootstrap seedTests), verbatim; ids are
// deterministic 20-char stand-ins matching the contract's id length.
interface SeedRow {
  readonly id: string;
  readonly code: string;
  readonly name: string;
  readonly category: string;
  readonly ability: string;
}

const seed = (
  [
    ["red_dot", "随机红点触摸测试", "universal", "reactionSpeed"],
    ["number_stars", "数字星空", "universal", "reactionSpeed"],
    ["dual_zone", "双区同步躲避", "universal", "reactionSpeed"],
    ["target_recognition", "目标识别测试", "universal", "reactionSpeed"],
    ["flash_aiming", "闪现瞄准测试", "fps", "aimingAccuracy"],
    ["moving_target", "移动靶射击测试", "fps", "trackingStability"],
    ["tracking_moving", "追踪移动目标测试", "fps", "trackingStability"],
    ["skill_trajectory", "技能轨迹测试", "universal", "skillReleaseAccuracy"],
    ["offense_defense", "攻防决策测试", "moba", "decisionEfficiency"],
    ["resource_match", "资源配对测试", "moba", "resourceAllocation"],
    ["character_counter", "角色Counter测试", "moba", "resourceAllocation"],
    ["economic_sim", "经济模拟测试", "moba", "resourceAllocation"],
    ["multi_target", "多目标点击测试", "fps", "aimingAccuracy"],
  ] as const
).map<SeedRow>(([code, name, category, ability]) => ({
  id: code.padEnd(20, "0"),
  code,
  name,
  category,
  ability,
}));

const RED_DOT_INTRODUCTION =
  "红点出现在屏幕任意位置，尽可能快地点击它的中心。点错位置或点空都会记录为失误，并立即刷新下一个红点。";

// red_dot engine params, values = the legacy-calibrated constants (the
// compiled-in defaults the game falls back to anyway — the KV makes the
// config-driven path itself observable under mock).
const RED_DOT_ENGINE_PARAMS: Record<string, string> = {
  spawn_padding_px: "50",
  target_lifetime_ms: "1500",
  target_size_px: "20",
};

// Test-service-only interception: with a fallback transport wired in (the
// real one), every other service — session, auth, profile — passes
// through untouched, so mock mode keeps a fully live login/profile chain
// and swaps only the test read face.
export function createTestMockTransport(fallback?: Transport): Transport {
  const routerTransport = createRouterTransport((router) => {
    router.service(TestService, {
      listTests: () => ({
        tests: seed.map((row) => ({ ...row, durationSeconds: 30 })),
      }),
      getTest: (req) => {
        const row = seed.find((t) => t.id === req.id);
        if (row === undefined) {
          throw new ConnectError(`mock: unknown test id ${req.id}`, Code.NotFound);
        }
        return {
          test: { ...row, durationSeconds: 30 },
          ...(row.code === "red_dot" ? { introduction: RED_DOT_INTRODUCTION } : {}),
        };
      },
      getTestConfig: (req) => {
        const row = seed.find((t) => t.id === req.id);
        if (row === undefined) {
          throw new ConnectError(`mock: unknown test id ${req.id}`, Code.NotFound);
        }
        return {
          config: {
            testId: row.id,
            durationSeconds: 30,
            backgroundInvalidSeconds: 3,
            engineParams: row.code === "red_dot" ? RED_DOT_ENGINE_PARAMS : {},
          },
        };
      },
      submitTestAttempt: () => {
        throw new ConnectError("mock: submit seam deferred until E0 freezes", Code.Unimplemented);
      },
    });
  });
  if (fallback === undefined) {
    return routerTransport;
  }
  // Dispatch by descriptor identity: connect-query hands the very method
  // objects this module imports, so a Set membership check routes the test
  // face to the mock and everything else through untouched.
  const mocked = new Set<unknown>([listTests, getTest, getTestConfig, submitTestAttempt]);
  const unary = <I extends DescMessage, O extends DescMessage>(
    method: DescMethodUnary<I, O>,
    signal: AbortSignal | undefined,
    timeoutMs: number | undefined,
    header: HeadersInit | undefined,
    input: MessageInitShape<I>,
    contextValues?: ContextValues,
  ): Promise<UnaryResponse<I, O>> =>
    mocked.has(method)
      ? routerTransport.unary(method, signal, timeoutMs, header, input, contextValues)
      : fallback.unary(method, signal, timeoutMs, header, input, contextValues);
  const stream = <I extends DescMessage, O extends DescMessage>(
    method: DescMethodStreaming<I, O>,
    signal: AbortSignal | undefined,
    timeoutMs: number | undefined,
    header: HeadersInit | undefined,
    input: AsyncIterable<MessageInitShape<I>>,
    contextValues?: ContextValues,
  ): Promise<StreamResponse<I, O>> =>
    fallback.stream(method, signal, timeoutMs, header, input, contextValues);
  return { stream, unary } satisfies Transport;
}
