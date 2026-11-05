# 需求总账 — 85 页原型 × 旧码功能 × 现行处置(单一真值)

> 建立:2026-10-30 roadmap 重构役(用户令)。目的:进度对账不再依赖 progress 快照——**需求分母与每页处置只在本档维护**,plan.md 排片、progress.md 认领、coverage/两矩阵的状态列自本日起退役为快照。
> 防失真:继承本目录 README 五条协议(清单计数/出处必填/计数收口)。本档出处=四路穷举(2026-10-30,agent 并行):①旧码双后端(esport-server 58 端点+esport-admin 133 端点/36 表);②旧码三前端(admin-game 23 视图/ctr-game 64 视图/manage 资产);③原型 85 页逐页(`ls *.html|wc -l` 实测 44+41,无缺页);④新实现盘点(esports 11 service/37 RPC/15 表+ctr 39 路由+admin 5 页)。
> 处置词表:✅已建(面检另计)/◐部分/❌未建/⏸待裁(P 票)/🗑已裁废弃/📐底座承接/🌫原型非本期。

## §一 ctr 原型 44 页(真源=`lab/esports/manage/design-systems/ctr/prototype/`)

| # | 页 | 处置 | 归属/注 |
|---|---|---|---|
| 1 | index | 📄 | 评审工具页,非产品屏 |
| 2 | home | ◐ | 已建;续测卡/最近成绩/统计卡已接(E5-f3);面检=V 线铺开 |
| 3 | login | ✅ | E5+E10+E6c-a;G-001..009 已处置 |
| 4 | profile-nickname | ✅ | G-201 微信回填已修 |
| 5 | profile-age | ✅ | G-206 已裁修复 |
| 6 | profile-setup(games) | ✅ | B 批④去向待 ljr 一句(→R4) |
| 7 | profile-overview | ◐ | 九区在;模块集对核毕(R2 11-01,B批⑤);会员卡随 PAY3/训练趋势 Phase2 占位 |
| 8 | password | ✅ | E6c-c 三路 |
| 9 | match-genre-pick | ✅ | B 批①推荐角标已落(R2 11-01 核验补注,前批所落) |
| 10 | match-guide | ◐ | 名错位(重建 /match);面检待 |
| 11 | match-switch-confirm | ✅ | |
| 12 | test-guide | ✅ | |
| 13 | test-intro | ✅ | mt-F1 文案键已裁(→R3) |
| 14 | test-running | ✅ | E7/H-A;唯一横屏 |
| 15 | test-result | ✅ | E5-f2 |
| 16 | test-done | ✅ | D-AI 分享主 CTA |
| 17 | test-records | ❌ | 三态页未建;归 M2 余组(记录读面) |
| 18 | report-single | ◐ | 七区块在;unlock/member CTA 已 live(片3 带参入口,ref-less 墙 member-only);点击 live 走查待解锁态记录,随真通道刀/激活码走查复核 |
| 19 | report-center | ✅ | E5b-d;G-101..106 已处置 |
| 20 | report-career | ✅ | E5b-c 开放态=旧仓语义级 |
| 21 | report-cognitive | ✅ | 同上 |
| 22 | report-trend | ✅ | 同上 |
| 23 | member-detail | ✅ | PAY3 片4 建成(2c3ecb7c,11-05):三态折(活窗紫条/过期灰条含日期/never 无条)+权益 4+FAQ5+紫 CTA;走查三态+计数 0→1 活证 |
| 24 | pay-confirm | ✅ | PAY3 片2 建成(e3159a9e,11-05):P7 三态/未成年红卡+逃生双钮/缺出生回跳/模拟收银台(MockCashierService.MarkPaid);mock 全链走查毕;真通道三红条=gap-ledger 五批 |
| 25 | pay-result | ✅ | PAY3 片3 建成(d9ddd12e,11-05):四态 GetOrder 真值驱动+D-AU③ 双文案键+收据四行;五腿 live 走查毕(含 h5 手机仿真);cancel 文案偏离登记=gap-ledger 六批 |
| 26 | message-center | ✅ | /messages 双 Tab(消息/我的反馈,?t=);未读横幅+行红点+tabbar/home 双红点(2026-11-05 MSG-1 ctr 腿,live 走查毕) |
| 27 | feedback-submit | ✅ | /feedback 六类/10-500 校验逐字(2026-11-05,提交→toast→fb 队列 live 毕;G-202 接真跳随落) |
| 28 | feedback-detail | ✅ | /feedback/$feedbackId 对话式+追问翻态+关单态;notice 形=/messages/$messageId(开信即读+类型 CTA 深链)(2026-11-05 live 毕) |
| 29 | agreement | ✅ | 版本行真源化(loader 预取 GetAgreementCurrent)+登录 consent 上行(三腿登录法)随 E9 余刀落(11-01 走查毕);章节文案仍 ljr 法务定稿占位 |
| 30 | age-required | ✅ | 付费闸挂接→PAY3;逃生按钮在遗留清单 |
| 31 | bind-phone | ✅ | E6c-c |
| 32 | bind-wechat | ✅ | E6c-c 三态 |
| 33 | account-security | ✅ | E6c-c 三行真状态 |
| 34 | account-merge-confirm | ✅ | E6c-b/c 四场景 |
| 35 | resume-remind | ◐ | 续测卡逐字对齐毕(R2 11-01);形态差(卡片 vs 整页,原型 oracle=整页非浮层)挂 P 票(G-208 余族审计) |
| 36 | share-invite | ❌ | →MSG-2(D-AS②);P23 前置 |
| 37 | scheme-b-entry | ❌ | D-AS③ 窗外随 M3/P4 |
| 38 | scheme-b-guide | ❌ | 同上 |
| 39 | scheme-b-report | ❌ | 同上;P14 已解散(11-02 D-BA,生产无学校数据)→窗外随 M3 |
| 40 | exit-confirm | ✅ | H-A kit 首用;G-208 首刀已收 |
| 41 | version-update | 📋 | 已裁 B(D-BC 11-02):ctr 轻量提示+admin 版本公告页,随 M2 余组排程 |
| 42 | loading | 📄 | kit 在,消费者随各片 |
| 43 | empty | 📄 | 同上 |
| 44 | error | 📄 | 同上 |

**计数:44 = ✅29 + ◐5 + ❌5 + ⏸1 + 📄4(2026-11-05 二对账:message 族三页随 MSG-1 ctr 腿翻 ✅;❌ 余=test-records+share-invite+scheme-b 3+version-update)。** blocks/future 补充面:forgot-password ✅(/login/forgot)、change-password ✅(E6c-c)、notification-detail ✅(=feedback-detail?kind=notice 形,落 /messages/$messageId)、subscription-popup/theme-hero ⏸(P13 剪裁已定不入)、moba-6dim ⏸(P15 已裁口径)、future/22 页=远期参考不入账。

## §二 admin 原型 41 页(真源=`manage/design-systems/admin/outputs/prototype/`;侧栏 9 组权威=_nav.js)

| 处置桶 | 页 | 计数 | 归属 |
|---|---|---|---|
| ✅/◐ 已建 | 01-user-list ✅、02-user-detail ◐(assessments+membership 两 Tab 随 M2-4 首刀建成 197fdbc6,11-01 走查+双轴审查毕;余 trace(P22)/messages(已随 MSG-1 admin 腿闭合) 占位)、01-member-list ✅、02-member-detail ✅(两会员页随 M2-1 建成,10-31 走查毕)、activation-code ✅(批次码表/停用/导出/码值搜索随 M2-3 建成,11-01 走查毕)、game-config ◐(引擎参数+题库只读抽屉在;难度三档权重/上下架缺)、agreement-manage ✅、agreement-consent ✅(两协议页随 E9 首刀建成 067a89ce,11-01 走查毕;版本表/查看全文真法务文/建即当前/切换/删除拒当前版/同意记录预填+昵称搜索+CSV 导出;渠道列递延 P23,手机号搜索递延 G-306 族)、compliance-config ✅(随 E9 余刀建成 32f5cf0a,11-01 走查毕;出生采集开关/付费年龄线敏感二次确认/两条付费文案覆盖+默认折面)、sku-manage ✅、04-order-record ✅、05-order-detail ✅(三面随 M2-2 建成 9860eb34,10-31 走查+双轴审查毕;pricing 区 skus+orders 两族)、activation-code-detail ✅(批次页 $batchNo 逐码明细随 M2-3 建成 45c276d4,11-01 走查毕)、feedback-manage ✅、07-feedback-detail ✅、announcement ✅(三页随 MSG-1 admin 腿建成 11-05,V10 全环 live 走查毕:反馈队列/快捷回复+详情对话/公告两步发布;窄化 R-09..R-13 在案) | **16** | 余面→M2-4/余组;另 report-policies 页为新建面(A4 开关承载,不在 41 页内) |
| 📐 底座 IAM 承接(**D-AY**:直接用骨架,不为 esports 再造/不改名) | audit-log、06-audit-detail、admin-user-manage、role-manage、permission-manage、menu-manage | 6 | system 既有面;操作级敏感清单对照(A8)留待 |
| 🌫 原型非本期 | attribution、channel-manage、landing-material、report-config | 4 | _nav.js 灰标(P5/梯①) |
| 🎯 M2 核心组 | — | **0** | 全员升已建桶(11-02 对账):member 两页随 M2-1(347a9ce7)、sku-manage/04-order-record/05-order-detail 随 M2-2(9860eb34)、activation-code+activation-code-detail 随 M2-3(45c276d4);到期提醒/批量续费/每日扫描按 M2-1 裁面不含 |
| 📋 E9 增量(D-AS⑥) | — | 0 | compliance-config 已随 E9 余刀建成升已建桶(11-01);执法消费面(注册年龄带/付费墙)随 PAY3/注册面 |
| 📋 D-AS 消息组 | — | 0 | 三页随 MSG-1 建成升已建桶(11-05,V10 闭环毕);窄化两裁决+R-08..R-13 随 W6 ljr(接洽单已立 11-06=pm-review-items PM-13..20);P21 已裁(D-BC 11-02:A1 服务端行级) |
| 📋 M2 余组(梯①) | dashboard、user-portrait、09-user-import、question-bank、question-detail、test-record、08-test-record-detail、test-type、game-category、payment-channel、rating-manage、system-config、app-version | 13 | 各项开工前先过对应 P 票:P20(dashboard/portrait 数据源)、P2(评级)、P24(归类)、P25(渠道面);app-version 已裁 B 建(D-BC 11-02),system-config=裁不建候选(config 珍稀原则) |
| 📋 M3/P4 | tob-tenant-list、tob-tenant-detail | 2 | 窗外 buffer;P14 已解散(11-02 D-BA),随 M3/P4 |

**计数:41 = 已建 16 + 底座 6 + 非本期 4 + M2 核心 0 + E9 增量 0 + 消息组 0 + 余组 13 + M3 2。**(11-05 对账吸收 MSG-1 三页;已建率 16/41≈39%;窗内可达成面 =已建 16+底座 6=22 页,余组 13+M3 2=15 页窗外,非本期 4 不计。)

## §三 旧码功能 × 处置(不在 85 页内者;出处=旧码穷举 2026-10-30)

| 旧功能 | 处置 | 依据/归属 |
|---|---|---|
| 测试排名 GET /test/rank + PK /test/pk + upload 补传 | 🗑 废弃 | D-AS④ 签认 |
| skill_prediction 款 | 🗑 裁出 | §0 对齐口径(multi_target 入) |
| 三档会员 ¥99/¥299/¥499 定价 | 🗑 作废 | D-AB/D-AC 取代(¥19.9/¥29.9) |
| sound-reaction 占位页 / BiometricMonitor / basic-results mock | 🗑 不迁 | demo/mock,原型未收 |
| Sentry 全量埋点(六族百事件) | ⏸ 远期 | B2;与 dashboard 漏斗的循环依赖→**P20** 解销 |
| 登录环境九件套(IP/省市/UA/设备解析+地域回填 Runner) | ⏸ | **P22**(隐私最小化 vs trace Tab 数据源) |
| ?ref= 推荐码明文+注册地域记账 | ⏸ 重批中 | A12 旧裁待重批→**P23** |
| 消息中心/反馈回复闭环/公告 | 📋 入窗 | D-AS①→MSG-1 |
| 手动解锁三开关+解锁过期任务 | 📋 改形入窗 | R7(过期=事实推导);运营三操作=M2-1 |
| 学校版报告(生成/7 天期/学科优势)+school 表 | 📋 读面=E5b 已建位 | P14 已解散(11-02 D-BA,生产无学校数据);D-AS⑤ 默认锁+手动解锁语义不变,M3 真租户再议 |
| App 版本管理+强更弹窗 | ⏸ | **P19** |
| Excel 导入导出族(用户导入/题库导入/激活码导出) | 📋 余组 | 题库/激活码导出随 M2-3+余组;用户导入=余组 |
| 题库三表 CRUD 23 端点(攻防/角色/克制) | 📋 余组 | E9 只读抽屉已建;CRUD/Excel 随余组(10-28 已注) |
| 双 Java 服务共享库/Flyway/Snowflake/JWT 细节 | 🗑 不迁 | clean-room 律(README §0.1) |

## §四 现状计数(2026-11-05 二盘点;对账基准)

- 新实现:esports=21 service/83 RPC/23 表/2 ticker(支付收敛 1min+对账 24h;MSG 四表+mockcashier 面已入);ctr=47 路由(+4=messages 树 2+feedback 树 2,MSG-1 ctr 腿)+13 引擎(前后端一一对应)+0 业务 store;admin=esports 面 14 页(13 计账+report-policies 账外)。
- 对照:旧 191 端点(admin 133)/36 表/ctr 64 视图/admin 23 视图。新版**反超项**=支付后端(旧仅枚举)/匹配引擎(旧纯设计)/全量服务端判分(旧 4 款作弊面)/账号合并(旧无)/报告快照冻结(旧无)/消息反馈线程域(旧单发回复)。
- 投入分布(esports/main 提交数):10-09 起全窗 ctr 140 : 后端 99 : admin 41;**10-30 校正后(至 11-05)ctr 42 : 后端 25 : admin 14**——admin 占比升至约 ctr 的 1/3,CP2/CP3 增 admin 构成项已见效。

## §五 对账环

- 本档为唯一处置真值:切片收口时对应行 ◐/❌→✅ 并注 hash;P 票裁毕即改 ⏸→📋/🗑。
- 与 gap-ledger 分工:本档=「建没建/归谁」;gap-ledger=「建得像不像」(逐字/态比对,六批=产品口径级偏离登记)。两档互不吸收。
- 对账节拍:2026-11-05 全档校准(PAY3 四片+Q 线收口吸收;§四 盘点刷新;原型仓 manage/design-systems git 零演进实证=10-30 四路穷举基线仍有效)。
