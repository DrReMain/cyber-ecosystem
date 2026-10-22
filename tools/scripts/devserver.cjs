#!/usr/bin/env node
// Dev-server lifecycle for coding agents: `up` spawns a project's dev target
// detached (own process group) and pins its identity in .run/<name>/state.json;
// `down` kills exactly that recorded group — never by port — and removes the
// whole session dir, so dev.log exists only while its server is running.

const { spawn, execFileSync } = require("node:child_process");
const fs = require("node:fs");
const path = require("node:path");

const WORKSPACE_ROOT = path.resolve(__dirname, "..", "..");
const RUN_DIR = path.join(WORKSPACE_ROOT, ".run");
const READY_TIMEOUT_MS = 120000; // generous for cold Go builds; a dead process short-circuits the wait
const TERM_GRACE_MS = 5000;

const die = (message) => {
  console.error(message);
  process.exit(1);
};

const sleep = (ms) => new Promise((resolve) => setTimeout(resolve, ms));

const rel = (p) => path.relative(WORKSPACE_ROOT, p);

function parseArgs(argv) {
  const parsed = { _: [] };
  for (let i = 0; i < argv.length; i++) {
    if (argv[i].startsWith("--")) {
      parsed[argv[i].slice(2)] = argv[++i];
    } else {
      parsed._.push(argv[i]);
    }
  }
  return parsed;
}

// ps field for a pid, "" when the pid is gone (identity checks + reporting)
function psField(field, pid) {
  try {
    return execFileSync("ps", ["-o", `${field}=`, "-p", String(pid)], {
      encoding: "utf8",
    }).trim();
  } catch {
    return "";
  }
}

// lstart pins a recorded pid against silent PID reuse
const identityOf = (pid) => psField("lstart", pid);

// a negative pid signals the whole process group spawned by `up`
const groupAlive = (pgid) => {
  try {
    process.kill(-pgid, 0);
    return true;
  } catch (err) {
    return err.code === "EPERM";
  }
};

// report-only: port state is a signal, never a kill target
function portListeners(port) {
  try {
    return execFileSync("lsof", ["-nP", `-tiTCP:${port}`, "-sTCP:LISTEN"], {
      encoding: "utf8",
    })
      .split("\n")
      .filter(Boolean);
  } catch {
    return [];
  }
}

function readState(name) {
  const file = path.join(RUN_DIR, name, "state.json");
  return fs.existsSync(file) ? JSON.parse(fs.readFileSync(file, "utf8")) : null;
}

const sessionAlive = (state) => groupAlive(state.pid) && identityOf(state.pid) === state.lstart;

const tailText = (file, lines = 30) => {
  if (!fs.existsSync(file)) return "";
  return fs.readFileSync(file, "utf8").trimEnd().split("\n").slice(-lines).join("\n");
};

// ports stay in each project's own config (vite.config.ts / configs/config.yaml)
// as the single source; a bind failure names its conflicting port itself, so
// the occupant is reported from the native error instead of a declared port
function reportBindConflict(logText) {
  const line = logText.split("\n").find((l) => /already in use/i.test(l));
  if (!line) return;
  const port = (line.match(/Port (\d+)/) || line.match(/listen tcp \S+:(\d{2,5})/) || [])[1];
  if (!port) return;
  for (const pid of portListeners(port)) {
    console.error(`port ${port} is held by pid ${pid}: ${psField("command", pid)}`);
  }
}

// TERM to the group, grace period, then KILL — the only kill surface
async function killGroup(pgid) {
  try {
    process.kill(-pgid, "SIGTERM");
  } catch {}
  const deadline = Date.now() + TERM_GRACE_MS;
  while (Date.now() < deadline && groupAlive(pgid)) {
    await sleep(200);
  }
  if (groupAlive(pgid)) {
    try {
      process.kill(-pgid, "SIGKILL");
    } catch {}
  }
  await sleep(300);
  return !groupAlive(pgid);
}

async function up(args) {
  const { project } = args;
  if (!project) {
    die("usage: devserver.cjs up --project <nx-project> [--ready <regex>]");
  }
  const name = project.split("/").pop();
  const dir = path.join(RUN_DIR, name);
  const logFile = path.join(dir, "dev.log");

  const existing = readState(name);
  if (existing && sessionAlive(existing)) {
    console.log(`${name}: already up (pid ${existing.pid}, since ${existing.startedAt})`);
    console.log(`log: ${rel(logFile)} — read or tail -f while running; removed on stop`);
    return;
  }

  fs.rmSync(dir, { recursive: true, force: true });
  fs.mkdirSync(dir, { recursive: true });
  const logFd = fs.openSync(logFile, "w");
  const child = spawn(path.join(WORKSPACE_ROOT, "nx"), ["run", `${project}:dev`], {
    cwd: WORKSPACE_ROOT,
    detached: true,
    stdio: ["ignore", logFd, logFd],
    env: { ...process.env, NO_COLOR: "1" }, // keep dev.log free of ANSI escapes for agents
  });
  child.on("error", () => {}); // surfaced via the missing-lstart branch below
  child.unref(); // the script exits on ready; the detached server keeps running
  fs.closeSync(logFd);

  process.on("SIGINT", async () => {
    await killGroup(child.pid);
    fs.rmSync(dir, { recursive: true, force: true });
    process.exit(130);
  });

  const lstart = identityOf(child.pid);
  if (!lstart) {
    const tail = tailText(logFile);
    fs.rmSync(dir, { recursive: true, force: true });
    die(`${name}: dev process exited immediately\n${tail}`);
  }
  fs.writeFileSync(
    path.join(dir, "state.json"),
    `${JSON.stringify(
      { pid: child.pid, lstart, project, startedAt: new Date().toISOString() },
      null,
      2,
    )}\n`,
  );

  // vite prints "Local:", the kratos connect plane prints "server listening";
  // the trailing .* keeps the whole matched line in the up report; --ready overrides
  const ready = new RegExp(args.ready ?? "Local:.*|listening.*");
  const deadline = Date.now() + READY_TIMEOUT_MS;
  while (Date.now() < deadline) {
    const output = fs.existsSync(logFile) ? fs.readFileSync(logFile, "utf8") : "";
    const match = output.match(ready);
    if (match) {
      console.log(`${name}: dev server up`);
      console.log(`  ${match[0].trim()}`);
      console.log(`  pid: ${child.pid} (own process group)`);
      console.log(`  log: ${rel(logFile)} — read or tail -f while running; removed on stop`);
      console.log(`  stop: ./nx run ${project}:dev:down`);
      return;
    }
    if (!groupAlive(child.pid)) {
      const tail = tailText(logFile);
      await killGroup(child.pid);
      fs.rmSync(dir, { recursive: true, force: true });
      reportBindConflict(tail);
      die(`${name}: dev process died before ready\n${tail}`);
    }
    await sleep(250);
  }
  const tail = tailText(logFile);
  await killGroup(child.pid);
  fs.rmSync(dir, { recursive: true, force: true });
  die(`${name}: not ready after ${READY_TIMEOUT_MS / 1000}s — killed, session cleaned\n${tail}`);
}

async function down(args) {
  const { project } = args;
  if (!project) {
    die("usage: devserver.cjs down --project <nx-project>");
  }
  const name = project.split("/").pop();
  const dir = path.join(RUN_DIR, name);
  const state = readState(name);
  if (!state) {
    console.log(`${name}: no recorded session (nothing to stop)`);
    return;
  }
  if (!sessionAlive(state)) {
    fs.rmSync(dir, { recursive: true, force: true });
    console.log(`${name}: session was already dead — state and log discarded`);
    return;
  }
  const gone = await killGroup(state.pid);
  if (!gone) {
    console.error(`${name}: group ${state.pid} survived SIGKILL — dir kept for inspection`);
    process.exit(1);
  }
  fs.rmSync(dir, { recursive: true, force: true });
  console.log(`${name}: down (group ${state.pid} terminated, session dir removed)`);
}

(async () => {
  if (process.platform === "win32") {
    die("devserver lifecycle is POSIX-only (process groups) — Windows is a known gap");
  }
  const [command, ...rest] = process.argv.slice(2);
  const args = parseArgs(rest);
  if (command === "up") {
    await up(args);
  } else if (command === "down") {
    await down(args);
  } else {
    die("usage: devserver.cjs <up|down> --project <nx-project> [--ready <regex>]");
  }
})();
