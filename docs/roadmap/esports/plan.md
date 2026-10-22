# esports 开发计划 — 60 天作战计划(大框架定稿)

> 归属:esports 前沿开发档。总则/架构定案/裁决台账见 README;切片状态与认领记 progress.md。
> **D0 = 2026-10-18,承诺日 = 12-17 新系统上线切换(范围含支付)**;专版 buffer 至 12 月底-1 月中(D-K)。
> 执行模式(D-L):AI 并行作战——细切片、契约先行、双线全并行、每日集成;工程师职责 = 指挥 coding agent + 审查 + 集成;完成门禁与协作规约见 README §5。
> 规模:S(≤1 天)/ M(2-4 天)/ L(≈1 周)/ XL(多人多批)。

---

## 时间轴总览(检查点与降级梯见文末)

> **实况(2026-10-23,W1 末)**:W1-W2 已超额——E6b 完整形态、E0 统一提交面、E2、E3 前半(Tier A 6/10)提前落;H1/H2/H-A/H-B 全交付(mock 净拔除,读面已真);P1/P3/P9 提前定案。进行中:E3 余款、H-C(已分配)。**风险:支付线零启动——CP1(11-08)门含「支付沙箱通」。**逐片状态见 progress.md 认领表。

| 窗口 | 日期 | 服务线(建议用户) | 客户端线(建议 hzy) | 产品线(ljr) |
|---|---|---|---|---|
| W1 | 10.19-10.25 | E6b 登录重构(裁决落地:sms/绑定/合并)——骨架演进后最小闭环按 esports 设计重做 | hzy 外围启动线开工:H1 系统态件(onboarding 首片)→H2 协议页→H-A 引擎框架启动 | P1 分析启动(✅ 10-21 定案);支付商户号/凭据打包确认(✅ 10-21,见 prd.md 就绪记录;遗留:经营类目补设、H5 权限申请) |
| W2 | 10.26-11.01 | E6b 重构收口 + E0 余面推进(统一提交/报告读面/entitlement;H-B rawDetail 反哺) + E2 重启(接通闭环已验,备份重放) | H-A 收口(mock config 下 red_dot 可玩+时钟正确性)→H-B 反应瞄准批起步 | **P1 定案 ✅ 10-21(D-AB/D-AC)** |
| W3 | 11.02-11.08 | E3 评分引擎(框架+Tier A 前半) | E7 游戏引擎框架 + 反应瞄准批起步 | P3 完成页/付费墙交互项 |
| W4 | 11.09-11.15 | E3 收口(13 款 Scorer+档位层)、E4 匹配 | E7 反应瞄准批收口(4 款)、E9 最小 admin 面 | — |
| W5 | 11.16-11.22 | E8 权益、PAY1 订单域 | E7 追踪技能批(4 款)、E5 报告前端 | P7 会员态付费墙稿启动 |
| W6 | 11.23-11.29 | E5 报告域收口、M2 核心组(会员/激活码全面) | E7 决策资源批(5 款)、M2 后台页面 | **P4 ToB 裁决(11 月底前)** |
| W7 | 11.30-12.06 | PAY2 微信支付接入+联调(沙箱→生产)、PAY3 合规执法 | E10 全链 e2e、M2 核心组页面 | **P7 定稿(12-02 前)** |
| W8 | 12.07-12.13 | M2 余组(可降级梯①)、迁移预演(生产副本+nginx 仿真) | 迁移演练配合、缺陷修复 | 上线验收清单走查 |
| W9 | 12.14-12.17 | **停机 cutover + L4 验证 + 上线**;旧栈停机封存 | 上线值守 | 上线内容核对 |

并行线仅为建议默认(认领制不变,首周例会正式认领);E0 契约冻结前 E3/E5/E6b 只做设计稿(E6a 纯壳零 RPC,不受此限,D-R;H 线同待遇——只消费已冻结读面,mock 落 transport 缝,禁自造类型);proto+gen 为共享敏感面。

---

## M1 公版 C 端全链闭环(目标窗口见上表)

| 片 | 内容与验收要点(语义级验收,D-O) | 依赖 | 规模 |
|---|---|---|---|
| ✅ E0 契约首刀 | esports v1 proto:测试域(类型/配置读面、统一提交(建议单 RPC + oneof 明细)、明细/成绩上报、排行)、画像、匹配、单次报告读面、entitlement 最小面(scope_kind/scope_ref,付费只发 trial;D-AB)。nullable 语义显式(optional),判定阈值随配置下发。验收:proto 全链生成+lint+breaking 绿,评审冻结(D-U:增量推进——种子片=画像+游戏目录已于 2026-10-18 冻结,余面逐片评审) | README §0 | M |
| ✅ E1 服务骨架 | kratos-new-service 复制集:双面 accessLadder(ADMIN/MEMBER)、catalog 注册、deploy compose 面、system 消费(登录/会员链)。验收:服务起活,会员链探针 e2e 绿(**兼作 hzy onboarding 切片**) | E0 | M |
| ✅ E2 schema+种子 | ent 首批:test/game/preference/config/contribution/题库/test_record/ability_score/entitlement 事件;种子重放(修坑:dual_zone 枚举统一、multi_target 补行、skill_prediction 不入、枚举约束防回归)。验收:migrate 归一+种子 e2e。2026-10-20 缩围补登(D-AA:entitlement→E8、contribution→E3;multi_target 补行落 D-Z) | E1 | M |
| ⏳ E3 评分引擎 | 三层落位:指标层(13 款指标定义单源档,本片创建)/门槛层(百分位档位)/评分层(无上限线性+权威分位双轨);Tier A 九款全量重算 + Tier B 四款闸门;red_dot 锚点初始校准常量;ScorerRegistry 插件契约(加测试零改主干)。验收:域函数单测+提交→重算→报告分 e2e | E0,E2 | XL |
| E4 匹配引擎 | match_config(预留 scenario 维度)+类型轴多数投票+两层编排+当日 seed+session 锁死落库+重测 3/1。验收:同日稳定/跨日变化/重排换 seed e2e | E2 | M |
| E5 报告框架+能力报告 | 板块插件化(supports)+缺维度不渲染七硬约束+完整度指示+同龄百分位(Z-score CDF+置信分层)+档位四色;M1 只实现单次能力报告所需板块。验收:缺维场景渲染收拢 e2e | E3 | L |
| ✅ E6a ctr 纯壳 | 零 RPC:TanStack Start 新工程(Nx 目标面)+ react-aria/tailwind 接线(D-R)+ token 层(旧仓 DESIGN.md 为经验数据输入,按 D-O 合理实施;D-S dark-only 单套)+ 路由骨架(仅根路由+空白落地页,D-T:能力配置为交付面) + 微信 WebView 适配基线(safe-area/键盘 vh/无 hover 假设/横竖屏地基)+ shared-ts UI-free 四包(error/cookie/store/router-progress)接线;i18n=paraglide zh-CN+en-US(baseLocale zh-CN);**SEO 全量面(2026-10-18 补入,C 端公网)**:robots.txt + sitemap(hreflang/x-default)+ 流程页 noindex + 逐页 title。验收:dev/build/typecheck/lint 全绿,dark-only 语义 token 就位 | — | M |
| ✅ E6b 登录+画像流 | 微信/手机登录(消费 system 既有契约,不等 E0)、画像流(昵称必填/出生年月 8 档/未满 13 监护人硬闸/游戏偏好) | E6a,E0(画像流半片) | L |
| E7 游戏框架+13 款 | 统一引擎框架(rAF+dt+refs 或 canvas,墙钟计时+暂停顺延单实现)、遥测采集(新契约)、参数化单流程路由、配置驱动运行时、真响应式横竖屏。三批:反应瞄准(red_dot/flash_aiming/multi_target/target_recognition)W3-W4;追踪技能(tracking_moving/moving_target/skill_trajectory/dual_zone)W5;决策资源(number_stars/offense_defense/character_counter/resource_match/economic_sim)W6。验收:每批 live 冒烟+遥测字段对齐(语义级) | E6,E3 | XL |
| E8 权益最小面 | entitlement 事件模型+激活码全生命周期(生成/批量/兑换原子守卫防超卖)+报告访问判定(policy open/membership/tenant/closed + scope 覆盖;D-AB)。验收:兑换并发不超卖 e2e | E2 | M |
| E9 最小 admin 面 | admin client 挂 esports 模块:激活码+测试项/配置管理+报告访问策略(职业/认知 open/membership/closed)+题库种子查看;复用既有表格/搜索/动作列范式与 O-1/2/3 裁定。验收:开码→兑换→解锁闭环 e2e | E0 | M |
| E10 全链 e2e | 登录→画像→匹配→4 款→完成→单次报告→激活码解锁全链 live e2e;原型断言按语义级移植为验收基线(D-O);docs:check | 全部 | M |

## hzy 外围启动线(2026-10-19 增补;W1-W2 桥接,定义归此、状态记 progress)

登录重构与 E0 余面为承重面留 lzx;hzy 切片圈定 ctr 非登录区,文件面与 lzx 线不相交(合流点仅 routeTree.gen 与未来一次提交适配器)。边界四律:①config 消费用 gen 真类型(已冻结 GetTest/GetTestConfig),mock 只落 transport 缝;②游戏运行时禁 import auth/services(插件契约准入);③不烤权威指标——客户端采集 rawDetail,评分归服务端(旧架构同构);④提交缝后置,E0 冻结后薄适配器接上,游戏零改动。任务分配判据(D-AJ,2026-10-23):不涉及骨架/底座(shared-ts/shared-go、proto 契约面、system、deploy)的任务可分配 hzy——不再限于本节圈定的 ctr 非登录区。

| 片 | 内容与验收要点 | 依赖 | 规模 |
|---|---|---|---|
| ✅ H1 ctr 系统态件 | error/empty/loading/退出确认+骨架屏+氛围布局(旧版 UserFlowAmbientLayout 语义级);兼 hzy onboarding 首片。验收:三链绿+走查 | E6a | S |
| ✅ H2 协议页+版本化承载 | agreement/age-required 页面+路由+版本参数;文案占位等 ljr 定稿;URL 形状开工前与 lzx 定约。验收:三链绿+走查 | H1 | S |
| ✅ H-A 引擎框架片 | E7 框架前拉:GameClock 单实现(墙钟+暂停顺延,fake timers 可测;旧版 13 处手搓计时收敛为 1)+参数化路由 /test/:code+GameModule 插件契约+overlay 族。验收:mock config 下 red_dot 可玩+时钟正确性测试 | H1 | M |
| ✅ H-B 反应瞄准批 | E7 首批前拉:red_dot/flash_aiming/multi_target/target_recognition;harness 免登录;rawDetail 采集反哺 E0 统一提交契约;玩法语义级对照旧版(D-O)。验收:每款 harness 可玩+明细字段与旧版 per-game SubmitData 语义对齐+mock 插桩净拔除(test-mock.ts 与其测试删除、transport.ts/env.ts 插桩还原、VITE_TEST_MOCK 零残留;拔除前插桩点全带 TODO(mock-removal) 标记——2026-10-20 用户裁定:临时插桩不常驻且必须 TODO 可寻) | H-A | M-L |
| ▶ H-C 判定回路闭环(已分配 10-23) | 边界四律④预定形态兑现:提交缝薄适配器(TestRunHost 终局采集——明细流+契约三计数 misses/missed_high/spawned_enemy+后台段;attempt_no E4 前恒 1,序数真律归 E4 会话)→SubmitTestAttempt 单发,游戏零改动;EndFace 增服务端判分回执卡(原始分/权威分/档位/valid/无效原因/retest_left,无效原因词表双语文案,权威分曲线待校准款按回执降级渲染,本地明细卡降为辅助;重试钮维持本地练习语义,真律归 E4)。验收=真人全链:登录→/match→一局 red_dot→服务端分数回执渲染+会话失效反例——esports 首条人工可 e2e 调试的完整测评闭环(mock 已于 H-B 拔净、读面已真,本片仅余判定回路) | H-B,E3-c | M |

## 支付片(D-B2:纳入承诺窗口,保持可插拔)

| 片 | 内容与验收 | 依赖 | 规模 |
|---|---|---|---|
| PAY1 订单域 | 订单/幂等(同单复用+结算幂等)/状态机;订单行带 scope_kind/scope_ref(单次=trial,会员=period);SKU 读面(原价/现价结构,D-F 无立减展示) | **P1 已定(D-AB/D-AC,10-21)** | M |
| PAY2 微信支付接入 | JSAPI(微信内)+H5(独立网页)统一下单/回调验签/对账;沙箱全流程→生产联调;**商户号 W1 即确认** | PAY1 | L |
| PAY3 合规执法+付费墙 | 未成年拦截执法点(付费补填+18 岁禁付)/仅人工退款入口/会员态付费墙(P7 视觉稿,期内隐藏单次只留续费) | PAY2,P7(12-02) | M |

## M2 运营与变现(核心组 W5-W7;余组 W8 可降级梯①)

- **核心组(承诺窗口内)**:会员/权益终态(按 D-AB)+SKU 多档;激活码全面(批次/导出/核销视图);订单+解锁记录;用户列表+档案核心 Tab。
- **余组(梯①候选,上线后首迭代)**:评级管理(P2)/增长裂变(分享卡+ref HMAC+归因)/消息中心/反馈客服/ToB 租户视图/仪表盘/报告配置面。

## M3 专版(承诺日外 buffer:12 月底-1 月中)

P4 裁决(11 月底前)→企业 App 授权登录→租户发码额度/功能差异配置;依赖 HL 文档(W1 跟进确认到位时点)。

## M4 存量迁移 cutover(预演 W8,切换 W9)

详见 migration.md;P6 演练日程确认(12-07 前);旧栈切换后停机封存 2 周(可回滚)期满下线归档(D-P)。

## M5 上线后候选(未排程,仅登记;开工时补切片定义)

- **微信小程序端**(D-AE,2026-10-21 甲方需求):C 端第二表面,业务与 ctr H5 相同;只服务公版,不做专版。前置=ctr H5 完成(12-17 上线);开工=先小程序栈设计评审(栈区 CONVENTIONS 随首批真实代码立档,AGENTS §7)再移植业务。
- **小程序栈选型初步讨论(2026-10-21,评审锚点非定案;正式定案走 D-AE 预埋的栈设计评审)**:
  - **Taro 4.x 锚定**:React+TS 跨端路线无对手(uni-app React 二等/Remax 停维/原生 WXML 弃资产);5.x 未开源不押注,开源后作升级窗重评。最大版本风险=React 19 未官宣支持(ctr 现挂 19.3+Compiler,spike 首验;降 18 估可控)。
  - **复用三档**(证据等级:旧仓仅方向性,新栈全量重验):①真复用=13 款 games logic.ts(framework-free)+paraglide messages+gen-connect-ts+shared-error;②缝适配=transport(Taro.request 注入 fetch)+react-query v5(接管 focus/online manager)+store 持久化(Taro.setStorage)+Tailwind(weapp-tailwindcss,v4 支持待验——旧仓验的是 WindiCSS v3 语法族,不迁移);③不可用=RAC/TanStack Router/cookie/SSR/Sentry-react,组件库候选 NutUI 当前线。
  - **渲染分族**:ctr 引擎=手撕 DOM 非 canvas(red-dot 直摸 createElement);离散族(反应瞄准/决策资源)→Taro View 声明式(React diff 天然合规「禁动态标签」);连续族(追踪技能)→必须 canvas,**帧率=测量保真下限非体验项**(测试原始数据即产品输出)。渲染缝(共享指令层 vs 每端独立 renderer)=评审头号议题;上线前不动 H5 引擎(D-Q)。
  - **平台缝三件(比 UI 重)**:登录(wx.login code2session;openid≠公众号 openid,互通需 UnionID 同主体+system 新 provider 面)、会话(无 cookie→Bearer,2026-09-13 定案已预留,核 system 发 token 契约)、支付(商户号绑小程序 appid)+request 域名白名单+分包(13 款按 subpackage)。
  - **webview 内嵌 H5 撤案**:审核打击空壳 webview,整体划掉不留兜底(用户裁定)。
  - **节奏(交付优先+骨架一等公民 D-AF)**:开工首片=spike(壳+transport 适配+一款连续族 canvas 真机+一款离散族对照);参考仓 cyber-mini/cyber-mini-staff(Taro 3.6 系)定性只读范式参考,同 lab/esports 地位。**待 ljr/甲方**:appid/主体是否与公众号同主体、H5/小程序账号是否互通、首发 13 款全量还是子集(子集首发对齐降级梯② hot-add 形态)。

---

## 检查点与降级梯(D-M,预签执行)

| 检查点 | 日期 | 通过标准 | 不过则 |
|---|---|---|---|
| CP1 | 11-08 | E0/E1/E2 完成;E3 过半;E6a/E6b 完成(登录+画像可跑);商户号确认+支付沙箱通 | 拉梯①(M2 余组移上线后) |
| CP2 | 11-22 | 游戏 ≥8 款(两批);E4/E8 完;E5 框架通;PAY1 完 | 拉梯②(游戏余款上线后两周 hot-add) |
| CP3 | 12-02 | M1 全链 e2e 绿;PAY2 沙箱全流程;M2 核心组可用 | 拉梯③(分批交付协商)+梯④(支付受阻激活码兜底、支付热补) |

检查点由三人 10 分钟站会核对,结果记 progress.md;拉闸即执行,不临场争论(D-M)。
