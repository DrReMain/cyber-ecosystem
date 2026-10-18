# esports 迁移专项 — 存量 cutover(M4)

> 归属:esports 前沿迁移档。裁决 D-C/D-C2;窗口 = 预演 12.07-12.13(W8)、停机切换 12.14-12.17(W9,承诺日 12-17,D-K);演练日程 P6 于 12-07 前由用户确认。旧仓生产库为千行级小体量,迁移成本不在量、在语义。
> 旧系统处置(D-P):即刻冻结(仅致命 bug);切换后旧栈停机封存 2 周(期间可快速回滚),期满下线归档;部署沿用旧域名/旧服务器,切换日反代/DNS 指向新栈(细节 W8 前定)。

## 四档重算策略(D-C)

| 档 | 测试 | 处置 |
|---|---|---|
| A 明细自足 | red_dot/offense_defense/character_counter/skill_trajectory/flash_aiming/dual_zone | 原始明细全量重放重算 |
| B 聚合可重放 | number_stars/moving_target/target_recognition/multi_target | summary 聚合重放 |
| C 只能平移 | tracking_moving/economic_sim | 旧分平移+legacy 口径标 |
| D 作废 | skill_prediction | 死端点,无有效存量 |

## 逐表口径

- **身份列 → system**:username/phone/open_id/微信身份迁为会员身份+identity;credential 密码 verify-then-rehash(MD5 无法验证者强制重置);微信资料 4 列合并为一条画像 JSON。
- **画像/偏好/溯源 → esports**:出生年月/年龄段/头像/游戏偏好(含待归类项保持待归类)/邀请关系/登录溯源九件套(随 login_record 语义迁或冷存)。存量「仅年龄段无出生年月」用户(旧 API 允许)6→8 档不可精确映射:迁为画像不完整态(档位置空),首个画像触点补填出生年月后服务端重派生。
- **权益翻译**:解锁三旗标与 activation_records → entitlement 解锁记录;未核销激活码 → 新兑换码批次(保留 report_type 语义);已核销/过期码丢弃靠流水保历史。
- **不迁重算**:user 统计四列/ability_score/platform_analytics/game_analytics(由 test_record 重算)。
- **保真搬运(D-C2)**:报告 JSON 快照(user_career_profile/school_report)= 旧算法产物+R3 永久可查看语义,原样只读迁入;consent/审计从零。
- **不迁**:RBAC 六件套/menu/`*_bak_*` 快照表/flyway 历史。
- **种子重放**:game/test_type/test_config/contribution/题库四表/职业模板与统计 —— 以旧仓 init.sql 种子为内容基线重放进版本化迁移(修已知坑,见 plan.md E2)。
- **ID 与时间**:存量 Snowflake ID 沿用(不改写主键);时间戳按 Asia/Shanghai 墙钟解释转 UTC;游戏时钟毫秒保留为原始 long;duration 由 ms 差重算。

## 迁移顺序

身份(产出并持久化 旧↔新 ID 映射)→ 组织(school→租户/学校档案+管理员成员绑定)→ 权益 → 主数据种子(先于行为数据)→ 行为数据(偏好→test_record→溯源)→ 报告快照 → 收尾(同意/审计/分析重算)。

## cutover 验收清单(血泪教训条款)

1. 生产副本脱敏预演(UK 脱敏后仍唯一);空 staging 验证无效,禁止作为通过依据。
2. 预演环境补 nginx 仿真 + 生产形态造数(鸿蒙 UA/IPv6/CGNAT/软删旧行/无注册 IP 老用户)。
3. 旧→新 ID 映射持久化落盘,且覆盖软删行(上次事故根因:软删行被跳过→子表幽灵 FK)。
4. FK 解析禁止按命名约定猜表名(上次静默 skip 事故)。
5. 导入后哨兵值复位(旧 system_config 迁移哨兵 99 必须复位为业务布尔,否则新系统读到 99)。
6. 行数对账 + 抽样逐字段 + 分布 SQL(score 分位按测试分组、raw 非空计数)。
7. 回滚预案 = 迁移前快照 + 持久化映射表,任一验收失败从快照重跑。
