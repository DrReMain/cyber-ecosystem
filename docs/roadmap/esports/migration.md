# esports 迁移专项 — 存量 cutover(M4)

> 归属:esports 前沿迁移档。裁决 D-C/D-C2/D-BA;窗口 = **预演 ~11.14-11.17(W4,副本 11-14 就绪即演)、停机切换 12.01-12.03(力争)/12.08(保底;承诺日 12-17 不动作保险丝,D-K)**;演练日程 P6 于 11-12 前由用户确认。旧仓生产库为千行级小体量,迁移成本不在量、在语义。
> 旧系统处置(D-P):即刻冻结(仅致命 bug);切换后旧栈停机封存 2 周(期间可快速回滚),期满下线归档;部署沿用旧域名/旧服务器,切换日反代/DNS 指向新栈(细节 W8 前定)。
> 基线注(2026-10-23 D-AL,据旧仓 db/README 核实):生产基线=master **V1-V15**;feature/block34-groundwork(V17 协议两表/V18 8 档字典/V19 自定义归类)**未发布无存量**——同意记录从零(与下文「consent 从零」一致)、8 档字典与 game_category 无迁移面;旧库两处已知 schema 不一致(JSON 形态混用/BIGINT-DATETIME 混用)由新仓 schema 自行定夺,不承继。

## 四档重算策略(D-C)

| 档 | 测试 | 处置 |
|---|---|---|
| A 明细自足 | red_dot/offense_defense/character_counter/skill_trajectory/flash_aiming/dual_zone/resource_match | 原始明细全量重放重算 |
| B 聚合可重放 | number_stars/moving_target/target_recognition/multi_target | summary 聚合重放 |
| C 只能平移 | tracking_moving/economic_sim | 旧分平移+legacy 口径标 |
| D 作废 | skill_prediction | 死端点,无有效存量 |

> **2026-11-02 recompute 审计注**:①「平移」已由 behavior 刀对全部迁移行落地(score→authoritative_score,rawScore→raw_score,grade=band(score)),四档现况=C 档处置泛化为默认,重放是「升级」而非「必经」。②A/B 重放闸门=该款标定曲线(P16:现仅 red_dot 分位冻结;curvePendingResult 结构性不出权威分,曲线前重放=把旧标定分行降级成无权威分/无档位)。③重放窗口与口径标载体=P26.1(ljr):推荐 cutover 一刻迁移器内重放当日已定曲线款(逐字复制 scorer+曲线,bandFor 先例形状),余款平移+legacy 标,cutover 后不重放。**(11-04 补注,用户问「游戏测试完分数/评级」带出)新测侧同缝:14 款 scorer 仅 red_dot 冻结曲线在码(提交即出权威分+档位),余 13 款 curvePendingResult/raw-only 结构性不出权威分——历史行评级随本票重放口径落,新测 13 款出评级的时点=曲线从重放器进 live scorer 的入码片,未排票;建议本票裁决一并钉(搬运方=lzx,口径/窗口=ljr),在此之前新测=诚实「未评级」态(设计内非 bug,D-AO③ 文案在)**。④B 档 summary 聚合重放可行性:旧 summary 聚合面对得上新 composite 输入(number_stars 例:maxSeqLen 缺席→复杂度项降级重放);逐款咬合核对归 recompute 机械片夹具。

> **2026-11-06 补行注**:A 档补 resource_match——生产存量 101 行(在 2698 总量/2685 迁移量对账内),原表 13 款漏列;归属依据=metric.md §6.11 E3-h 定稿卡(Tier A,事实流全量重算),重算闸门同 P16 曲线律(曲线前平移)。档位依 metric 定稿证据落(D-BC 报备)。

## 逐表口径

- **身份列 → system**:username/phone/open_id/微信身份迁为会员身份+identity;**密码口径(11-06 实数定谳)**:纯 BCRYPT digest 直迁(生产 0 行),MD5 274 行(250 空=微信自注族;余 24=真 hex 8+todeleteuser 软删族字面明文 16)+BCRYPT_OVER_MD5 11 行一律关闭密码面——不可离线验证,兼容零入产品(形态与新栈 provider-born 无密码账号同构),登录走手机/微信凭据(照迁);设过真密码者仅 19 人;活行无凭据者 lockedOut(实数 6)预演对账;微信 4 列口径修订(11-02 随首刀定):wechat_sex→sex 枚举已迁,地区 3 列新模型无消费面不迁、预演计数报备。
- **画像/偏好/溯源 → esports**:出生年月/年龄段/头像/游戏偏好(含待归类项保持待归类)/邀请关系/登录溯源九件套(随 login_record 语义迁或冷存)。存量「仅年龄段无出生年月」用户(旧 API 允许)6→8 档不可精确映射:迁为画像不完整态(档位置空),首个画像触点补填出生年月后服务端重派生。
- **权益翻译**:解锁三旗标与 activation_records → entitlement 解锁记录;未核销激活码 → 新兑换码批次(保留 report_type 语义);已核销/过期码丢弃靠流水保历史。
- **不迁重算**:user 统计四列/ability_score/platform_analytics/game_analytics(由 test_record 重算)。**2026-11-02 recompute 审计**:四目标全数零迁移器写入——platform/game analytics 新 schema 未建(归 P20 特性从 test_record 现算),user 统计四列=读时推导(新 profile 无统计列,原型统计条走 API 现算),ability_score=report 冻结流懒写(读面幂等生成,迁移器抢写反而破坏懒冻结语义);「由 test_record 重算」的正确读法=新系统读时自行重算,迁移器不碰。
- **保真搬运(D-C2)**:报告 JSON 快照(user_career_profile/school_report)= 旧算法产物+R3 永久可查看语义,原样只读迁入;consent/审计从零。(生产实况 11-02 核实:school_report 无存量,组织层随之删除,见迁移顺序注;**11-06 真转储证伪——463 行实存,处置=D-MIG-1,见预演备料节**)
- **不迁**:RBAC 六件套(`admin_user`/`admin_user_role`/`role`/`role_permission`/`permission`/`menu`)/`*_bak_*` 快照表/flyway 历史/`sku_pricing`(旧报告三档定价表,D-AB 新计费取代且旧库无网关无存量交易)/`content_management`(banner/notice/guide kv 内容表,旧码零消费端=死表;notice 面由 MSG-1 公告承接)/`system_config`(逐键核对新栈配置面,默认不迁)。
- **表名对照(2026-11-02 全 36 表逐名过账,处置组↔表名,零命中补齐)**:身份=`user`;画像行为=`user_game_preference`/`login_record`(溯源,P22)/`user_referral`(归因,P23);权益=`activation_code`/`activation_records`+`user` 解锁三旗标列;报告快照=`user_career_profile`/`school_report`(**463 行实存**(11-06 转储),处置=D-MIG-1,guard 遇非空仍停);种子=`game`/`test_type`/`test_config`/`game_ability_contribution`/题库四表(`moba_character`/`offense_scenario`/`character_counter_relation`/`offense_choice_explanation`)/`career_recommendation_template`/`career_statistics`/`test_type_introduction`(档案面随余组);重算=`ability_score`/`platform_analytics`/`game_analytics`+`user` 统计四列;MSG 域=`user_message`/`user_feedback`(随 MSG-1;未读语义旧仓实锤=A1,见 prd.md P21;**11-04 域落齐=新四表 user_message/announcement/user_feedback/feedback_message,迁移映射三则=旧 type system→announcement、旧六类前自由串 category(suggestion/bug/other)→新枚举 other/feature_bug/other 兜 other、旧 ref_type+ref_id 收窄为 ref_id(类型即定域);旧单发 admin_reply/replied_at 拆入 feedback_message 线程(admin 行带 admin_id);announcement/feedback_message 无旧表=纯新增;旧 is_read 行级未读同形直迁**);待裁=`app_version`(P19,证据倾向 B:旧 H5 版本弹窗实存)。
- **种子重放**:game/test_type/test_config/contribution/题库四表/职业模板与统计 —— **内容真相源=生产转储(2026-11-06 重锚,证据见预演备料节)**,init.sql 降辅助参考:实锤=生产 test_type.ability 已按 09-26 GAME_LOGIC_SPEC「数据层 4 处修正」批执行(target_recognition→reactionSpeed/moving_target→trackingStability/dual_zone 蛇形修),init.sql 仍是修正前世代;gac 16 行与题库正典(20 场景×60 解析)两源逐行一致。新仓载体=bootstrap/seed.go(E2;init 迁移零 SQL 种子);**归因悬点(校准片 P16/P26.1 第一裁决项):moving_target 主归因 seed.go=trackingStability(随生产)vs gac 主行=aimingAccuracy(规格 B2 将其 TS 列 0.5 次行)**。副能力终态=4 行 0.5(09-26 拍板=生产现存 3 行+flash_aiming→DE 新增,生产未及上线)——校准勿按生产 gac 回「修」成 3 行。题库垃圾:生产另有 2 条手工场景(「222」/「测试场景xx」)被 2 行 offense_defense 明细引用,重放按 DETAIL_MISMATCH 拒后落平移(设计内),预演对账预期之。修已知坑清单随 E 线入 progress-archive。
- **ID 与时间**:存量 Snowflake ID 沿用(不改写主键);时间戳按 Asia/Shanghai 墙钟解释转 UTC;游戏时钟毫秒保留为原始 long;duration 由 ms 差重算。

## 迁移顺序

身份(产出并持久化 旧↔新 ID 映射)→ 权益 → 主数据种子(先于行为数据)→ 行为数据(偏好→test_record→溯源)→ 报告快照 → 收尾(同意/审计/分析重算)。

**「组织(school→租户)」层删除(11-02 D-BA:生产数据全公版,旧版无专版,无学校数据可迁;ljr 补证 school 域存量管理员/组织/学员/report 均测试数据不迁)**——导入器遇 school/school_report 非空行=停机报错待议,不静默跳过(血泪条款 4 同族防线);真租户随 M3 出现再立二次迁移。(**11-06 census 证伪「均为测试数据」:463 行中 22 行属 7 真实存活用户、121 行 user_id 孤儿——事实前提推翻,行级处置挂 D-MIG-1;层删除裁定本体待裁决后复核**)

## 预演备料实录(2026-11-06,副本较 11-14 就绪门提前 8 天)

- **物料落地(血泪条款 1 首次实做)**:生产备份直载本地——lab compose `mysql8`(宿主 13306,utf8mb4/+8:00 与源对齐)建 `esport` schema 后导入 106MB Navicat 转储(11-05 22:27 生产导出),一次绿。
- **谱系三源互证(裁「生产是否跑过 main 外的表结构更新」)**:①生产 flyway 台账=BASELINE(07-21)+V1..V15(末笔 09-04「user device type brand」),V16+ 零痕;②lab 五目录=旧仓全量 git 克隆,master HEAD(911e125)=09-04 与 V15 同日同代、其后零提交,`feature/block34-groundwork`(09-27)=V17 协议两表(规格入库、**代码未实现**)/V18 年龄段字典/V19 自定义归类;③转储 37 表无任何 V17+ 产物。结论:**master 即生产世代代码,block34 未发布;转储=唯一 schema 真相源**——init.sql 仅 19 表,是 CI 建库残缺基线(master db0f654 自认「建库基线缺失」),其 `user.nickname NOT NULL` 等声明以转储为准(nickname 实可空,9/285 NULL);V1-V15 实体 SQL 散在 esport-admin/sql/ 具名脚本与 manage/sql/(雪flake FK 修复族)。
- **census 真数据首跑三绿(identity 除冲突停机)**:user 285(260 活/25 软删;openid NULL 255、birth NULL 157、username NULL 30);entitlement 绿(era-2 school 旗标 257 行弃=默认 1 不可归因;**334 码全为全局型→0 迁**;孤儿 activation_records 2);behavior 绿(2698→2685 迁,skill_prediction 弃 13,自报旗标 2125,login_record 217 冷存,user_referral 6 待 P23)。
- **guard 必停=预演第一脚**:school=4(测试学校×3 软删+**存活零成员「萧山中学」**);school_report=463=342 行挂 10 用户(**全部未绑校**:320 行=Aqiu 重注册族软删行,**22 行=7 真实存活用户**:鲁智深 12/kk 3/玩家0866 2/黄皓予 2/陵川 1/黑豆 1/Amy 1)+121 行 user_id 孤儿;另有悬空 school_id=162(无 school 行)。
- **随片两修+执行面**:srcUser.Nickname→NullString;test_record 读取改 PK 序+Go 侧四键排序(MySQL 1038 宽行 filesort 爆 buffer——cutover 宿主 buffer 不可假设)。migrator build+race test+lint 绿。直执 go 已禁(11-06 用户指示)→临时 nx target `esports-migrator:run/check`(project.json+README 带 TODO 拆除清单;DSN=~/.xgwl-secrets/esports-mig/source.env 引号形)。
- **回滚腿+造数件毕并彩排(11-06,清单 2/7 两项)**:`esports-migrator/rehearsal/` 四件=snapshot/restore 脚本对(dump 出仓至 lab mig-snapshots/,restore 后全表精确 counts 对 sidecar,不一致即.abort)+synthetic/revert 造数对。彩排链全绿:快照(37 表 sidecar)→造数(migsyn 4 用户+2 登录行;behavior census login_record 217→219 增量精确对上)→restore(exit 0,37 表 counts 全对,造数抹净回基线)→双应用幂等(仍 4 行)+revert 净。造数口径:鸿蒙 NEXT/IPv6/CGNAT/组合四形态为真缺面(生产存量 OpenHarmony 6.1 仅 2 行、IPv6=0);软删 25 行/无注册 IP 117 行**真实存量已足不造**;造数行 id 段 990000xxx/994000xxx(低于雪花段,预查零撞),demo 密码 MD5(migsyn-demo)走 verify-then-rehash 走查。identity census 打印被 D-MIG-2 冲突组拦在 buildPlans(11-06 实测)→造数对账探针=SQL 直查,预演日同理。

### 预演前决策(拦 11-14/15 绿门)

1. **D-MIG-1 school 域处置**(guard 停机即此):证据如上。选项 (a) 裁「测试/残留不迁」+guard 显式旁路旗标(默认仍停、运行日志留证、cutover 夜同旗标)——ljr 需对「7 真实用户 22 行」复核签字;(b) 22 行保真迁(需新 schema 落点,预演前代价大);(c) 封存旧栈只读承接(受 D-P 两周封存窗约束,非永久)。推荐 (a)。
2. **D-MIG-2 账号冲突消解**(identity 停机即此):6 组 23 行=131****8167×7、Aqiu openid×2、test×2、test2×2、todeleteuser×5(全软删)、newtestuser×5(全存活同手机号 13900000000);混合组活行全为雪flake 行。推荐机械规则(零产品判断/零丢行/可审计):组内存活行得原账号,多存活取 updated_at 最新;其余账号改 `<account>#<legacy_id>`(软删身份登录面已死,migration_map 全量留痕,census 打印改写计数)。替代=测试族整批排除(丢数据,须产品裁)。

## 预演执行稿(2026-11-06 预制;回滚腿与造数件已彩排,见备料节)

前置门:D-MIG-1/2 裁决形(拦 S3);snapshot/recompute 是否入演随 P26/P26.1(理想 11-12)。演练日程 P6 已代裁定 11-15;副本 11-06 已就绪,裁回皆可提前(报备制)。

| 步 | 动作 | 判据 |
|---|---|---|
| S0 | `rehearsal/snapshot-source.sh` 打预演态快照 | 37 表 counts sidecar 落(lab mig-snapshots/) |
| S1 | `rehearsal/synthetic.sql` 造数(四形态 4 用户+2 登录行) | migsyn 4 行 SQL 证;login_record 217→219 |
| S2 | 目标库重建:drop/recreate → atlas migrate → bootstrap seed | 种子重放绿(2 垃圾场景 DETAIL_MISMATCH 拒后平移=设计内预期;副能力终态 4 行勿回修 3) |
| S3 | `--phase all` 五阶段(guard 按 D-MIG-1 裁决形:(a)=显式旁路旗标+运行日志留证) | 各阶段 census+写入计数入录 |
| S4 | `--phase reconcile --sample N` | 全绿零 fail |
| S5 | 新栈走查:造数账号密码登录(MD5 verify-then-rehash 亲证)/鸿蒙 UA 访问/IPv6·CGNAT 审计形态 | 走查录;nginx 仿真=本机反代近似(真形状 W8 前定,条目挂账) |
| S6 | **绿门**=对账全绿+走查过+migration_map 落盘(含软删行) | 任一不过 → S7 |
| S7 | 回滚腿:目标侧=重走 S2-S3(可重建);源侧只读无需回滚,误写则 `restore-source.sh <快照>` | 11-06 彩排证:restore 后全表 counts 对 sidecar |

对账预期锚(基线 11-06 census,造数增量随 S1):user 285(260 活/25 删)+4 活;phone 凭据 distinct+4;login_record 冷存 217+2;entitlement/behavior 面不变(造数不触解锁/偏好/成绩);school 停项 4/463 按 D-MIG-1 裁决形消解;账号冲突 6 组 23 行按 D-MIG-2 规则改写计数入 census 打印。

## cutover 验收清单(血泪教训条款)

1. 生产备份直载本地预演(**不脱敏——2026-11-02 用户裁定**,毕了销毁细则待定);空 staging 验证无效,禁止作为通过依据。
2. 预演环境补 nginx 仿真 + 生产形态造数(鸿蒙 UA/IPv6/CGNAT/软删旧行/无注册 IP 老用户)。
3. 旧→新 ID 映射持久化落盘,且覆盖软删行(上次事故根因:软删行被跳过→子表幽灵 FK)。
4. FK 解析禁止按命名约定猜表名(上次静默 skip 事故)。
5. 导入后哨兵值复位(旧 system_config 迁移哨兵 99 必须复位为业务布尔,否则新系统读到 99)。
6. 行数对账 + 抽样逐字段 + 分布 SQL(score 分位按测试分组、raw 非空计数)。
7. 回滚预案 = 迁移前快照 + 持久化映射表,任一验收失败从快照重跑。
