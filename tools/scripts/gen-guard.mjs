// gen-guard — staged gen/ paths are legal only beside the contract or the
// generator bump that produces them; anything else is a hand edit to
// derived output.
import { stagedFiles } from './lib.mjs';

const staged = stagedFiles();
const touchesGen = staged.some((f) => f.startsWith('gen/'));
const carriesSource = staged.some((f) => f.endsWith('.proto') || f === 'pnpm-lock.yaml');
if (touchesGen && !carriesSource) {
  console.error("gen/ is derived — never hand-edit; stage the .proto (or the generator-bump's pnpm-lock) and run node .nx/nxw.js run proto:generate");
  process.exit(1);
}
