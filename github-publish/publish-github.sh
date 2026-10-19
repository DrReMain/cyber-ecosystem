#!/usr/bin/env bash
# Publish a clean snapshot of this repository to the public GitHub mirror.
#
# The public branch carries one commit per milestone ("snapshot chain"):
# each new snapshot's parent is the previous public head, so the public
# history shows milestone progression without any internal commit detail.
# A public head is recognized by its "release: " subject prefix. When the
# remote head predates this convention (legacy detailed history), the
# snapshot starts a fresh root and force-pushes over it — pass
# --prune-tags on that first publish to also drop tags anchoring the old
# history.
#
# Snapshots are built with plumbing commands against a throwaway index,
# so the working tree and the current branch are never touched. The
# snapshot always reflects the committed state of the source ref.
#
# Usage:
#   ./nx run github-publish:publish -- [options]
# Options:
#   -m <msg>       commit subject; auto-prefixed with "release: " if missing
#   -r <ref>       source ref to snapshot (default HEAD)
#   -t <tag>       annotated tag to publish on the snapshot commit
#   --prune-tags   delete every remote tag before pushing (first publish)
#   --dry-run      build and verify the snapshot, do not push
set -euo pipefail

REMOTE_NAME=github
REMOTE_URL=https://github.com/DrReMain/cyber-ecosystem.git
PUBLIC_BRANCH=main
MARKER="release: "
EXCLUDE_FILE=github-publish/publish-github.exclude

MSG=""
REF="HEAD"
TAG=""
PRUNE_TAGS=0
DRY_RUN=0
while [[ $# -gt 0 ]]; do
  case "$1" in
    -m) MSG="${2:?}"; shift 2 ;;
    -r) REF="${2:?}"; shift 2 ;;
    -t) TAG="${2:?}"; shift 2 ;;
    --prune-tags) PRUNE_TAGS=1; shift ;;
    --dry-run) DRY_RUN=1; shift ;;
    *) echo "error: unknown option: $1" >&2; exit 1 ;;
  esac
done

cd "$(git rev-parse --show-toplevel)"
[[ -f "$EXCLUDE_FILE" ]] || { echo "error: $EXCLUDE_FILE not found" >&2; exit 1; }
[[ $MSG == "$MARKER"* ]] || MSG="${MARKER}${MSG:-snapshot $(date -u +%Y-%m-%d)}"

SRC_COMMIT=$(git rev-parse --verify "${REF}^{commit}")

git remote get-url "$REMOTE_NAME" >/dev/null 2>&1 || git remote add "$REMOTE_NAME" "$REMOTE_URL"
git fetch "$REMOTE_NAME" "+refs/heads/${PUBLIC_BRANCH}:refs/remotes/${REMOTE_NAME}/${PUBLIC_BRANCH}"

# Parent = previous public snapshot; a legacy head (no marker) means fresh root.
PARENT=""
PARENT_KIND="fresh root (force-push over legacy history)"
if git rev-parse -q --verify "${REMOTE_NAME}/${PUBLIC_BRANCH}" >/dev/null; then
  REMOTE_HEAD=$(git rev-parse "${REMOTE_NAME}/${PUBLIC_BRANCH}")
  REMOTE_SUBJECT=$(git log -1 --format=%s "$REMOTE_HEAD")
  if [[ $REMOTE_SUBJECT == "$MARKER"* ]]; then
    PARENT="$REMOTE_HEAD"
    PARENT_KIND="chain (parent is the previous public snapshot)"
  fi
fi

# Build the public tree in a throwaway index: source tree minus exclusions.
TMPIDX=$(mktemp)
trap 'rm -f "$TMPIDX"' EXIT
export GIT_INDEX_FILE=$TMPIDX
git read-tree "$SRC_COMMIT"
SRC_COUNT=$(git ls-files | wc -l | tr -d ' ')

while IFS= read -r path || [[ -n "$path" ]]; do
  [[ -z "$path" || "$path" == \#* ]] && continue
  if ! git ls-files --error-unmatch "$path" >/dev/null 2>&1; then
    echo "warn: exclude entry not in source tree (typo?): $path" >&2
    continue
  fi
  git rm -r -f --cached --quiet --ignore-unmatch -- "$path"
  if [[ -n "$(git ls-files -- "$path")" ]]; then
    echo "error: exclusion still present in tree: $path" >&2; exit 1
  fi
  echo "excluded: $path"
done < "$EXCLUDE_FILE"

PUBLIC_COUNT=$(git ls-files | wc -l | tr -d ' ')
TREE=$(git write-tree)
unset GIT_INDEX_FILE

if [[ -n "$PARENT" ]]; then
  COMMIT=$(git commit-tree "$TREE" -p "$PARENT" -m "$MSG")
else
  COMMIT=$(git commit-tree "$TREE" -m "$MSG")
fi

echo
echo "source:      $REF ($SRC_COMMIT, $SRC_COUNT files)"
echo "public tree: $TREE ($PUBLIC_COUNT files)"
echo "public commit: $COMMIT"
echo "parent:      ${PARENT:-none} — $PARENT_KIND"
echo "subject:     $MSG"
[[ -n "$TAG" ]] && echo "tag:         $TAG"
[[ $DRY_RUN -eq 1 ]] && { echo "dry-run: nothing pushed"; exit 0; }

if [[ $PRUNE_TAGS -eq 1 ]]; then
  # Drop every remote tag except the one about to be published.
  while read -r t; do
    [[ -z "$t" || "$t" == "$TAG" ]] && continue
    echo "deleting remote tag: $t"
    git push "$REMOTE_NAME" --delete "refs/tags/$t"
  done < <(git ls-remote --tags "$REMOTE_NAME" | awk '{print $2}' | sed 's|^refs/tags/||' | grep -v '\^{}$' || true)
fi

git push --force-with-lease "$REMOTE_NAME" "${COMMIT}:refs/heads/${PUBLIC_BRANCH}"

if [[ -n "$TAG" ]]; then
  git tag -a -f "$TAG" -m "$MSG" "$COMMIT"
  git push "$REMOTE_NAME" "refs/tags/$TAG"
fi

echo "published $MSG -> $REMOTE_NAME/$PUBLIC_BRANCH ($COMMIT)"
