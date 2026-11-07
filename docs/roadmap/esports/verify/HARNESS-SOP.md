# 保真 harness SOP(封存档)

> 2026-10-28:login 样板验证通过后,按「先开发、后验证」的串行计划从 app 树撤下。本档是三件套的唯一存档——源码由脚本从活文件原样嵌入,**重激活只允许照抄落盘,禁止凭记忆重写**。
> 重激活时机:esports 开发阶段完成后。语料根:`ESPORTS_PROTO_DIR`(默认 lab 路径,见 gen spec);状态直达:原型 URL `?preset=<name>`(预设清单在 `_state.js` preset())。

## 激活步骤

1. `mkdir -p app/clients/ctr/e2e/prototype/fixtures`
2. 按下文三节把源码原样落盘(文件名见节标题)
3. `project.json` 的 `targets` 加入下文 fixtures 块
4. 仓根 `./nx run ctr:fixtures` —— 再生锚(门控 GEN_FIXTURES + 链尾 biome format,生成物出生即合规)
5. 仓根 `./nx run ctr:e2e` —— fidelity 常红=预期(gap 台账在案);gen spec 无环境变量自动 skip

## 三件套源码(机器嵌入,2026-10-28 版)

### `app/clients/ctr/e2e/prototype/text.mjs`

````mjs
// Shared text normalization for prototype-fidelity checks: the prototype
// corpus and the running app must go through the exact same shape here, so
// both sides import this module instead of re-implementing the rules.
export function normalizeLines(raw) {
  return String(raw ?? "")
    .split("\n")
    .map((l) => l.replace(/\s+/g, " ").trim())
    .filter((l) => l.length > 0);
}
````
### `app/clients/ctr/e2e/prototype/gen-fixtures.spec.ts`

````ts
import { mkdirSync, writeFileSync } from "node:fs";
import { fileURLToPath } from "node:url";
import { test } from "@playwright/test";
import { normalizeLines } from "./text.mjs";

// Regenerates the prototype text fixtures — the literal oracle the fidelity
// specs assert against. Opt-in via GEN_FIXTURES=1 (nx target `fixtures`);
// a normal e2e run skips this file, so the oracle only moves on purpose.
// Reviewer-only chrome (version badge, logic-check gear + drawer, env
// switch) is stripped; the bottom tabbar is product UI and stays.
const PROTO =
  process.env.ESPORTS_PROTO_DIR ??
  "/Users/shijiao/Desktop/lab/esports/manage/design-systems/ctr/prototype";
const REVIEW_CHROME = ".vchip, .ckbtn, #__ck, .envsw";
const PAGES = ["login", "home?preset=ready", "home?preset=resume", "home?preset=tested-fps"];

test.skip(!process.env.GEN_FIXTURES, "fixture regeneration is opt-in (GEN_FIXTURES=1)");

test("regenerate prototype text fixtures", async ({ browser }) => {
  const outDir = fileURLToPath(new URL("./fixtures/", import.meta.url));
  mkdirSync(outDir, { recursive: true });
  for (const spec of PAGES) {
    const [pageName, query = ""] = spec.split("?");
    const url = `file://${PROTO}/${pageName}.html${query ? `?${query}` : ""}`;
    const context = await browser.newContext(); // fresh storage per page
    const page = await context.newPage();
    await page.goto(url, { waitUntil: "load" });
    await page.waitForTimeout(300);
    await page.evaluate((sel) => {
      document.querySelectorAll(sel).forEach((n) => {
        n.remove();
      });
    }, REVIEW_CHROME);
    const title = await page.title();
    // innerText drops input placeholders, so they get their own channel;
    // hidden inputs (e.g. the inactive login tab's field) are excluded to
    // match what innerText does for text nodes.
    const { raw, placeholders } = await page.evaluate(() => ({
      // biome-ignore lint/performance/useDomNodeTextContent: innerText is the deliberate oracle — visibility-respecting text is what fidelity asserts; textContent would admit hidden nodes and script bodies.
      raw: document.body.innerText,
      placeholders: [
        ...document.querySelectorAll<HTMLInputElement | HTMLTextAreaElement>("input, textarea"),
      ]
        .filter((el) => el.getClientRects().length > 0)
        .map((el) => el.placeholder)
        .filter(Boolean),
    }));
    const preset = new URLSearchParams(query).get("preset");
    const file = `${pageName}${preset ? `__${preset}` : ""}.json`;
    writeFileSync(
      `${outDir}${file}`,
      `${JSON.stringify(
        { page: pageName, preset, source: url, title, lines: normalizeLines(raw), placeholders },
        null,
        2,
      )}\n`,
    );
    await context.close();
  }
});
````
### `app/clients/ctr/e2e/prototype/login.fidelity.spec.ts`

````ts
import { readFileSync } from "node:fs";
import { expect, test } from "@playwright/test";
import { normalizeLines } from "./text.mjs";

// The oracle corpus is zh-CN and the app negotiates language, so the
// comparison pins the locale instead of inheriting the runner default —
// an en-US runner would diff against the app's English face, not the
// product's target face.
test.use({ locale: "zh-CN" });

interface Fixture {
  page: string;
  preset: string | null;
  source: string;
  title: string;
  lines: string[];
  placeholders: string[];
}

const fixture = JSON.parse(
  readFileSync(new URL("./fixtures/login.json", import.meta.url), "utf8"),
) as Fixture;

// The oracle is the machine snapshot of the prototype page; expected copy
// comes only from that snapshot, never from the implementation, so every
// divergence below is a reported gap, not an adjusted expectation.
test("login text matches the prototype snapshot (line-for-line)", async ({ page }) => {
  await page.goto("/login");
  await page.waitForLoadState("load");
  await page.waitForTimeout(400);
  const { raw, placeholders } = await page.evaluate(() => ({
    // biome-ignore lint/performance/useDomNodeTextContent: innerText is the deliberate oracle — visibility-respecting text is what fidelity asserts; textContent would admit hidden nodes and script bodies.
    raw: document.body.innerText,
    placeholders: [
      ...document.querySelectorAll<HTMLInputElement | HTMLTextAreaElement>("input, textarea"),
    ]
      .filter((el) => el.getClientRects().length > 0)
      .map((el) => el.placeholder)
      .filter(Boolean),
  }));
  const actual = normalizeLines(raw);

  const missing = multisetMinus(fixture.lines, actual);
  const extra = multisetMinus(actual, fixture.lines);
  const missingPh = multisetMinus(fixture.placeholders, placeholders);
  const extraPh = multisetMinus(placeholders, fixture.placeholders);
  const sameOrder = JSON.stringify(actual) === JSON.stringify(fixture.lines);
  const appTitle = await page.title();

  expect(
    missing.length === 0 &&
      extra.length === 0 &&
      missingPh.length === 0 &&
      extraPh.length === 0 &&
      sameOrder &&
      appTitle === fixture.title,
    [
      `title: prototype=${JSON.stringify(fixture.title)} app=${JSON.stringify(appTitle)}`,
      `missing lines: ${JSON.stringify(missing)}`,
      `extra lines: ${JSON.stringify(extra)}`,
      `missing placeholders: ${JSON.stringify(missingPh)}`,
      `extra placeholders: ${JSON.stringify(extraPh)}`,
      `orderMatch: ${sameOrder}`,
    ].join("\n"),
  ).toBe(true);
});

function multisetMinus(a: string[], b: string[]): string[] {
  const pool = new Map<string, number>();
  for (const line of b) pool.set(line, (pool.get(line) ?? 0) + 1);
  const out: string[] = [];
  for (const line of a) {
    const n = pool.get(line) ?? 0;
    if (n === 0) out.push(line);
    else pool.set(line, n - 1);
  }
  return out;
}
````

## project.json target

````json
"fixtures": {
  "executor": "nx:run-commands",
  "options": {
    "cwd": "{projectRoot}",
    "command": "GEN_FIXTURES=1 pnpm exec playwright test e2e/prototype/gen-fixtures.spec.ts && pnpm exec biome format --write e2e/prototype/fixtures"
  }
}
````

## 陷阱清单(全部实测踩过,激活后必看)

- 比对 spec 必须 `test.use({ locale: "zh-CN" })`:跑者默认 en-US,应用按语言协商渲染英文面=整页假 gap。
- biome-ignore 豁免必须**单行**紧贴违规行;多行注释块会隔断失效(suppressions/unused)。
- `querySelectorAll<HTMLInputElement | HTMLTextAreaElement>(...)` 显式泛型,否则 tsc 报 Element 无 placeholder。
- innerText 是刻意选的 oracle(可见性语义),已带豁免;textContent 会吞隐藏节点与脚本体,不可换。
- 提取剥离评审壳:`.vchip/.ckbtn/#__ck/.envsw`;底部 tabbar 是产品件保留。placeholder 走独立通道(innerText 不含),只收可见输入框。
- **app 侧对称剥离 dev 壳**(11-08 home 片立):TailwindIndicator(`[data-dev-chrome]`,收起态渲染断点名 "xl")与 TanStack Devtools 开关按钮会污染 oracle——收集时同剥,与原型侧 REVIEW_CHROME 同法;login 红组因此 4→3(xl 噪声清除)。
- **含 `$` 的 SQL(argon2id 哈希 `$argon2id$v=19$…`)必须走 stdin(`docker exec -i … -c` 禁)**:双引号 -c 里 `$argon2id` 等被 shell 当变量展开吃成空串,登录报 password mismatch(11-08 实锤)。
- **profile.games 的 game_id 是目录 code 命名空间,不是 ent 行主键**(11-08 profile 片实锤):ListGames 的 id 即 code(cs2/honor/…),UpdateProfile 的 ExistingCodes 校验并写 code;种子若写行主键,wizard 游戏步预选与类型投票**静默**失配(行数对、票全归「未归类」,无任何报错)。处方=`jsonb_build_object('game_id', code)`。
- **cohort 匹配键=birth year+sex(月与 age_band 不参与;report/data.go Cohort)**(11-08 实锤):共享 dev 池里同年同性别的历史行会入池——冻结百分位的 spec 必须选无碰撞的出生年(本机有他人 2008 male 校准批 wtfixuser1-9,把冻结 78 挤成 93;种子年改 2004-06,band 列原样存 18-22 仍诚实,oracle 无出生行故面零变化)。
- **时钟喂给的枚举行须钉时钟**:年龄步年份下拉(1966-2020)来自原型硬编码 nowY=2026,app 读真实时钟 `new Date()`——跨年即整列漂移;测试 `page.clock.install` 钉 2026(resume 片已立的同法)。
- gen-fixtures 机制注(11-08 profile 片):PAGES 条目支持 `#key` 显式 fixture 后缀(同 preset 多态页如 password ?m= 三态命名);fixture 的 source 字段归一为相对 `page.html?query`(双机再生成零 diff)。
- 写文件一律绝对路径+落完即 ls:并行 tool 调用 cwd 快照可能不一致(实测落进幽灵嵌套路径)。
- 退出码重定向到文件后单独取,禁管道尾(`|tail;$?` 恒 0)。
- `pnpm exec node` 被本环境策略拒绝:提取一律走 Playwright spec 形态(本档三件套即是)。

## 样板验证记录(2026-10-28)

login:e2e 红=9 条 gap(gap-ledger 首批,红 8+按态比对 1),biome check 0 错,tsc 0 错,fixtures target 绿。
铺开序按 `ctr-prototype-matrix.md`:先 ✅/◐ 19 页逐页逐预设,❌ 页只出红条。E5-f3(0ae81a28)已落 home/profile/report 面,重激活后 home 样板可直接跑。
