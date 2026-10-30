import { execSync } from "node:child_process";
import { createServer } from "node:http";
import { expect, type Page, test } from "@playwright/test";

// The E6c acceptance over the three identity routes, browser-driven against
// the live stack (CTR_E2E=1: postgres + redis + system + esports up):
// the sms leg reads the code the mock channel recorded, walks the born
// member through the wizard, and lands the row family; the wechat leg rides
// the endpoint-override seam — a stand-in OAuth server answers both the
// authorize hop and the token exchange, so the browser logs in with no
// WeChat and no human; the conflict leg proves the occupied-phone bind
// stops at the preview card with no code ever sent. The password leg is
// fullchain's territory and is not repeated here.
//
// Stack requirements this spec assumes (and skips with a hint when absent):
// the sms provider on the mock backend (APP_SMS_BACKEND=mock — the code
// lands in redis), and APP_WECHAT_ENDPOINT=http://127.0.0.1:18443 pointing
// at this run's stand-in.

test.skip(!process.env.CTR_E2E, "identity routes need the live stack (CTR_E2E=1)");
test.use({ locale: "zh-CN" });
test.setTimeout(180_000);

const SYSTEM_HTTP = "http://localhost:11001";
const PG = "docker exec cyber-ecosystem-postgresql-1 psql -U postgres";
const REDIS = "docker exec cyber-ecosystem-redis-1 redis-cli --raw";
const WX_PORT = 18443;
const WX_OPENID = "e2epwwxopenid0001";
const WX_CODE = "e2e-wx-code";

async function api(path: string, body: unknown, cookie?: string) {
  const res = await fetch(`${SYSTEM_HTTP}${path}`, {
    method: "POST",
    headers: { "content-type": "application/json", ...(cookie ? { cookie } : {}) },
    body: JSON.stringify(body),
  });
  return {
    status: res.status,
    json: (await res.json().catch(() => ({}))) as Record<string, unknown>,
  };
}

function sql(db: "system" | "esports", stmt: string) {
  return execSync(`${PG} -d ${db} -t -A -c "${stmt.replaceAll('"', '\\"')}"`)
    .toString()
    .trim();
}

// Erases a member end to end across both databases — the run leaves no row
// behind, the same law fullchain's finally block states.
function eraseMember(match: { byAccount?: string; byPhone?: string; byOpenid?: string }) {
  const where = match.byAccount
    ? `account='${match.byAccount}'`
    : match.byPhone
      ? `id IN (SELECT user_id FROM credential WHERE kind='sms' AND identifier='${match.byPhone}')`
      : `id IN (SELECT user_id FROM credential WHERE kind='wechat_mp' AND identifier='${match.byOpenid}')`;
  const ids = sql("system", `SELECT id FROM "user" WHERE ${where}`);
  if (!ids) return;
  const inList = ids
    .split("\n")
    .map((id) => `'${id.trim()}'`)
    .join(",");
  sql("esports", `DELETE FROM test_record WHERE user_id IN (${inList})`);
  sql("esports", `DELETE FROM match_session WHERE user_id IN (${inList})`);
  sql("esports", `DELETE FROM entitlement WHERE user_id IN (${inList})`);
  sql("esports", `DELETE FROM profile WHERE user_id IN (${inList})`);
  sql("system", `DELETE FROM membership WHERE user_id IN (${inList})`);
  sql("system", `DELETE FROM credential WHERE user_id IN (${inList})`);
  sql("system", `DELETE FROM "user" WHERE id IN (${inList})`);
}

// The three onboarding steps every born member walks before any guarded
// page lets them through.
async function walkWizard(page: Page) {
  await page.getByPlaceholder(/请输入昵称/).fill(`三路${Date.now() % 1000}`);
  await page.getByRole("button", { name: "下一步" }).click();
  await page.getByText("男", { exact: true }).click();
  const selects = page.getByRole("combobox");
  await selects.nth(0).selectOption({ index: 8 });
  await selects.nth(1).selectOption({ index: 5 });
  await page.getByRole("button", { name: "下一步" }).click();
  await page.getByText("CS2", { exact: true }).click();
  await page.getByText("无畏契约", { exact: true }).click();
  await page.getByRole("button", { name: "完成" }).click();
  await page.waitForURL("**/match", { timeout: 15_000 });
}

// The RAC pair hides its real input behind the skin — focus plus a real
// keyboard Space flips it (fullchain's finding).
async function agreeAll(page: Page) {
  const selectAll = page.getByRole("checkbox", { name: "全选" });
  await selectAll.focus();
  await page.keyboard.press(" ");
  await expect(selectAll).toBeChecked();
}

// The wechat stand-in mirrors the interface boundary: the authorize hop
// answers the redirect WeChat itself would issue (code + state echoed), the
// token endpoint answers the openid for the code it issued.
const wxServer = createServer((req, res) => {
  const url = new URL(req.url ?? "/", `http://127.0.0.1:${WX_PORT}`);
  if (url.pathname === "/connect/oauth2/authorize") {
    const cb = url.searchParams.get("redirect_uri") ?? "";
    const sep = cb.includes("?") ? "&" : "?";
    res.writeHead(302, {
      location: `${cb}${sep}code=${WX_CODE}&state=${encodeURIComponent(url.searchParams.get("state") ?? "")}`,
    });
    res.end();
    return;
  }
  if (url.pathname === "/sns/oauth2/access_token") {
    if (url.searchParams.get("code") !== WX_CODE) {
      res.end(`{"errcode":40029,"errmsg":"invalid code"}`);
      return;
    }
    res.end(`{"access_token":"t","openid":"${WX_OPENID}","errcode":0}`);
    return;
  }
  res.statusCode = 404;
  res.end();
});

test.beforeAll(async () => {
  await new Promise<void>((resolve, reject) => {
    wxServer.once("error", reject);
    wxServer.listen(WX_PORT, "127.0.0.1", () => resolve());
  });
});
test.afterAll(() => wxServer.close());

async function wechatSeamWired() {
  const { json } = await api("/api/v1/system/auth/begin-identity-login", {
    provider: "wechat",
    redirectUri: "https://ctr.local/cb",
  });
  return String(json.authorizeUrl ?? "").startsWith(`http://127.0.0.1:${WX_PORT}/`);
}

// --- the sms leg -----------------------------------------------------------------------------

test("sms login: send → redis code → login → wizard → member rows", async ({ page }) => {
  const phone = `139${(Date.now() % 1e8).toString().padStart(8, "0")}`;
  try {
    await page.goto("/login");
    await page.waitForLoadState("networkidle");
    await page.waitForTimeout(1_000);
    await page.getByPlaceholder("手机号").fill(phone);
    await agreeAll(page);
    await page.getByRole("button", { name: "发送验证码" }).click();
    const code = execSync(`${REDIS} GET auth:web:sms-code:${phone}`).toString().trim();
    expect(
      code,
      "no redis code — the stack must run the sms mock backend (APP_SMS_BACKEND=mock)",
    ).toMatch(/^\d{6}$/);
    await page.getByPlaceholder("6 位验证码").fill(code);
    await page.getByRole("button", { name: "登录 / 注册" }).click();
    await page.waitForURL("**/profile/setup/**", { timeout: 15_000 });
    await walkWizard(page);

    const seat = sql(
      "system",
      `SELECT u.account, m.population FROM credential c JOIN "user" u ON u.id=c.user_id JOIN membership m ON m.user_id=u.id` +
        ` WHERE c.kind='sms' AND c.identifier='${phone}'`,
    );
    expect(memberChainOk(seat), `sms member chain: [${seat}]`).toBeTruthy();
    // The esports side holds no user table — the cross-service id is a
    // plain string key, so the wizard's profile row is looked up in two
    // steps, never a cross-database join.
    const memberId = sql("system", `SELECT id FROM "user" WHERE account='${chainAccount(seat)}'`);
    const profile = sql("esports", `SELECT nickname FROM profile WHERE user_id='${memberId}'`);
    expect(profile).toBeTruthy();
  } finally {
    eraseMember({ byPhone: phone });
    execSync(
      `${REDIS} DEL auth:web:sms-code:${phone} auth:sms:send-daily:${phone} rate:auth:sms:send-window:${phone}`,
    );
  }
});

// --- the wechat leg --------------------------------------------------------------------------

test.describe("wechat leg", () => {
  // The page gates the button on the WeChat browser's UA signature.
  test.use({ userAgent: "Mozilla/5.0 (Linux; Android 14) MicroMessenger/8.0.49" });

  test("wechat login rides the endpoint override: mock hop → callback → member rows", async ({
    page,
  }) => {
    test.skip(
      !(await wechatSeamWired()),
      `stack must boot with APP_WECHAT_ENDPOINT=http://127.0.0.1:${WX_PORT}`,
    );
    try {
      await page.goto("/login");
      await page.waitForLoadState("networkidle");
      await page.waitForTimeout(1_000);
      await agreeAll(page);
      await page.getByRole("button", { name: "微信授权登录" }).click();
      // The hop leaves for the stand-in, which bounces straight back to the
      // client callback; the login lands on the born member's wizard.
      await page.waitForURL("**/profile/setup/**", { timeout: 15_000 });

      const seat = sql(
        "system",
        `SELECT u.account, m.population FROM credential c JOIN "user" u ON u.id=c.user_id JOIN membership m ON m.user_id=u.id` +
          ` WHERE c.kind='wechat_mp' AND c.identifier='${WX_OPENID}'`,
      );
      expect(memberChainOk(seat), `wechat member chain: [${seat}]`).toBeTruthy();
    } finally {
      eraseMember({ byOpenid: WX_OPENID });
    }
  });
});

// --- the conflict leg ------------------------------------------------------------------------

test("bind-phone on an occupied number stops at the conflict card, no code sent", async ({
  page,
}) => {
  const stamp = Date.now().toString().slice(-8);
  const heldPhone = `138${stamp}`; // 3 + 8 = 11 digits
  const bUser = `e2ecfbu${stamp}`;
  // The occupying side: a member holding an sms credential for the number.
  sql(
    "system",
    `INSERT INTO "user" (id, created_at, updated_at, account, password_hash, enabled) VALUES ('${bUser}', now(), now(), '${bUser}', '', true);`,
  );
  sql(
    "system",
    `INSERT INTO credential (id, created_at, updated_at, kind, identifier, user_id) VALUES ('${bUser}c', now(), now(), 'sms', '${heldPhone}', '${bUser}');`,
  );
  sql(
    "system",
    `INSERT INTO membership (id, created_at, updated_at, user_id, tenant_id, population) VALUES ('${bUser}m', now(), now(), '${bUser}', 'default', 'member');`,
  );
  // The caller: an admin-provisioned password member driving the real UI.
  const adminToken = (
    await api("/api/v1/system/auth/login", {
      account: "admin",
      password: "admin123456",
      population: "user",
    })
  ).json.sessionToken as string;
  const account = `e2ecfa${stamp}`;
  const created = await api(
    "/api/v1/system/users",
    { account, password: "e2ePass123", population: "member" },
    `session_token=${adminToken}`,
  );
  expect(created.status, `member provision: ${JSON.stringify(created.json)}`).toBe(200);
  try {
    await page.goto("/login");
    await page.waitForLoadState("networkidle");
    await page.waitForTimeout(1_000);
    await page.getByRole("button", { name: "密码登录" }).click();
    await page.getByPlaceholder("手机号").fill(account);
    await page.getByPlaceholder("密码（≥6 位）").fill("e2ePass123");
    await agreeAll(page);
    await page.getByRole("button", { name: "登录 / 注册" }).click();
    await page.waitForURL("**/profile/setup/**", { timeout: 15_000 });
    await walkWizard(page);

    await page.goto("/profile/security/bind-phone?from=me");
    // The hydration gap: a click before the client bundle settles dies
    // silently (no POST, no toast) — fullchain's finding, same law here.
    await page.waitForLoadState("networkidle");
    await page.waitForTimeout(1_000);
    const posts: string[] = [];
    page.on("response", (res) => {
      if (res.request().method() === "POST") posts.push(`${res.status()} ${res.url().slice(-64)}`);
    });
    await page.getByPlaceholder("手机号（11 位）").fill(heldPhone);
    await page.getByRole("button", { name: "发送验证码" }).click();
    // The preview's verdict, not a sent code: the occupied number answers
    // the conflict card with the merge CTA — occupancy is not a secret the
    // bind flow doesn't already reveal.
    try {
      await expect(page.getByText("该手机号已绑定其他账号")).toBeVisible({ timeout: 10_000 });
    } catch {
      const toasts = await page
        .locator("[data-sonner-toast]")
        .allTextContents()
        .catch(() => []);
      const sentCode = execSync(`${REDIS} EXISTS auth:web:sms-code:${heldPhone}`).toString().trim();
      throw new Error(
        `no conflict card; POSTs: [${posts.join(" | ")}]; toasts: [${toasts.join(" | ")}]; code-sent=${sentCode}`,
      );
    }
    await expect(page.getByText("本次绑定已拦截、未做任何改动")).toBeVisible();
    // This scenario's pair is empty on both sides (neither carries a test
    // record, membership, or unlock), and the corpus bar reads both-empty
    // as BLOCKED — so the card's CTA is the guidance one, not the merge
    // direction a weighted caller would see.
    await expect(page.getByRole("button", { name: /为什么不能自动合并/ })).toBeVisible();
    const leaked = execSync(`${REDIS} EXISTS auth:web:sms-code:${heldPhone}`).toString().trim();
    expect(leaked, "no code may be sent for a number the preview already holds").toBe("0");
  } finally {
    eraseMember({ byAccount: account });
    eraseMember({ byAccount: bUser });
  }
});

// The sms/wechat chains return "xidAccount|member" — a born member's xid
// account is non-empty and the seat reads member.
function memberChainOk(joined: string) {
  return /^[a-zA-Z0-9]+\|member$/.test(joined);
}
function chainAccount(joined: string) {
  return joined.split("|")[0] ?? "";
}
