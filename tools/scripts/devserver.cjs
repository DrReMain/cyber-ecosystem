#!/usr/bin/env node
// Dev-server lifecycle for coding agents: `up` spawns a project's dev target
// detached (own process group) and pins its identity in .run/<name>/state.json;
// `down` kills exactly that recorded group — never by port — and removes the
// whole session dir, so dev.log exists only while its server is running.

const { spawn, spawnSync, execFileSync } = require("node:child_process");
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

// The session dir name is a project name's last segment, and it must be one
// clean path segment: anything else ("..", ".", foreign bytes) would aim
// rmSync outside the run dir — "foo/.." resolves to the workspace root
// itself.
function sessionName(project) {
  const name = project.split("/").pop();
  if (!/^[A-Za-z0-9][A-Za-z0-9._-]*$/.test(name)) {
    die(`invalid project name: ${project}`);
  }
  return name;
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

// lstart pins a recorded pid against silent PID reuse. Windows has no
// lstart face; the constant identity degenerates the pin to liveness-only
// there (PID reuse on a dev box within a session's life is not a real
// threat model).
const identityOf = (pid) => (process.platform === "win32" ? "windows" : psField("lstart", pid));

// pidAlive is portable — signal 0 probes existence on every platform
const pidAlive = (pid) => {
  try {
    process.kill(pid, 0);
    return true;
  } catch (err) {
    return err.code === "EPERM";
  }
};

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
  if (process.platform === "win32") {
    // netstat is the Windows face; a LISTENING row's local address carries
    // the port and the trailing column is the owning pid
    try {
      const out = execFileSync("netstat", ["-ano", "-p", "TCP"], { encoding: "utf8" });
      const pids = [];
      for (const line of out.split("\n")) {
        if (/LISTENING/i.test(line) && new RegExp(`[:.]${port}\\s`).test(line)) {
          const pid = Number(line.trim().split(/\s+/).pop());
          if (pid) pids.push(String(pid));
        }
      }
      return pids;
    } catch {
      return [];
    }
  }
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

// report-only: a recorded port still listening after the kill means a
// process escaped both the tree walk and the group signals — name it,
// never kill by port (the port's owner is a fact, not a target)
function reportPortStraggler(state) {
  const ports = state.ports ?? (state.port ? [state.port] : []);
  let held = false;
  for (const port of ports) {
    const listeners = portListeners(port);
    for (const pid of listeners) {
      console.error(`port ${port} is still held by pid ${pid}: ${psField("command", pid)}`);
      held = true;
    }
  }
  return held;
}

// NO_COLOR is a convention, not a law — local CLI beautifiers re-enable
// color through force-flags set in shell rc files, which override it. The
// child gets the flags deleted outright, plus TERM=dumb for spinners.
function colorlessEnv() {
  const env = { ...process.env, NO_COLOR: "1", TERM: "dumb", CLICOLOR: "0" };
  delete env.FORCE_COLOR;
  delete env.CLICOLOR_FORCE;
  return env;
}

// POSIX spawns the binary itself. Windows routes through cmd.exe: wrapper
// shims (pnpm → pnpm.cmd) are not executable with shell:false (Node refuses
// .bat/.cmd), and cmd resolves .exe/.cmd uniformly — the same resolution the
// repo's own nx.bat relies on; the argv here is the repo's fixed config,
// never external input.
function spawnDev(file, args, options) {
  if (process.platform === "win32") {
    return spawn("cmd.exe", ["/c", file, ...args], options);
  }
  return spawn(file, args, options);
}

// The dev targets live in each project's project.json; this finds a
// project's root by name across the repo's three lanes — a bounded layout
// scan (the documented directory law), not an nx-graph rebuild.
const PROJECT_LANES = ["app/services", "app/clients", "shared-ts"];
function resolveProjectRoot(project) {
  for (const lane of PROJECT_LANES) {
    const laneDir = path.join(WORKSPACE_ROOT, lane);
    if (!fs.existsSync(laneDir)) continue;
    for (const entry of fs.readdirSync(laneDir, { withFileTypes: true })) {
      if (!entry.isDirectory()) continue;
      const file = path.join(laneDir, entry.name, "project.json");
      if (!fs.existsSync(file)) continue;
      if (JSON.parse(fs.readFileSync(file, "utf8")).name === project) {
        return path.join(laneDir, entry.name);
      }
    }
  }
  return die(`cannot locate project ${project} under ${PROJECT_LANES.join(", ")}`);
}

// groupAlive is POSIX-only (negative-pid signals); on Windows liveness is
// the plain pid probe
const sessionAlive = (state) =>
  process.platform === "win32"
    ? pidAlive(state.pid) && identityOf(state.pid) === state.lstart
    : groupAlive(state.pid) && identityOf(state.pid) === state.lstart;

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

// direct children of a pid, empty when the pid is gone
function childrenOf(pid) {
  try {
    return execFileSync("pgrep", ["-P", String(pid)], { encoding: "utf8" })
      .split("\n")
      .filter(Boolean)
      .map(Number);
  } catch {
    return [];
  }
}

// The task tree spans process groups: `go tool kratos run` and `pnpm exec`
// both sit their subtree in a group of their own, so the spawn leader's
// group alone never reaches the server. The kill walks descendants first
// and signals every group the tree touches — pid-pinned, never by port.
// Windows has no signal-able groups: taskkill /T walks the tree itself,
// polite pass first, /F after the same grace window.
async function killTree(rootPid) {
  if (process.platform === "win32") {
    spawn("taskkill", ["/PID", String(rootPid), "/T"], { stdio: "ignore" });
    const soft = Date.now() + TERM_GRACE_MS;
    while (Date.now() < soft && pidAlive(rootPid)) {
      await sleep(200);
    }
    if (pidAlive(rootPid)) {
      spawnSync("taskkill", ["/PID", String(rootPid), "/T", "/F"], { stdio: "ignore" });
    }
    await sleep(300);
    return !pidAlive(rootPid);
  }
  const pids = [rootPid];
  const walk = (p) => {
    for (const c of childrenOf(p)) {
      pids.push(c);
      walk(c);
    }
  };
  walk(rootPid);
  const groups = [...new Set(pids.map((p) => Number(psField("pgid", p)) || p))];
  for (const g of groups) {
    try {
      process.kill(-g, "SIGTERM");
    } catch {}
  }
  const deadline = Date.now() + TERM_GRACE_MS;
  while (Date.now() < deadline && pids.some((p) => pidAlive(p))) {
    await sleep(200);
  }
  for (const p of pids) {
    if (!pidAlive(p)) {
      continue;
    }
    try {
      process.kill(p, "SIGKILL");
    } catch {}
  }
  await sleep(300);
  return pids.every((p) => !pidAlive(p));
}

async function up(args) {
  const { project } = args;
  if (!project) {
    die("usage: devserver.cjs up --project <nx-project> [--ready <regex>]");
  }
  const name = sessionName(project);
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
  // spawn the project's own dev command straight from its project.json —
  // the dev target stays the single source, and the server's ancestry loses
  // the nested nx/shell hops entirely
  const projectRoot = resolveProjectRoot(project);
  const projectJson = JSON.parse(fs.readFileSync(path.join(projectRoot, "project.json"), "utf8"));
  const devCommand = projectJson.targets?.dev?.options?.command;
  if (typeof devCommand !== "string" || !devCommand.trim()) {
    die(`${project}: project.json declares no dev command`);
  }
  const [devBin, ...devArgs] = devCommand.trim().split(/\s+/);
  // the target's env block rides along (dev-plane values like a dev-only
  // master key), so the dev target is the single source for nx runs and
  // direct devserver spawns alike
  const targetEnv = projectJson.targets?.dev?.options?.env || {};
  const child = spawnDev(devBin, devArgs, {
    cwd: projectRoot,
    detached: true,
    stdio: ["ignore", logFd, logFd],
    env: { ...colorlessEnv(), ...targetEnv }, // keep dev.log free of ANSI escapes for agents
  });
  child.on("error", () => {}); // surfaced via the missing-lstart branch below
  child.unref(); // the script exits on ready; the detached server keeps running
  fs.closeSync(logFd);

  process.on("SIGINT", async () => {
    await killTree(child.pid);
    fs.rmSync(dir, { recursive: true, force: true });
    process.exit(130);
  });

  const lstart = identityOf(child.pid);
  if (!lstart) {
    const tail = tailText(logFile);
    fs.rmSync(dir, { recursive: true, force: true });
    die(`${name}: dev process exited immediately\n${tail}`);
  }
  const state = { pid: child.pid, lstart, project, startedAt: new Date().toISOString() };
  fs.writeFileSync(path.join(dir, "state.json"), `${JSON.stringify(state, null, 2)}\n`);

  // vite prints "Local:", the kratos connect plane prints "server listening";
  // the trailing .* keeps the whole matched line in the up report; --ready overrides
  const ready = new RegExp(args.ready ?? "Local:.*|listening.*");
  const deadline = Date.now() + READY_TIMEOUT_MS;
  while (Date.now() < deadline) {
    const output = fs.existsSync(logFile) ? fs.readFileSync(logFile, "utf8") : "";
    const match = output.match(ready);
    if (match) {
      // every port the log names by ready time (vite its URL, kratos an
      // addr= per listener) — recorded for the down-side straggler check
      const ports = [
        ...output.matchAll(/localhost:(\d{2,5})/g),
        ...output.matchAll(/listen (?:tcp )?[\d.]+:(\d{2,5})/g),
        ...output.matchAll(/addr=(?:[\d.]+|\[[^\]]*\]):(\d{2,5})/g),
      ].map((m) => Number(m[1]));
      state.ports = [...new Set(ports)];
      if (state.ports.length > 0) {
        fs.writeFileSync(path.join(dir, "state.json"), `${JSON.stringify(state, null, 2)}\n`);
      }
      console.log(`${name}: dev server up`);
      console.log(`  ${match[0].trim()}`);
      console.log(`  pid: ${child.pid} (own process group)`);
      console.log(`  log: ${rel(logFile)} — read or tail -f while running; removed on stop`);
      console.log(`  stop: ./nx run ${project}:dev-down`);
      return;
    }
    if (!pidAlive(child.pid)) {
      const tail = tailText(logFile);
      await killTree(child.pid);
      fs.rmSync(dir, { recursive: true, force: true });
      reportBindConflict(tail);
      die(`${name}: dev process died before ready\n${tail}`);
    }
    await sleep(250);
  }
  const tail = tailText(logFile);
  await killTree(child.pid);
  fs.rmSync(dir, { recursive: true, force: true });
  die(`${name}: not ready after ${READY_TIMEOUT_MS / 1000}s — killed, session cleaned\n${tail}`);
}

async function down(args) {
  const { project } = args;
  if (!project) {
    die("usage: devserver.cjs down --project <nx-project>");
  }
  const name = sessionName(project);
  const dir = path.join(RUN_DIR, name);
  const state = readState(name);
  if (!state) {
    console.log(`${name}: no recorded session (nothing to stop)`);
    return;
  }
  if (!sessionAlive(state)) {
    if (reportPortStraggler(state)) {
      die(`${name}: recorded session is dead but its port lives — manual teardown needed (state kept)`);
    }
    fs.rmSync(dir, { recursive: true, force: true });
    console.log(`${name}: session was already dead — state and log discarded`);
    return;
  }
  const gone = await killTree(state.pid);
  const straggler = reportPortStraggler(state);
  const ports = state.ports ?? (state.port ? [state.port] : []);
  if (!gone || straggler) {
    die(`${name}: teardown incomplete — state kept for inspection`);
  }
  fs.rmSync(dir, { recursive: true, force: true });
  console.log(
    `${name}: down (tree from ${state.pid} terminated${ports.length > 0 ? `, ports ${ports.join(",")} free` : ""}, session dir removed)`,
  );
}

(async () => {
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
