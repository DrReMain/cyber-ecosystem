# esports 迁移专项 — 存量 cutover(M4)

> 归属:esports 前沿迁移档。裁决 D-C/D-C2/D-BA;窗口 = **预演 ~11.14-11.17(W4,副本 11-14 就绪即演)、停机切换 12.01-12.03(力争)/12.08(保底;承诺日 12-17 不动作保险丝,D-K)**;演练日程 P6 于 11-12 前由用户确认。旧仓生产库为千行级小体量,迁移成本不在量、在语义。
> 旧系统处置(D-P):即刻冻结(仅致命 bug);切换后旧栈停机封存 2 周(期间可快速回滚),期满下线归档;部署沿用旧域名/旧服务器,切换日反代/DNS 指向新栈(细节 W8 前定)。
> 基线注(2026-10-23 D-AL,据旧仓 db/README 核实):生产基线=master **V1-V15**;feature/block34-groundwork(V17 协议两表/V18 8 档字典/V19 自定义归类)**未发布无存量**——同意记录从零(与下文「consent 从零」一致)、8 档字典与 game_category 无迁移面;旧库两处已知 schema 不一致(JSON 形态混用/BIGINT-DATETIME 混用)由新仓 schema 自行定夺,不承继。

## 四档重算策略(D-C)

| 档 | 测试 | 处置 |
|---|---|---|
| A 明细自足 | red_dot/offense_defense/character_counter/skill_trajectory/flash_aiming/dual_zone | 原始明细全量重放重算 |
| B 聚合可重放 | number_stars/moving_target/target_recognition/multi_target | summary 聚合重放 |
| C 只能平移 | tracking_moving/economic_sim | 旧分平移+legacy 口径标 |
| D 作废 | skill_prediction | 死端点,无有效存量 |

> **2026-11-02 recompute 审计注**:①「平移」已由 behavior 刀对全部迁移行落地(score→authoritative_score,rawScore→raw_score,grade=band(score)),四档现况=C 档处置泛化为默认,重放是「升级」而非「必经」。②A/B 重放闸门=该款标定曲线(P16:现仅 red_dot 分位冻结;curvePendingResult 结构性不出权威分,曲线前重放=把旧标定分行降级成无权威分/无档位)。③重放窗口与口径标载体=P26.1(ljr):推荐 cutover 一刻迁移器内重放当日已定曲线款(逐字复制 scorer+曲线,bandFor 先例形状),余款平移+legacy 标,cutover 后不重放。④B 档 summary 聚合重放可行性:旧 summary 聚合面对得上新 composite 输入(number_stars 例:maxSeqLen 缺席→复杂度项降级重放);逐款咬合核对归 recompute 机械片夹具。

## 逐表口径

- **身份列 → system**:username/phone/open_id/微信身份迁为会员身份+identity;credential 密码 verify-then-rehash(MD5 无法验证者强制重置);微信 4 列口径修订(11-02 随首刀定):wechat_sex→sex 枚举已迁,地区 3 列新模型无消费面不迁、预演计数报备。
- **画像/偏好/溯源 → esports**:出生年月/年龄段/头像/游戏偏好(含待归类项保持待归类)/邀请关系/登录溯源九件套(随 login_record 语义迁或冷存)。存量「仅年龄段无出生年月」用户(旧 API 允许)6→8 档不可精确映射:迁为画像不完整态(档位置空),首个画像触点补填出生年月后服务端重派生。
- **权益翻译**:解锁三旗标与 activation_records → entitlement 解锁记录;未核销激活码 → 新兑换码批次(保留 report_type 语义);已核销/过期码丢弃靠流水保历史。
- **不迁重算**:user 统计四列/ability_score/platform_analytics/game_analytics(由 test_record 重算)。**2026-11-02 recompute 审计**:四目标全数零迁移器写入——platform/game analytics 新 schema 未建(归 P20 特性从 test_record 现算),user 统计四列=读时推导(新 profile 无统计列,原型统计条走 API 现算),ability_score=report 冻结流懒写(读面幂等生成,迁移器抢写反而破坏懒冻结语义);「由 test_record 重算」的正确读法=新系统读时自行重算,迁移器不碰。
- **保真搬运(D-C2)**:报告 JSON 快照(user_career_profile/school_report)= 旧算法产物+R3 永久可查看语义,原样只读迁入;consent/审计从零。(生产实况 11-02 核实:school_report 无存量,组织层随之删除,见迁移顺序注)
- **不迁**:RBAC 六件套(`admin_user`/`admin_user_role`/`role`/`role_permission`/`permission`/`menu`)/`*_bak_*` 快照表/flyway 历史/`sku_pricing`(旧报告三档定价表,D-AB 新计费取代且旧库无网关无存量交易)/`content_management`(banner/notice/guide kv 内容表,旧码零消费端=死表;notice 面由 MSG-1 公告承接)/`system_config`(逐键核对新栈配置面,默认不迁)。
- **表名对照(2026-11-02 全 36 表逐名过账,处置组↔表名,零命中补齐)**:身份=`user`;画像行为=`user_game_preference`/`login_record`(溯源,P22)/`user_referral`(归因,P23);权益=`activation_code`/`activation_records`+`user` 解锁三旗标列;报告快照=`user_career_profile`/`school_report`(无存量,D-BA 停机守卫);种子=`game`/`test_type`/`test_config`/`game_ability_contribution`/题库四表(`moba_character`/`offense_scenario`/`character_counter_relation`/`offense_choice_explanation`)/`career_recommendation_template`/`career_statistics`/`test_type_introduction`(档案面随余组);重算=`ability_score`/`platform_analytics`/`game_analytics`+`user` 统计四列;MSG 域=`user_message`/`user_feedback`(随 MSG-1;未读语义旧仓实锤=A1,见 prd.md P21);待裁=`app_version`(P19,证据倾向 B:旧 H5 版本弹窗实存)。
- **种子重放**:game/test_type/test_config/contribution/题库四表/职业模板与统计 —— 以旧仓 init.sql 种子为内容基线重放进版本化迁移(修已知坑,见 plan.md E2)。
- **ID 与时间**:存量 Snowflake ID 沿用(不改写主键);时间戳按 Asia/Shanghai 墙钟解释转 UTC;游戏时钟毫秒保留为原始 long;duration 由 ms 差重算。

## 迁移顺序

身份(产出并持久化 旧↔新 ID 映射)→ 权益 → 主数据种子(先于行为数据)→ 行为数据(偏好→test_record→溯源)→ 报告快照 → 收尾(同意/审计/分析重算)。

**「组织(school→租户)」层删除(11-02 D-BA:生产数据全公版,旧版无专版,无学校数据可迁;ljr 补证 school 域存量管理员/组织/学员/report 均测试数据不迁)**——导入器遇 school/school_report 非空行=停机报错待议,不静默跳过(血泪条款 4 同族防线);真租户随 M3 出现再立二次迁移。

## cutover 验收清单(血泪教训条款)

1. 生产备份直载本地预演(**不脱敏——2026-11-02 用户裁定**,毕了销毁细则待定);空 staging 验证无效,禁止作为通过依据。
2. 预演环境补 nginx 仿真 + 生产形态造数(鸿蒙 UA/IPv6/CGNAT/软删旧行/无注册 IP 老用户)。
3. 旧→新 ID 映射持久化落盘,且覆盖软删行(上次事故根因:软删行被跳过→子表幽灵 FK)。
4. FK 解析禁止按命名约定猜表名(上次静默 skip 事故)。
5. 导入后哨兵值复位(旧 system_config 迁移哨兵 99 必须复位为业务布尔,否则新系统读到 99)。
6. 行数对账 + 抽样逐字段 + 分布 SQL(score 分位按测试分组、raw 非空计数)。
7. 回滚预案 = 迁移前快照 + 持久化映射表,任一验收失败从快照重跑。
