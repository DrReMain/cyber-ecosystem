import { execSync } from "node:child_process";
import { expect, type Page, test } from "@playwright/test";

// The M1 full-chain acceptance, D-O semantic grade: login → wizard → match →
// four games → session done → single report → activation-code unlock, one
// member, one browser, no shortcuts. Runs only against the live dev stack
// (CTR_E2E=1): postgres + system + esports up, seeds in place.
//
// The reaction-slot draw is day-salted, and red_dot is the only game whose
// calibration curve is frozen — every other scorer returns curve-pending
// (raw only, authoritative null), and the report anchor needs an
// authoritative record. The draw mechanic itself is E4-d's e2e territory;
// here the pool is narrowed for the run's window so the reaction slot lands
// red_dot deterministically, and the finally block restores the catalog,
// drops the cohort fixtures, and erases the run's member end to end.
test.skip(!process.env.CTR_E2E, "full-chain needs the live stack (CTR_E2E=1)");
test.use({ locale: "zh-CN" });
test.setTimeout(480_000);

const SYSTEM_HTTP = "http://localhost:11001";
const ESPORTS_HTTP = "http://localhost:11003";
const PG_EXEC = "docker exec cyber-ecosystem-postgresql-1 psql -U postgres -d esports";
const REACTION_RIVALS = ["dual_zone", "number_stars", "target_recognition"];

async function api(baseUrl: string, path: string, body: unknown, cookie?: string) {
  const res = await fetch(`${baseUrl}${path}`, {
    method: "POST",
    headers: {
      "content-type": "application/json",
      ...(cookie ? { cookie } : {}),
    },
    body: JSON.stringify(body),
  });
  return {
    status: res.status,
    json: (await res.json().catch(() => ({}))) as Record<string, unknown>,
  };
}

async function loginToken(account: string, password: string, population: string) {
  const { json } = await api(SYSTEM_HTTP, "/api/v1/system/auth/login", {
    account,
    password,
    population,
  });
  return String(json.sessionToken ?? "");
}

test("login → wizard → match → 4 games → done → report → code unlock", async ({ page }) => {
  // --- provision: fresh member + narrowed reaction pool -------------------
  const adminToken = await loginToken("admin", "admin123456", "user");
  expect(adminToken.length).toBeGreaterThan(0);
  const account = `e2e${Date.now().toString(36)}`;
  const password = "e2ePass123";
  const created = await api(
    SYSTEM_HTTP,
    "/api/v1/system/users",
    { account, password, population: "member" },
    `session_token=${adminToken}`,
  );
  expect(created.status, `member provision: ${JSON.stringify(created.json)}`).toBe(200);

  const restore = REACTION_RIVALS.map((code) => `'${code}'`).join(",");
  execSync(`${PG_EXEC} -c "UPDATE test SET category='moba' WHERE code IN (${restore});"`);

  try {
    // --- login face ---------------------------------------------------------
    const consoleErrors: string[] = [];
    page.on("console", (msg) => {
      if (msg.type() === "error") consoleErrors.push(msg.text().slice(0, 300));
    });
    page.on("pageerror", (err) => consoleErrors.push(`pageerror: ${err.message.slice(0, 300)}`));
    await page.goto("/login");
    // "load" fires before the client bundle hydrates; a submit clicked in
    // that gap dies silently (no POST, no toast — the native checkbox even
    // toggles without React knowing). Let the module graph settle first.
    await page.waitForLoadState("networkidle");
    await page.waitForTimeout(1_000);
    // The code tab is the default since the E6c-a login face; this chain
    // rides the password pair, so flip to it before filling.
    await page.getByRole("button", { name: "密码登录" }).click();
    await page.getByPlaceholder("手机号").fill(account);
    await page.getByPlaceholder("密码（≥6 位）").fill(password);
    // The RAC pair is visually hijacked: the real input hides behind the
    // AgreeBox skin, so pointer clicks land on the box and programmatic
    // dispatch is swallowed — focus plus a real keyboard Space flips it.
    const selectAll = page.getByRole("checkbox", { name: "全选" });
    await selectAll.focus();
    await page.keyboard.press(" ");
    await expect(selectAll).toBeChecked();
    const posts: string[] = [];
    page.on("response", (res) => {
      if (res.request().method() === "POST") posts.push(`${res.status()} ${res.url().slice(-72)}`);
    });
    await page.getByRole("button", { name: "登录 / 注册" }).click();
    try {
      await page.waitForURL("**/profile/setup/**", { timeout: 15_000 });
    } catch {
      // Diagnostics only on the failure path — a successful login removes
      // the button, so reading it eagerly would hang the whole test.
      const toasts = await page
        .locator("[data-sonner-toast]")
        .allTextContents()
        .catch(() => []);
      const submitText = await page
        .getByRole("button", { name: "登录 / 注册" })
        .textContent({ timeout: 2_000 })
        .catch(() => "(button gone)");
      throw new Error(
        `login stalled; submitText: [${submitText}]; POSTs: [${posts.join(" | ")}]; toasts: [${toasts.join(" | ")}]; console: [${consoleErrors.join(" || ")}]`,
      );
    }

    // --- wizard: nickname → birth → games ----------------------------------
    await page.getByPlaceholder(/请输入昵称/).fill(`链路${Date.now() % 1000}`);
    await page.getByRole("button", { name: "下一步" }).click();

    // Native radios ride sr-only inputs inside their labels — click the
    // visible label text; label semantics toggle the hidden input.
    await page.getByText("男", { exact: true }).click();
    const selects = page.getByRole("combobox");
    await selects.nth(0).selectOption({ index: 8 });
    await selects.nth(1).selectOption({ index: 5 });
    await page.getByRole("button", { name: "下一步" }).click();

    await page.getByText("CS2", { exact: true }).click();
    await page.getByText("无畏契约", { exact: true }).click();
    await page.getByRole("button", { name: "完成" }).click();
    await page.waitForURL("**/match", { timeout: 15_000 });

    // --- four games ---------------------------------------------------------
    for (let step = 0; step < 4; step++) {
      await page.goto("/match");
      await page.getByRole("button", { name: "开始测试" }).click();
      await page.waitForURL("**/test/**", { timeout: 15_000 });
      // The linear start button walks the reaction slot first; the narrowed
      // pool must hand it to red_dot.
      if (step === 0) {
        await expect(page).toHaveURL(/\/test\/red_dot/);
      }
      await page.getByRole("button", { name: /继续 · 开始测试/ }).click();

      if (page.url().includes("red_dot")) {
        await page.waitForTimeout(4_000); // countdown
        await shootTargets(page);
      } else {
        // Zero-interaction honest runs: valid empty streams grade curve-pending.
        await page.waitForTimeout(3_000);
      }
      await expect(page.getByText("重新测试")).toBeVisible({ timeout: 60_000 });
      await page.waitForTimeout(800);
    }

    // --- cohort fixture: the A-② free seat is percentile-gated, and the
    // percentile needs same-band same-sex siblings past the cohort floor
    // (the f3 walkthrough crossed it with nine SQL samples). The siblings
    // live entirely in the esports db — cross-service ids are plain strings.
    const memberId = execSync(
      `docker exec cyber-ecosystem-postgresql-1 psql -U postgres -d system -t -A -c "SELECT id FROM \\"user\\" WHERE account='${account}';"`,
    )
      .toString()
      .trim();
    const profileRow = execSync(
      `${PG_EXEC} -t -A -F'|' -c "SELECT birth_year, birth_month, sex, age_band FROM profile WHERE user_id='${memberId}';"`,
    )
      .toString()
      .trim();
    const [birthYear, birthMonth, sex, ageBand] = profileRow.split("|");
    expect(sex, `member profile row: [${profileRow}]`).toBeTruthy();
    execSync(
      `${PG_EXEC} -c "INSERT INTO profile (id, created_at, updated_at, tenant_id, user_id, nickname, birth_year, birth_month, sex, age_band)` +
        ` SELECT 'e2efixp'||g, now(), now(), 'default', 'e2efixuser'||g, '样本'||g, ${birthYear}, ${birthMonth}, '${sex}', '${ageBand}' FROM generate_series(1,9) g;"`,
    );
    execSync(
      `${PG_EXEC} -c "INSERT INTO test_record (id, created_at, updated_at, tenant_id, test_code, user_id, attempt_no, started_at, ended_at, raw_detail, raw_score, authoritative_score, grade, valid)` +
        ` SELECT 'e2efixr'||g, now() - (g || ' hours')::interval, now(), 'default', 'red_dot', 'e2efixuser'||g, 1, now(), now(), '{}'::jsonb, 50+g, 50+g, 'B', true FROM generate_series(1,9) g;"`,
    );

    // --- session done page --------------------------------------------------
    await page.goto("/match");
    await page.getByRole("button", { name: "查看结果" }).click();
    await page.waitForURL("**/report/session/**", { timeout: 15_000 });
    await expect(page.getByText(/\/\s*4/).first()).toBeVisible();

    // --- single report ------------------------------------------------------
    await page.getByRole("button", { name: "查看能力报告" }).click();
    await page.waitForURL(/\/report\/(?!session)/, { timeout: 15_000 });
    await expect(page.getByText("能力测评报告")).toBeVisible();
    // E5-e honest faces: the free seat rides the one authoritative row, the
    // curve-pending majority renders absent (missing dims never fake a
    // score), and the paywall stays up.
    await expect(page.getByText("最高分 · 已解锁")).toBeVisible();
    await expect(page.getByText(/已测 1 \/ 4 维/)).toBeVisible();
    await expect(page.getByText("未测的能力不虚构分数")).toBeVisible();
    await expect(page.getByText(/¥\d/).first()).toBeVisible();

    // --- activation-code unlock (member session, connect transport) ---------
    const minted = await api(
      ESPORTS_HTTP,
      "/api/v1/esports/activation-codes/generate",
      { reportType: "REPORT_TYPE_CAREER", count: 1, expireDays: 7 },
      `session_token=${adminToken}`,
    );
    expect(minted.status, `code mint: ${JSON.stringify(minted.json)}`).toBe(200);
    const code = String(((minted.json.codes as string[]) ?? [])[0] ?? "");
    expect(code.length).toBe(18); // AC + 16

    const redeem = await rpcInSession(
      page,
      "cyber.esports.v1.EntitlementService/RedeemActivationCode",
      { code },
    );
    expect(redeem.status, `redeem: ${JSON.stringify(redeem.body)}`).toBe(200);
    const access = await rpcInSession(page, "cyber.esports.v1.EntitlementService/GetReportAccess", {
      reportType: "REPORT_TYPE_CAREER",
    });
    expect(access.status).toBe(200);
    // The code outranks the policy — the exact contract promise, wire-named
    // as the enum value (transport speaks value names, not short names).
    expect(access.body).toMatchObject({
      allowed: true,
      reason: "REPORT_ACCESS_REASON_CODE_REDEEMED",
    });
  } finally {
    execSync(`${PG_EXEC} -c "UPDATE test SET category='universal' WHERE code IN (${restore});"`);
    execSync(`${PG_EXEC} -c "DELETE FROM test_record WHERE id LIKE 'e2efixr%';"`);
    execSync(`${PG_EXEC} -c "DELETE FROM profile WHERE id LIKE 'e2efixp%';"`);
    const runIds = execSync(
      `docker exec cyber-ecosystem-postgresql-1 psql -U postgres -d system -t -A -c "SELECT id FROM \\"user\\" WHERE account='${account}';"`,
    )
      .toString()
      .trim();
    if (runIds) {
      const quoted = runIds
        .split("\n")
        .map((id) => `'${id.trim()}'`)
        .join(",");
      execSync(`${PG_EXEC} -c "DELETE FROM test_record WHERE user_id IN (${quoted});"`);
      execSync(`${PG_EXEC} -c "DELETE FROM match_session WHERE user_id IN (${quoted});"`);
      execSync(`${PG_EXEC} -c "DELETE FROM entitlement WHERE user_id IN (${quoted});"`);
      execSync(`${PG_EXEC} -c "DELETE FROM profile WHERE user_id IN (${quoted});"`);
      execSync(
        `docker exec cyber-ecosystem-postgresql-1 psql -U postgres -d system -c "DELETE FROM membership WHERE user_id IN (${quoted});"`,
      );
      execSync(
        `docker exec cyber-ecosystem-postgresql-1 psql -U postgres -d system -c "DELETE FROM \\"user\\" WHERE id IN (${quoted});"`,
      );
    }
  }
});

async function rpcInSession(page: Page, procedure: string, body: unknown) {
  return page.evaluate(
    async ({ proc, payload }) => {
      const res = await fetch(`/connect/${proc}`, {
        method: "POST",
        headers: {
          "content-type": "application/json",
          "connect-protocol-version": "1",
          "x-session-namespace": "member",
        },
        credentials: "include",
        body: JSON.stringify(payload),
      });
      return { status: res.status, body: await res.json().catch(() => ({})) };
    },
    { proc: procedure, payload: body },
  );
}

// red_dot hits by stage coordinates: the engine listens on the stage and
// computes hits from pointer coordinates, so each poll dispatches a
// pointerdown at the live target's center. The ~300ms jittered cadence keeps
// reaction times on the human side of the absurd-reaction gate.
async function shootTargets(page: Page) {
  await page.evaluate(
    () =>
      new Promise<void>((resolve) => {
        const deadline = performance.now() + 31_000;
        const tick = () => {
          const target = document.querySelector<HTMLElement>(".absolute.rounded-full.bg-danger");
          if (target?.parentElement) {
            const rect = target.getBoundingClientRect();
            target.parentElement.dispatchEvent(
              new PointerEvent("pointerdown", {
                clientX: rect.x + rect.width / 2,
                clientY: rect.y + rect.height / 2,
                bubbles: true,
                pointerId: 1,
                isPrimary: true,
              }),
            );
          }
          if (performance.now() < deadline) {
            setTimeout(tick, 260 + Math.random() * 160);
          } else {
            resolve();
          }
        };
        tick();
      }),
  );
}
