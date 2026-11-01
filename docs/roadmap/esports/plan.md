# esports 开发计划 — 60 天作战计划(2026-10-30 重构版)

> 归属:esports 前沿开发档。总则/架构定案/裁决台账见 README §0..§2;切片状态与认领记 progress.md;**需求分母与页面处置唯一真值=verify/demand-ledger.md**(2026-10-30 起,coverage/两矩阵状态列退役为语义锚)。
> D0=2026-10-18,承诺日=**12-17 上线切换**(范围含支付);专版 buffer 至 12 月底-1 月中(D-K)。
> 执行模式(D-L)+设计对齐前置法(D-AQ)+团队分工法(**D-AV**,2026-10-30):lzx 可认领任何切片(含 hzy 精修带);hzy 只被分配不自行认领;未裁产品项一律停问(§2.2 P 票=lrj/用户专属)。
> 规模:S(≤1 天)/M(2-4 天)/L(≈1 周)/XL。**已收池(M1 全量+H 线+PAY-C0/C1+PAY1+质量批)已压实迁 progress-archive.md(2026-10-30 重构役),本档只载未完片。**

## 实况(2026-10-30,W2 尾)

- M1 全片收满(E0-E10+E5b+E6c 全簇)+质量批三件收口(99c5bb06);CP1 提前 14 天实测达成,CP2 原前瞻项全实质达成。
- 支付:后端+真握手绿;PAY2 卡行政件(**用户即办**:经营类目补设+H5 支付权限+APIv3 密钥 32 位核对);PAY3 未启动,P7 视觉稿已落库(1c30b767)。
- 需求总账已建:**admin 原型 41 页仅 4 页建成**——M2 核心组入 CP2 门(本重构修正),防「门不含则永缓」。
- hzy:E-line-b ✅ 10-30(2b0bd350);R3 mt-F1 已落(64ac733c),余 P10 语言行。

## 窗口(W1-W4 已过并压实 archive;W5 起现行;并行线为建议默认,D-AV① 允许 lzx 跨带)

| 窗 | 日期 | 服务/管理线(lzx) | 精修带(hzy,分配制) | 产品线(ljr) |
|---|---|---|---|---|
| W5 | 11.16-11.22 | M2-1 会员终态+运营三操作→M2-2 SKU/订单面;PAY0 | R1 V4 视觉批→R2 ctr ◐ 面精修 | **P19/P21/P22 过票**;P4 推进(11 月底) |
| W6 | 11.23-11.29 | M2-3 激活码全面→M2-4 用户详情 Tab;E9 增量;MIG1 开工(**P14 前置**) | R3 两枚已裁小片→R4 | P14;P23 |
| W7 | 11.30-12.06 | PAY2 联调(行政件就绪后)+PAY3 合规执法+付费墙全链;MSG-1 消息反馈公告 | R5 admin 精修 | **P7 定稿(12-02)**;P6 推进 |
| W8 | 12.07-12.13 | MSG-2 分享裂变;MIG1 生产副本预演(P6);余组按梯①裁后执行 | 走查配合 | **P6(12-07)** |
| W9 | 12.14-12.17 | 停机 cutover+L4 验证+上线;旧栈停机封存 | 上线值守 | 上线内容核对 |

## M2 核心组(admin 运营面;CP2 门构成项;来源=demand-ledger §二)

| 片 | 内容与验收要点 | 依赖 | 规模 |
|---|---|---|---|
| M2-1 会员终态+运营三操作(首刀,10-30 用户裁) | Entitlement 加 status/terminated_at(**过期=时间推导,无状态列/无每日扫描,R7**);EntitleAdminService 增 Grant/Terminate/List/Get 四 RPC;system User 消息加手机号掩码投影;admin 会员列表(全部/有效/≤7 天到期/已过期 Tab+来源筛)/详情(有效期/来源/续订史 append-only,**续费=从到期顺延 31 天追加新行**)/开通-续费-终止 Modal+五语言;**admin 菜单重构+命名对齐(G-301/302,真值=verify/admin-nav-anchor.md:原型无「电竞」组,esports 自有组按原型落位,users 族正名用户列表/用户详情;系统与合规组=骨架直接承接 D-AY 不动)**。验收=ENT_E2E 扩展用例+chrome 走查(oracle=verify/admin-member-anchor.md)+审计行核验+菜单对齐;不做=到期提醒下发/批量续费/季卡年卡数据行 | E8,PAY1 | M-L |
| M2-2 SKU 多档+订单面 | admin SKU 管理(改价敏感入审计/上下架即时同步 C 端)+订单列表/详情/关闭(消费既有 OrderService 零契约增量)+解锁记录视图(去向/来源)。验收=PAY_E2E 兼走 admin 读回+O-1/2/3 断言 | PAY1,M2-1 | M |
| M2-3 激活码全面 | 契约增量:批次码表明细/停用(未核销即失效)/导出 CSV;admin 码表页(activation-code-detail:逐码核销人/时间+进度条≥95% 补码提示)。验收=ENT_E2E 扩展+导出走查 | E8 | M |
| M2-4 用户详情余 Tab | assessments=ListTestAttempts 消费(契约已在,零增量);membership=随 M2-1;trace=**P22 裁后**定数据源;messages=随 MSG-1。验收=四 Tab 真数据+O-1 断言形状 | M2-1,P22,MSG-1 | S-M |
| E9 增量(D-AS⑥) | 协议版本管理+同意记录查询(consent 契约缺口自 H2 挂账,本片闭:versions.ts 真源化 GET /agreements/current)+compliance-config(出生采集开关/付费年龄线/提示文案,敏感二次确认+入审计)。验收=admin 三页+ctr 协议版本行真值 | H2 | M |

## D-AS 召回组(W7-W8;未完走梯③)

| 片 | 内容与验收要点 | 依赖 | 规模 |
|---|---|---|---|
| MSG-1 消息中心+反馈+公告 | 契约=message/feedback 域新建(消息类型:系统公告/报告生成/会员说明;**未读契约先过 P21**);ctr 三页(message-center 双 Tab/feedback-submit 六类枚举/feedback-detail 对话式+追问回 pending);admin feedback-manage+07-detail+announcement(回复→C 端站内信闭环=旧语义 V10);未读红点。验收=双端走查+G-202 红条转绿(反馈 CTA 接真跳) | P21 | L |
| MSG-2 分享裂变 | ctr share-invite 页(免费层 5 维卡+动态三档文案+未成年代称+保存图片;D-AI② 分享动作即解锁预设);ref 验签与归因记账(**P23 裁后落**);邀请关系表。验收=分享链路走查+归因记账断言 | P23,D-AI | M |

## 支付线(未完片;家族定案四则与已收片 PAY-C0/C1/PAY1 压实 archive,法条仍在 README §2 与 kratos/第三方边界律)

| 片 | 内容与验收要点 | 依赖 | 规模 |
|---|---|---|---|
| PAY0 pay 模块拆分 | module/pay 双聚合根拆分(module/sku+module/order,跨模块读走 sku repo port;guard 豁免清空)。**触发器=下一片触碰 module/pay 的切片开工前**。验收=guard 零豁免+PAY_E2E 全链绿 | kratos §3 | S-M |
| PAY2 通道接入与联调 | deploy 环境面切 wechat;0.01 真实闭环+全额退款(CP3 门);生产联调+对账;H5 权限就绪后开 H5 scene;上线前登录切真实公众号 appid(JSAPI 绑域 openid 约束,登记上线清单);行政件就绪前代码面全 mock 不阻塞开发序 | 行政件(用户即办) | M |
| PAY3 合规执法+付费墙全链 | 合规:未成年拦截执法点(付费补填+18 岁禁付,满岁月粒度保守口径=次月起算)/仅人工退款入口/备案四落点。ctr 付费墙:**pay-confirm(三 SKU/会员期内隐单次只留续费/模拟收银台三态)+pay-result(四态)+age-required 闸挂接(逃生按钮)+member-detail 页**——P7 视觉稿(p7-member-paywall.html)=实现源,D-AU 四参数生效。验收=微信内/外双路真金或 mock 全链 e2e+未成年拦截反例 | PAY2,P7 | M-L |

## MIG1 迁移工具片(W6-W7 建,W8 预演;10-30 校准:功能面定型后开工,依赖绿≠现在建)

导入器落地 migration.md 全部口径:旧↔新 ID 映射持久化(覆盖软删行)、四档重算重放(A 明细/B 聚合/C 平移+legacy 标/D 作废)、身份翻译(密码 verify-then-rehash)、权益翻译(旧库无支付网关)、画像不完整态、报告快照保真搬运、种子重放、对账 SQL 族(行数+抽样逐字段+分布)。验收=migration.md 血泪清单逐条过+W8 生产副本脱敏预演对账全绿。**P14(school→租户时序)开工前必裁**;详见 migration.md;旧栈切换后停机封存 2 周(D-P)。

## hzy 精修带(分配制 D-AV;来源=gap-ledger 红条+coverage §七.5 B 批;lzx 可代领)

| 片 | 内容与验收 | 规模 |
|---|---|---|
| R1 V4 视觉值对齐批(**✅ 11-01 落 5c2f272a**:实况远小于 10-23 快照——五面仅剩 3 处旧词汇,均已像素等值换 rounded-control;agreement/switch/guide/pick 早已 V4;B批⑦ 真问题=user-detail 返回文案 M2-1 抄串「会员列表」,五语已正名「用户列表」,O-1 形状三详情页齐) | match 子页(pick/switch/guide)+agreement+age-required rounded 旧词汇→V4 词汇(B批⑥);admin 详情返回文案带目标名「返回 XXX 列表」(B批⑦,O-1 断言形状) | S |
| R2 ctr ◐ 面精修+gap 核对 | genre-pick 推荐角标(B批①);nickname emoji 显式拦截(B批②);profile-overview 模块集对核(B批⑤:训练趋势空态/报告中心内嵌/安全三行);home 续测卡 vs resume-remind 形态面检;每项收口=gap-ledger 对应条转绿或登记 | M |
| R3 已裁小片两枚(**已派**;mt-F1 ✅ 64ac733c) | 余 P10 语言设置行(D-AT:LocaleSelect 挂 profile 设置区);mt-F1 双出口文案键已落(D-AR:EndFace A 桩出口句+回读灰死钮,N=1 回执) | S |
| R4 上收两枚 | age.tsx 改消费 age-required/birth.ts(B批③);向导 games 步完成去向改直连 /match(B批④,**需 ljr 一句确认** home 中枢是否有意) | S |
| R5 admin 已建面交互文案精修 | O-1/2/3 断言形状复扫(admin 对齐批,progress 旧注「待排」)+五页空态/确认语对原型 | M |

## M2 余组(梯①=上线后首迭代;**每项开工前先过对应 P 票**)

dashboard+user-portrait(**P20** 数据源)/评级管理(**P2** 扩面:档位映射+rating-manage 页+A10 门槛配置化 vs config 珍稀原则)/报告配置面(梯①)/测试记录明细+题库 CRUD/Excel/批量导入+test-type 档案面(上下架/介绍编辑契约增量)/09-user-import/game-category(**P24**)/payment-channel(**P25**)/app-version(**P19**)/system-config(⏸ 裁不建候选)。

## M3 专版(承诺日外 buffer)与 M5 上线后(未排程)

- M3:P4 裁决(11 月底前)→企业 App 授权登录→租户发码额度/功能差异配置(tob-tenant 两页);依赖 HL 文档。**不分叉铁律(D-AW)**:专版不得与公版代码分叉,对接差异一律落 capability provider/配置/模块注册表层(可插拔面最便宜层级优先);物理隔离递延 delivery-time fork 最晚点;首租户预判=租户 app 内完成对接不穿透公版 H5。
- M5 微信小程序端(D-AE):只服务公版;开工=栈设计评审先行(Taro 4.x 锚定;复用三档=games logic/paraglide/gen-connect-ts 真复用+transport/store/tailwind 缝适配,RAC/Router/cookie/SSR 不可用;连续族 canvas=帧率保真下限;平台缝三件=登录 code2session+Bearer 会话+支付绑 appid;首片=spike 一款连续族+一款离散族真机)。待 ljr/甲方:appid 同主体否/账号互通否/首发全量或子集。

---

## 检查点与降级梯(D-M;**2026-10-30 修订:门含 admin**,依据=demand-ledger admin 4/41)

| 检查点 | 日期 | 通过标准 | 不过则 |
|---|---|---|---|
| CP1 | 11-08 | ✅ 10-25 提前 14 天实测达成(验签握手绿) | — |
| CP2 | 11-22 | 原项(游戏≥8/E4/E8/E5 框架/PAY1)已实质达成;**新增构成:M2 核心组 admin 面可用(M2-1..M2-3 必达,M2-4/E9 增量力争)** | M2-4/E9 增量移 W7;核心三片不足走梯③ |
| CP3 | 12-02 | M1 全链 e2e 绿;PAY2 0.01 真钱闭环;**PAY3 付费墙全链可走(ctr 下单→支付→解锁→报告)**;MSG-1 消息反馈可走 | 梯③(分批协商)+梯④(激活码兜底、支付热补) |

检查点由三人 10 分钟站会核对,结果记 progress.md;拉闸即执行,不临场争论(D-M)。
