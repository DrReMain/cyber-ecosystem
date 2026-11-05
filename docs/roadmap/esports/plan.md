# esports 开发计划 — 60 天作战计划(2026-11-05 瘦身校准版)

> 归属:esports 前沿开发档。总则/架构定案/裁决台账见 README §0..§2;切片状态与认领记 progress.md;**需求分母与页面处置唯一真值=verify/demand-ledger.md**;已收片全叙事=progress-archive.md+提交 proven-by 尾。
> D0=2026-10-18;cutover 内部目标 **12-01~12-03(力争)/12-08(保底)**,承诺日 12-17=保险丝(D-BA);专版 buffer 至 12 月底-1 月中(D-K)。
> 执行模式(D-L)+设计对齐前置法(D-AQ)+团队分工法(D-AV):lzx 可认领任何切片;hzy 只被分配;未裁产品项停问(§2.2 P 票)。
> 规模:S(≤1 天)/M(2-4 天)/L(≈1 周)/XL。**已收池全量=archive(M1/E 线/H 线/质量批/M2 核心组/E9/R 线/Q 线/MIG1 五阶段/MSG-1 首刀/PAY3 四片);本档只载未完片,收口即剪枝迁档。**

## 实况(2026-11-05 校准,W3 中)

- **编码面超前原窗口表约两周**:M2 核心组+E9 全收(10-31..11-01);Q 线 Q1-Q6 读写两腿全毕(11-03..04,hzy 校对+lzx 队列①-⑩全清);MSG-1 契约+domain 首刀(11-04,W5 件拉前);PAY3 四片全收(11-04..05,末笔 2c3ecb7c 双推,W5-W6 件拉前)。
- MIG1 五阶段毕(guard/identity/entitlement/behavior/reconcile);snapshot+recompute 机械片冻结待 **P26/P26.1(ljr,硬限 11-24)**;预演 11-15(D-BE)。
- 需求总账(11-05 对账):ctr 44 页=✅26+◐5+❌8+⏸1+📄4(message 族三页+test-records+share-invite+scheme-b×3+version-update 未建);admin 41 页已建 16(+report-policies 账外;MSG-1 三页 11-05 收);唯一真值=demand-ledger。
- **等待面**:PAY2 行政件**已清空(11-05)**,真通道联调门解除;P26/P26.1=MIG1 冻结件门;G-402 openid 契约(lzx 建议裁②服务端富集,报备中,见 gap-ledger 五批);MSG-1 窄化两裁决=W6 ljr。
- hzy 车道清仓(Q 线全毕);下一分配=W4-W5 上线走查执行稿修订(cutover-watch 按实况)+值守。

## 窗口(11-05 校准重排;锚点裁决不动:cutover 12-01~03/12-08、CP3 核验 11-28、例会取消 D-BA)

| 窗 | 日期 | 服务/管理线(lzx) | 精修带(hzy,分配制) | 产品线(ljr) |
|---|---|---|---|---|
| W3 余 | 11.06-09 | MSG-1 ctr 腿已收(11-05,提交 5d3bded5);G-403/404 先行件亦收(11-05);**admin 三页+V10 闭环亦毕(11-05 拉前)** | — | 行政件已清空(11-05) |
| W4 | 11.10-16 | MSG-1 窄化两裁决接洽(R-01..R-13 清单,W6 ljr);PAY2 联调启动(行政件就绪后);**MIG1 预演 11-15**(D-BE:副本 11-14 就绪即演,绿门含回滚腿) | 上线走查执行稿修订(cutover-watch) | P22 否决窗 11-14;P26 理想 11-12 |
| W5 | 11.17-23 | 真通道联调收口(0.01 真钱闭环+全额退款;G-402/403/404 三红收);MIG1 snapshot+recompute(P26/P26.1 裁后)含曲线入码片 sizing ⑧ | 值守演练配合 | P23/P20 推进 |
| W6 | 11.24-30 | **CP3 三腿内部核验 11-28**(M1 e2e/PAY2 0.01 真钱/PAY3 全链+MSG-1 可走);cutover 准备(真实 appid 切换+上线清单) | 上线值守(执行稿=cutover-watch) | P23/P20 11-28;appid+绑域 11-26 |
| cutover | 12.01-03/12.08 | 停机切换+L4 验证+上线;旧栈停机封存 2 周(D-P);迁移执行=lzx 专职(D-BB) | 值守 | 上线内容核对 |
| 上线后首迭代 | 12 月起 | MSG-2(P23 已排)+M2 余组梯①(各项开工前过对应 P 票) | 精修池 | P4 11 月底(专版 buffer 窗) |

## 未完片(全部)

| 片 | 内容与验收要点 | 依赖 | 规模 |
|---|---|---|---|
| MSG-1 余面(lzx 全片 D-BB) | **ctr 腿(11-05,5d3bded5)+admin 三页 V10 闭环(11-05 拉前)全收**:反馈管理+反馈详情+公告管理三页,V10 全环 live 走查毕(五态全走/三 scope 投递 DB 侧证),走查捕五伤全修;余=窄化两裁决接洽(两裁决+R-01..R-13 清单=W6 ljr),裁决若翻契约面小改随改;banner ctr 消费面未建(遗留清单);定时发布=已裁非特性延后 | — | M(余 S) |
| PAY2 通道联调 | deploy 环境面切 wechat;0.01 真实闭环+全额退款(CP3 门);生产联调+对账;**微信外浏览器形态降级「引导微信内打开完成支付」(H5 递延 D-BF;Native 扫码已开通 11-05)**;上线前登录切真实公众号 appid(JSAPI 绑域,登记上线清单) | **行政件已清空(11-05:APIv3 已核对、H5 递延、Native 已开通、经营类目降级=有订单结算时再改)**;**APIv3 材料全齐并活证(11-05:openssl 三重互证+APP_ 前缀引号形 boot 绿,boot 对账 sweep 签名请求获微信真实应答 NO_STATEMENT_EXIST=昨日账单未出;env 两律实锤=裸名静默忽略假绿/数字形 string 须引号形;凭据=本机 ~/.xgwl-secrets/wechat-pay/ + wiki,不进仓)** | M |
| 真通道刀(G-402/403/404) | **G-403/404 先行件已收 11-05**(qrcode@1.5.4+码图 tile+sheet visual 槽;returnUrl 三层修形=客户端绝对收据 URL→biz.go 缝 order→wechat/mock 网关 redirect_url 镜像;e2e+浏览器网络面实捕;唯微信侧活证随真通道);**G-402 JSAPI openid:lzx 建议裁②服务端按绑定富集**(证据:JSAPI v3 openid 仅服务端统一下单入参,前端 invoke 只消费 prepay 参数,客户端零消费方;契约暴露=凭据无收益扩散),D-BC 报备形,真通道刀执行时落;另 clientIPOf 信任拨盘(XFF 优先 vs peer 优先)同刀裁=部署面口径 | G-402 裁决;G-403/404 无依赖 | S-M |
| MIG1 冻结件+预演 | 五阶段毕;snapshot 承载表待 **P26**、recompute 口径标+重放窗口待 **P26.1**(含曲线入码片 sizing ⑧:14 款仅 red_dot 冻结曲线出权威分,余 13 款 curvePending 诚实「未评级」=设计内非 bug,入码时点随本票一并钉);**预演 11-15**;cutover 迁移执行=lzx 专职;口径全档=migration.md | P26/P26.1(ljr 硬限 11-24) | M |
| M2-4 余 Tab | trace=**暂缓后置**(11-03 用户裁,归梯次,P22 否决窗 11-14 后复评);messages Tab=随 MSG-1 admin 面 | P22/MSG-1 | S |
| Q 线尾(役毕残件) | ⑧曲线入码 sizing(随 P26.1);consent 渠道列契约微欠(一句裁,P23 归因记账可并);E4 三配置模块=P4(ljr);隐私真文「7 天保留」句 vs 反推续用任意龄=ljr/法务车道 | P26.1/P4/P23 | S |

## M2 余组(梯①=上线后首迭代;每项开工前先过对应 P 票)

dashboard+user-portrait(**P20** 数据源)/评级管理(**P2** 扩面:档位映射+rating-manage+A10 门槛配置化 vs config 珍稀原则)/报告配置面/test-records 三态页+题库 CRUD/Excel/批量导入+test-type 档案面(上下架/介绍编辑契约增量)/09-user-import/game-category(**P24**)/payment-channel(**P25**)/app-version+ctr version-update 轻量提示(已裁 B,D-BC)/system-config(⏸ 裁不建候选)。

## M3 专版(承诺日外 buffer)与 M5 上线后(未排程)

- M3:P4(11 月底)→企业 App 授权登录→租户发码额度/功能差异配置(tob-tenant 两页);依赖 HL 文档。**不分叉铁律(D-AW)**:专版不得与公版代码分叉,对接差异一律落 capability provider/配置/模块注册表层(可插拔面最便宜层级优先);首租户预判=租户 app 内完成对接不穿透公版 H5。
- M5 微信小程序(D-AE):只服务公版;开工=栈设计评审先行(Taro 4.x 锚定;复用三档=games logic/paraglide/gen-connect-ts 真复用+transport/store/tailwind 缝适配,RAC/Router/cookie/SSR 不可用;首片=spike 连续族+离散族各一款真机)。待 ljr/甲方:appid 同主体否/账号互通否/首发范围。

---

## 检查点与降级梯(D-M;2026-11-02 D-BA 拉前版)

| 检查点 | 日期 | 通过标准 | 不过则 |
|---|---|---|---|
| CP1 | 11-08 | ✅ 10-25 提前 14 天实测达成(验签握手绿) | — |
| CP2 | 11-22 | ✅ 11-01 提前 21 天达成(M2 核心组必达三片+力争两片全收) | — |
| CP3 | **11-28(内部核验)** | M1 全链 e2e 绿;PAY2 0.01 真钱闭环;PAY3 付费墙全链可走(mock 已活证,真金腿待 PAY2);MSG-1 消息反馈可走(待余面) | 梯③(分批协商)+梯④(激活码兜底、支付热补);拉闸即执行不临场争论 |
