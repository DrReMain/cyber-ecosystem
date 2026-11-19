// The affected base for the pre-push gates: the branch's upstream when it
// has one, else the merge-base against the lane's default branch. A lane is
// the first path segment of the branch name;
// an unlabeled branch belongs to the skeleton lane (main). The merge-base
// base walks every commit a no-upstream branch is about to push, where the
// old HEAD~1 fallback saw only the last one. An orphan lane shares no
// history with any lane default (merge-base exits 1), so its first push
// gates the whole tree against the empty tree; after that push the upstream
// exists and the normal path resumes.
import { git } from './lib.mjs';

export function affectedBase() {
  try {
    return git('rev-parse', '--abbrev-ref', '@{upstream}');
  } catch {
    // no upstream yet (first push of a new branch)
  }
  const lane = git('branch', '--show-current').split('/')[0];
  let base;
  try {
    base = git('rev-parse', '--verify', '-q', `origin/${lane}/main`);
  } catch {
    try {
      base = git('rev-parse', '--verify', '-q', 'origin/main');
    } catch {
      return 'HEAD~1';
    }
  }
  try {
    return git('merge-base', 'HEAD', base);
  } catch {
    // unrelated histories: an orphan lane's first push gates its whole
    // tree. The empty tree is only addressable while the object store
    // holds it, so write it and return what the write produced — never
    // a spelled-out sha, which can drift from the object it names.
    return git('hash-object', '-w', '-t', 'tree', '--stdin');
  }
}

if (process.argv[2] === '--print') console.log(affectedBase());
