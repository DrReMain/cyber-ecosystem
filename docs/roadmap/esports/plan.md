# esports 开发计划 — 60 天作战计划(2026-11-07 重构役三版)

> 归属:esports 前沿开发档。总则/架构定案/裁决台账见 README §0..§2;切片状态与认领记 progress.md;**需求分母与页面处置唯一真值=verify/demand-ledger.md**;覆盖映射与缺口=verify/coverage-map.md;已收片全叙事=progress-archive.md+提交 proven-by 尾。
> D0=2026-10-18;cutover 内部目标 **12-01~12-03(力争)/12-08(保底)**,承诺日 12-17=保险丝(D-BA);专版 buffer 至 12 月底-1 月中(D-K)。
> 执行模式(D-L)+设计对齐前置法(D-AQ)+团队分工法(D-AV):lzx 可认领任何切片;hzy 只被分配;未裁产品项停问(README §2.2 P 票)。
> 规模:S(≤1 天)/M(2-4 天)/L(≈1 周)/XL。本档只载未完片与未修缺陷;收口即剪枝迁 archive。

## 实况(2026-11-07,W3 末)

- 已收池:M2 核心组+E9(10-31..11-01)/Q 线 Q1-Q6(11-03..04)/PAY3 四片(11-04..05)/MSG-1 契约+ctr 腿+admin 三页+V10 闭环(11-04..05)/PAY2 Native 真通道+0.01 真钱+全额退款腿(11-05..06)/MIG1 五阶段+备料+回滚彩排(11-06)/走查 R1 轮(X1-1、X1-2、X2、X3、A 套 15 例、C1 六例)/G 伤单清偿 A-F 批+三路对抗审(11-06..07)。叙事全文=progress-archive+git proven-by。
- 需求分母(demand-ledger,2026-11-07 机核):ctr 44=✅29+◐5+❌5+⏸1+📄4;admin 41=已建 16+底座 6+非本期 4+余组 13+M3 2;新实现=21 service/85 RPC/23 ent 表/ctr 50 路由文件(1 root+9 布局壳+40 叶页)/admin esports 17 面。
- 等待面:**P26/P26.1(ljr,理想 11-12/硬限 11-24)=MIG1 冻结件门**;P22 否决窗 11-14;D-MIG-1/2=预演前决策;微信侧活证随 PAY2 联调;U2/U3=UI 校准余段。
- 覆盖核验(2026-11-07,八域三方对照=verify/coverage-map.md):56 例对已建面/RPC 的映射成立;缺口集中=**MOBA 侧零执行(8/13 引擎无任何用例跑过)**、admin 三个写 RPC 全链零执行、C2/C3/C4/C6 例族未走;新发现偏离/缺陷已分流入本档缺陷工单与 prd 待裁区。

## 窗口(11-05 校准重排;锚点裁决不动:cutover 12-01~03/12-08、CP3 核验 11-28)

| 窗 | 日期 | 服务/管理线(lzx) | 精修带(hzy,分配制) | 产品线(ljr) |
|---|---|---|---|---|
| W4 | 11.10-16 | MSG-1 窄化裁决窗(PM-13..20 接洽单=prd 待裁区);PAY2 联调启动;**MIG1 预演 11-15**(先过 D-MIG-1/2,绿门含回滚腿) | 上线走查执行稿修订(cutover-watch) | P22 否决窗 11-14;P26 理想 11-12 |
| W5 | 11.17-23 | 真通道联调收口(JSAPI 真实 appid+绑域+微信侧活证+微信外降级面);MIG1 snapshot+recompute(P26/P26.1 裁后)含曲线入码 sizing ⑧ | 值守演练配合 | P23/P20 推进 |
| W6 | 11.24-30 | **CP3 三腿内部核验 11-28**(M1 e2e/PAY2 0.01 真钱/PAY3 全链+MSG-1 可走);cutover 准备(真实 appid 切换+上线清单);U2/U3 若未毕随窗收 | 上线值守(执行稿=cutover-watch) | P23/P20 11-28;appid+绑域 11-26 |
| cutover | 12.01-03/12.08 | 停机切换+L4 验证+上线;旧栈停机封存 2 周(D-P);迁移执行=lzx 专职(D-BB) | 值守 | 上线内容核对 |
| 上线后首迭代 | 12 月起 | MSG-2(P23 已排)+M2 余组梯①(各项开工前过对应 P 票) | 精修池 | P4 11 月底(专版 buffer 窗) |

## 未完片(全部)

| 片 | 余面与验收要点 | 依赖 | 规模 |
|---|---|---|---|
| MSG-1 余面(lzx,D-BB 全片) | 窄化裁决落地(PM-13..20=prd 待裁区;全维持=零改动,裁翻契约面小改随改);banner ctr 消费面未建(随 MSG-2/精修带评估);定时发布=已裁非特性 | W6 ljr 裁决 | S |
| PAY2 通道联调(lzx) | 已毕=行政件全清+Native 腿全生命周期+0.01 真钱闭环+全额退款腿+NO_STATEMENT_EXIST 修(11-05..06)。余=JSAPI 腿真实 appid+绑定域(上线清单);微信侧活证(G-402/403/404 尾);微信外降级「引导微信内打开」(D-BF);XFF 信任拨盘台账收口;query 无触发器(convergeStaleAfter=10m)与 boot 放弃日 ERROR 两裁点随刀 | W4 起 | M |
| MIG1 冻结件+预演(lzx 专职,D-BB) | 已毕=五阶段+生产副本直载+谱系三源互证+census+回滚腿彩排+密码面实数定谳(11-06)。余=snapshot 承载表(P26)+recompute 口径标与重放窗口(P26.1,含曲线入码 sizing ⑧);**预演 11-15**(D-BE 绿门含回滚腿);预演前过 D-MIG-1(school 域处置)/D-MIG-2(账号冲突消解);cutover 迁移执行同专职 | P26/P26.1+D-MIG-1/2 | M |
| U2 ctr 原型对齐(harness 重激活) | 渲染可达已毕(U1);余=按 HARNESS-SOP 重激活保真 harness 对原型核视觉;**B-6(done 页三差)修法在此**;趋势进阶 CTA 偏离处置随本片(见 prd 待裁) | 走查轮前 | M |
| U3 admin 对齐基座范式 | 三裁定落地=列表页 tab 筛选族迁基座 Filter/Segmented 例外形;容器层级与留白对齐骨架基线;drawer 内简单列表降级朴素组件;ActionIconButton href locale 前缀族四实例同修 | — | S-M |
| M2-4 trace Tab | 暂缓后置(11-03 用户裁;P22 否决窗 11-14 后复评) | P22 | S |
| Q 线尾(役毕残件) | ⑧曲线入码 sizing(随 P26.1);consent 渠道列契约微欠(一句裁,可并 P23);E4 三配置模块=P4;隐私真文「7 天保留」句 vs 反推续用任意龄=ljr/法务车道 | P26.1/P4/P23 | S |

## 缺陷工单(未修;取证全文=git 391ccc87 gap-ledger 或 verify/coverage-map.md 缺口节)

| # | 缺陷 | 级 | 依据 | 归属 |
|---|---|---|---|---|
| B-1 | 支付到期竞速「已付」恢复死路:wechat classifyAPIError 将全部 HTTP 400 折为 ErrInvalidArgument,retirePending 只认 ChannelError→2h 边缘已付单卡 pending(钱已收/无权益);dev/mock 结构性照不见 | 最重 | G-420 F1(error.go:41-45/biz.go:1154,亲手复核) | PAY2 真通道刀共裁 |
| B-2 | 兑换幂等读在事务外:同族两码并发兑换对已失效 grant 双双落穿=静默烧一码 | 高 | G-421①(entitle/biz.go:421-435,亲手复核) | cutover 前修或挂险由用户裁 |
| B-3 | Terminate×竞速授予复活窗:Grant tx 内窗读与插入间 Terminate 可先提交→终止成功而新行自旧到期复活 | 高 | G-421②(entitle/biz.go:173,亲手复核) | 同上 |
| B-4 | esports(及 agent)全部 WRITE 面零审计行:kratos §5a 三兄弟缺一+proto/UI「入审计」承诺落空;修形=esports 生产者三件(MQ+Sink+ladder 双块,镜像 system server.go:38-58) | 高(cutover 审计合规) | G-417;消费侧已在(system/audit 订阅 audit_log 实锤) | lzx+用户裁定后落片;**并行会话已在途,落片时核对** |
| B-5 | pay 墙过期会员灰 banner 无生产者:MembershipWindow GT-now 恒滤→expired 态三死面;两形待裁(墙补读 digest LastMembershipUntil / 删死面三件) | 中 | G-416(SQL 造态活跑) | 随下片裁 |
| B-6 | session-done 页三差:已测收据行=评级口径非原型已玩口径;CTA 族反转(报告入口为主,原型明令不在本页);blob 氛围缺失 | 中 | G-410(用户裁定「视觉低质」) | U2 |
| B-7 | G-420 F4-F6 待复核定级:同单复用忽略渠道切换/对账游标内存态重启归昨/sku 丢 PeriodDays 结算零授予仅 ERROR | 待核 | G-420(代理引证) | 下片复核 |
| B-8 | pay result fail 卡「我的→订单」死指路(ListOrders 零客户端消费方,订单列表页原型未画) | 低 | coverage D5(pay/zh-CN.json:54;order/service.go:75) | 随 PAY2/C3-3 走查裁 |
| B-9 | entitlement.proto 幂等注文滞后:注文 per-(member,code) vs 实现 per-(member,family) 族宽形 | 低 | coverage D6-5(entitlement.proto:13-15;X3-4 宽形活证) | 契约注释一行修 |
| B-10 | **「手机号+密码」登录假承诺**:ctr set/change 密码页承诺「设置后即可用手机号+密码登录」(security_pw_set_pitch+总面「已设置——可用于账号密码登录」),但 system Login 只按 user.account 句柄查(exchangePassword→FindByAccount,无手机号→credential 解析分支);SMS-born 用户 account=ULID 句柄(dbn18mlthm88rvujd2t0 实证),手机号登录必 400 SYSTEM_LOGIN_FAILED「account not found」→**全体自然注册 member 的 set/change 密码=死出口**。底座 schema 注文明言「phone are bindable credentials, never the identity key」=esports 侧消费错位非底座 bug | **高** | R2 走查 11-07 三重证据:UI 承诺(security/password.tsx:148-149+zh-CN.json)+服务端行为(Login 400 日志)+数据模型(system user 表实查);B-8 邻案 | 修复方向待裁(三选):a) system Login 增手机号解析分支(底座动,重) b) ctr 文案撤回承诺(轻,产品语义损失) c) ljr 另裁 |

走查覆盖缺口(非缺陷,未执行的用例腿)=verify/coverage-map.md 缺口清单,后续多轮走查轮次消费。

## M2 余组(梯①=上线后首迭代;每项开工前先过对应 P 票)

dashboard+user-portrait(**P20** 数据源)/评级管理(**P2** 扩面:档位映射+rating-manage+A10 门槛配置化 vs config 珍稀原则)/报告配置面/test-records 三态页+题库 CRUD/Excel/批量导入+test-type 档案面(上下架/介绍编辑契约增量)/09-user-import/game-category(**P24**)/payment-channel(**P25**)/app-version+ctr version-update 轻量提示(已裁 B,D-BC)/system-config(⏸ 裁不建候选)。

## M3 专版(承诺日外 buffer)与 M5 上线后(未排程)

- M3:P4(11 月底)→企业 App 授权登录→租户发码额度/功能差异配置(tob-tenant 两页);依赖 HL 文档。**不分叉铁律(D-AW)**:专版不得与公版代码分叉,对接差异一律落 capability provider/配置/模块注册表层;首租户预判=租户 app 内完成对接不穿透公版 H5。
- M5 微信小程序(D-AE):只服务公版;开工=栈设计评审先行(Taro 4.x 锚定;复用三档=games logic/paraglide/gen-connect-ts 真复用+transport/store/tailwind 缝适配,RAC/Router/cookie/SSR 不可用;首片=spike 连续族+离散族各一款真机)。待 ljr/甲方:appid 同主体否/账号互通否/首发范围。

---

## 检查点与降级梯(D-M;2026-11-02 D-BA 拉前版)

| 检查点 | 日期 | 通过标准 | 不过则 |
|---|---|---|---|
| CP1 | 11-08 | ✅ 10-25 提前 14 天实测达成(验签握手绿;门经 10-25 校准=GET /v3/certificates,APIv3 无沙箱) | — |
| CP2 | 11-22 | ✅ 11-01 提前 21 天达成(M2 核心组必达三片+力争两片全收) | — |
| CP3 | **11-28(内部核验)** | M1 全链 e2e 绿;PAY2 0.01 真钱闭环(JSAPI/Native 腿;H5 递延 D-BF=新版上线后重审补活证);PAY3 付费墙全链可走(mock 已活证,真金腿待 PAY2);MSG-1 消息反馈可走(待余面) | 梯③(分批协商)+梯④(激活码兜底、支付热补);拉闸即执行不临场争论 |
