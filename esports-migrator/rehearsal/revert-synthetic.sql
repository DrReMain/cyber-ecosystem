-- Removes exactly the rows synthetic.sql adds — same double-keyed bands
-- (id band + username/user_id prefix guard), so it can never touch a real
-- row. Use when a rehearsal pass needs the copy back at its production
-- baseline without a full snapshot restore.
--
--   docker exec -i mysql8 sh -c 'exec mysql -uroot -p"$MYSQL_ROOT_PASSWORD"' \
--     < revert-synthetic.sql

USE esport;

DELETE FROM `user`
WHERE id BETWEEN 990000001 AND 990000004
  AND username LIKE 'migsyn-%';
DELETE FROM login_record
WHERE id BETWEEN 994000001 AND 994000002
  AND user_id BETWEEN 990000001 AND 990000004;
