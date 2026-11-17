// ledger-guard — docs/roadmap/ holds working ledgers; they never enter the
// repository (the skeleton lane's law; business lanes carry their own).
import { stagedFiles } from './lib.mjs';

if (stagedFiles().some((f) => f.startsWith('docs/roadmap/'))) {
  console.error('docs/roadmap/ holds working ledgers — never committed');
  process.exit(1);
}
