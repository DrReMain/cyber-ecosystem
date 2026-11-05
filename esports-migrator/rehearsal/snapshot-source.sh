#!/bin/sh
# Rehearsal source snapshot: dump the lab MySQL copy of the legacy store so
# any failed rehearsal pass can be replayed from a known point (cutover
# checklist item 7 — the rollback leg). The dump is production data,
# unsanitized by ruling 2026-11-02, so it lives outside the repo in the lab
# material area, next to the original Navicat export.
#
# Usage: snapshot-source.sh            (writes esport-<stamp>.sql.gz + .counts.txt)
set -eu

out_dir="${HOME}/Desktop/lab/esports/mig-snapshots"
mkdir -p "$out_dir"
stamp=$(date +%Y%m%d-%H%M%S)
dump="$out_dir/esport-$stamp.sql.gz"
counts="$out_dir/esport-$stamp.counts.txt"

# --add-drop-database --databases makes the dump self-contained: the restore
# needs no pre-state and lands as DROP DATABASE esport + full replay.
docker exec mysql8 sh -c 'exec mysqldump -uroot -p"$MYSQL_ROOT_PASSWORD" \
	--single-transaction --add-drop-database --databases esport' | gzip > "$dump"

# Exact row counts for every legacy table — information_schema.table_rows is
# an estimate and is NOT a reconciliation baseline. A restore must reproduce
# these numbers before a replay counts as a rollback.
tables=$(docker exec mysql8 sh -c 'exec mysql -uroot -p"$MYSQL_ROOT_PASSWORD" \
	-N -e "SHOW TABLES FROM esport"')
: > "$counts"
for t in $tables; do
	n=$(docker exec mysql8 sh -c "exec mysql -uroot -p\"\$MYSQL_ROOT_PASSWORD\" \
		-N -e 'SELECT COUNT(*) FROM esport.\`$t\`'")
	printf '%s\t%s\n' "$t" "$n" >> "$counts"
done

echo "snapshot: $dump"
echo "counts:   $counts ($(printf '%s\n' $tables | wc -l | tr -d ' ') tables)"
