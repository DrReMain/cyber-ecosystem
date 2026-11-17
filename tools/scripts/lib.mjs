// Shared helpers for the gate and guard scripts: a shell-free nx entry
// (.nx/nxw.js is the wrapper's js body — the official cross-platform way to
// drive nx without the bash wrapper) and shell-free git plumbing. No POSIX
// shell syntax anywhere: every script runs identically under macOS, Linux,
// and Windows.
import { execFileSync, spawnSync } from 'node:child_process';

export function git(...args) {
  return execFileSync('git', args, { encoding: 'utf8' }).trim();
}

export function stagedFiles() {
  return git('diff', '--cached', '--name-only').split('\n').filter(Boolean);
}

export function nx(args) {
  const r = spawnSync(process.execPath, ['.nx/nxw.js', ...args], { stdio: 'inherit' });
  if (r.status !== 0) process.exit(r.status ?? 1);
}
