-- Rehearsal-only synthetic rows for the lab copy of the legacy store
-- (cutover checklist item 2 — production-shaped data the current copy
-- lacks). Four user rows + two login rows cover the shapes the real copy
-- has zero or thin coverage of: a HarmonyOS NEXT device (the copy only saw
-- OpenHarmony 6.1), an IPv6 registrant, a CGNAT (100.64/10) registrant, and
-- the combined HarmonyOS+IPv6 case. Shapes the copy already has for real —
-- soft-deleted users (25) and register_ip-less old users (117) — are NOT
-- synthesized: real rows are better evidence than manufactured ones.
--
-- Idempotent: delete-then-insert over fixed id bands (user 990000001-4,
-- login_record 994000001-2) that sit below the snowflake segment and were
-- asserted empty. Every marker is double-keyed (id band + username/invite
-- prefix) so a revert can never touch a real row. The demo password is the
-- md5 of "migsyn-demo" with password_algorithm='MD5' — production shape
-- (the census found zero BCRYPT rows); per the 11-06 census ruling the
-- migrated account's password plane closes (forced reset), so the
-- walkthrough leg for these rows is SMS-code login, never password login
-- (no verify-then-rehash path exists). Apply inside the container:
--   docker exec -i mysql8 sh -c 'exec mysql -uroot -p"$MYSQL_ROOT_PASSWORD"' \
--     < synthetic.sql

USE esport;

DELETE FROM `user`
WHERE id BETWEEN 990000001 AND 990000004
  AND username LIKE 'migsyn-%';
DELETE FROM login_record
WHERE id BETWEEN 994000001 AND 994000002
  AND user_id BETWEEN 990000001 AND 990000004;

INSERT INTO `user`
	(id, username, password, password_algorithm, phone, nickname, status,
	 is_deleted, invite_code, register_ip, last_login_ip,
	 last_device_type, last_device_brand, last_device_model, last_os,
	 last_client_type, created_at, updated_at)
VALUES
	(990000001, 'migsyn-harmony',
	 '08950ecfaaa017613e694f87184c4313', 'MD5', '19900000901',
	 '预演造数鸿蒙', 1, 0, 'MIGSYN01', '59.101.12.34', '59.101.12.34',
	 'phone', 'HUAWEI', 'Mate 70 Pro', 'HarmonyOS NEXT 5.0', 'other',
	 '2026-11-06 10:00:00', '2026-11-06 10:00:00'),
	(990000002, 'migsyn-ipv6',
	 '08950ecfaaa017613e694f87184c4313', 'MD5', '19900000902',
	 '预演造数六', 1, 0, 'MIGSYN02',
	 '2408:8456:aa1:1111:2222:3333:4444:5555',
	 '2408:8456:aa1:1111:2222:3333:4444:5555',
	 NULL, NULL, NULL, NULL, NULL,
	 '2026-11-06 10:00:00', '2026-11-06 10:00:00'),
	(990000003, 'migsyn-cgnat',
	 '08950ecfaaa017613e694f87184c4313', 'MD5', '19900000903',
	 '预演造数卡', 1, 0, 'MIGSYN03', '100.71.12.34', '100.71.12.34',
	 NULL, NULL, NULL, NULL, NULL,
	 '2026-11-06 10:00:00', '2026-11-06 10:00:00'),
	(990000004, 'migsyn-harmony-ipv6',
	 '08950ecfaaa017613e694f87184c4313', 'MD5', '19900000904',
	 '预演造数组', 1, 0, 'MIGSYN04',
	 '2408:8456:aa1:1111:2222:3333:4444:5555',
	 '2408:8456:aa1:1111:2222:3333:4444:5555',
	 'phone', 'HUAWEI', 'Mate 70 Pro', 'HarmonyOS NEXT 5.0', 'other',
	 '2026-11-06 10:00:00', '2026-11-06 10:00:00');

INSERT INTO login_record
	(id, user_id, login_method, client_type, user_agent, login_ip,
	 has_ref, created_at)
VALUES
	(994000001, 990000001, 'password', 'other',
	 'Mozilla/5.0 (Phone; OpenHarmony 5.0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.36 ArkWeb/4.1.5.1 Mobile',
	 '2408:8456:aa1:1111:2222:3333:4444:5555', 0, '2026-11-06 10:05:00'),
	(994000002, 990000003, 'password', NULL,
	 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/130.0.0.0 Safari/537.36',
	 '100.71.12.34', 0, '2026-11-06 10:06:00');
