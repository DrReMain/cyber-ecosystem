// gates:drift — regenerate everything, then assert the working tree did not
// move: a changed tree means gen/ drifted from its sources (hand edits or a
// stale regeneration). The regenerated truth is left in the tree for review.
import { execFileSync } from 'node:child_process';
import { nx } from './lib.mjs';

const porcelain = () => execFileSync('git', ['status', '--porcelain'], { encoding: 'utf8' });
const before = porcelain();
nx(['run', 'proto:generate']);
nx(['run-many', '-t', 'generate', '--all']);
if (porcelain() !== before) {
  console.error('gates:drift failed - regeneration changed the working tree (the regenerated truth is now there: review and commit it, or fix the source), or generation errored; see output above');
  process.exit(1);
}
