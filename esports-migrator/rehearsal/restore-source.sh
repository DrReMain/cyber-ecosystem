#!/bin/sh
# Rollback leg, source side: replay a snapshot taken by snapshot-source.sh
# back into the lab MySQL. The dump is self-contained (DROP DATABASE esport +
# full replay), so this works from any pre-state. After the replay the
# per-table row counts are regenerated and diffed against the snapshot's
# sidecar — a mismatch aborts with the table names, because a partial restore
# must never be mistaken for a clean rollback point.
#
# Usage: restore-source.sh <esport-<stamp>.sql.gz>
set -eu

dump="${1:?usage: restore-source.sh <esport-<stamp>.sql.gz>}"
sidecar="${dump%.sql.gz}.counts.txt"
[ -f "$dump" ] || { echo "no such snapshot: $dump" >&2; exit 1; }
[ -f "$sidecar" ] || { echo "no sidecar counts next to snapshot: $sidecar" >&2; exit 1; }

echo "restoring $dump ..."
gunzip -c "$dump" | docker exec -i mysql8 sh -c 'exec mysql -uroot -p"$MYSQL_ROOT_PASSWORD"'

tables=$(docker exec mysql8 sh -c 'exec mysql -uroot -p"$MYSQL_ROOT_PASSWORD" \
	-N -e "SHOW TABLES FROM esport"')
after=$(mktemp)
: > "$after"
for t in $tables; do
	n=$(docker exec mysql8 sh -c "exec mysql -uroot -p\"\$MYSQL_ROOT_PASSWORD\" \
		-N -e 'SELECT COUNT(*) FROM esport.\`$t\`'")
	printf '%s\t%s\n' "$t" "$n" >> "$after"
done

# Exit status is the verdict: zero diff = the rollback point is whole.
if diff -u "$sidecar" "$after" > "$(dirname "$dump")/restore-$(basename "$sidecar")"; then
	echo "restore complete — row counts match snapshot sidecar."
else
	echo "RESTORE MISMATCH — row counts diverge from snapshot sidecar; see the diff artifact." >&2
	exit 1
fi
rm -f "$after"
