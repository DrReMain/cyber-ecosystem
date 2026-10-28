# esports 开发计划 — 60 天作战计划(大框架定稿)

> 归属:esports 前沿开发档。总则/架构定案/裁决台账见 README;切片状态与认领记 progress.md。
> **D0 = 2026-10-18,承诺日 = 12-17 新系统上线切换(范围含支付)**;专版 buffer 至 12 月底-1 月中(D-K)。
> 执行模式(D-L):AI 并行作战——细切片、契约先行、双线全并行、每日集成;工程师职责 = 指挥 coding agent + 审查 + 集成;完成门禁与协作规约见 README §5。
> 规模:S(≤1 天)/ M(2-4 天)/ L(≈1 周)/ XL(多人多批)。

---

## 时间轴总览(检查点与降级梯见文末)

> **实况(2026-10-28,W2 中段)**:W1 大幅超额并外溢——E0-E4/E6a/b/E7/E8 全收;E5 全族收满(E5-0..f3,10-26..28:框架+快照+report-single 页+读回面契约/结算线/概览线;E5-e 溶片=锚并 E-line-a/E10,10-28 拍板);PAY-C0 四片+PAY-C1 四片代码面完工+PAY1 五片,**CP1 支付门 10-25 提前 14 天实测达成**(真渠道验签握手绿;行政件=类目补设+H5 权限+APIv3 密钥 32 位核对);hzy 线全并 main(MR!6/!9);verify 工作台冻结(7ec87dea:原型保真验收线,重激活=开发完成后);**骨架侧测试剥离欠账回位已派(a=lzx/b=hzy,10-28)**;CP2 四项门实质已过线,余风险=E9/E10 未认领+ljr 待裁件(P13/P4/P7/P10;P15 已裁 D-AP,E5b 解锁可认领)。逐片状态见 progress.md 认领表。

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
| ✅ E3 评分引擎 | 三层落位:指标层(13 款指标定义单源档,本片创建)/门槛层(百分位档位)/评分层(无上限线性+权威分位双轨);Tier A 九款全量重算 + Tier B 四款闸门;red_dot 锚点初始校准常量;ScorerRegistry 插件契约(加测试零改主干)。验收:域函数单测+提交→重算→报告分 e2e。W3 三款(offense_defense/resource_match/character_counter)依赖题库种子增量:offense_scenario(optimal_choice/complexity/difficulty)+moba_character/character_counter_relation(counter_score)——内容=旧库经验数据(D-AL) | E0,E2 | XL |
| ✅ E4 匹配引擎 | match_config(预留 scenario 维度)+类型轴多数投票+两层编排+当日 seed+session 锁死落库+重测 3/1。细锚(D-AL):续测 7 天不重排、超窗以 test_record 反推、重测次数不恢复;切后台无效=阈值可配(>3s)+is_valid/invalid_reason 入库;done=缓存写入后有新 test_record;跳步=未测可跳、done 只看结果、前端不可指定款;游戏池=contribution(is_primary=1)∩category∈(genre,通用)。验收:同日稳定/跨日变化/重排换 seed e2e | E2 | M |
| ✅ E5 报告框架+能力报告(10-28 全族收满) | 板块插件化(supports)+缺维度不渲染七硬约束+完整度指示+同龄百分位(Z-score CDF+置信分层)+档位四色;M1 只实现单次能力报告所需板块。细锚(D-AL):同龄基准=同龄同性别(待确定只同龄;仅年龄段=区间基准标近似;性别=P8);完整度主指标 X/8 维;档位快照解锁时固化(已解锁不随口径变);较上局=该能力上一场测评(非同场次上次尝试);报告文案五区可后台配置。验收:缺维场景渲染收拢 e2e(锚处置 10-28:服务端=E-line-a,客户端断言=E10) | E3 | L |
| E5b 报告簇扩片(2026-10-23 终局对账立片;**P15 已裁 D-AP,范围=career 只读+trend 基础,认知仅策略渲染面,可开工**) | 报告中心入口+成长趋势基础层(b4 裁免费)+职业聚合页(D-AB 默认 open)+认知策略渲染面(ToC 默认 closed 零入口,策略面归 E8);服务端聚合=GL D1-D3,8 维公式源=旧仓 esport-server 的 docs 目录 OpenAPI_用户端接口文档.md(§3757 区)+REPORT_FRONTEND_SPEC §3 命名,三源合一入 metric.md 权重矩阵;系统套 6 级称号映射随 P2。验收=缺维渲染收拢+策略翻转 e2e | E5,P15,P2 | L |
| ✅ E6a ctr 纯壳 | 零 RPC:TanStack Start 新工程(Nx 目标面)+ react-aria/tailwind 接线(D-R)+ token 层(旧仓 DESIGN.md 为经验数据输入,按 D-O 合理实施;D-S dark-only 单套)+ 路由骨架(仅根路由+空白落地页,D-T:能力配置为交付面) + 微信 WebView 适配基线(safe-area/键盘 vh/无 hover 假设/横竖屏地基)+ shared-ts UI-free 四包(error/cookie/store/router-progress)接线;i18n=paraglide zh-CN+en-US(baseLocale zh-CN);**SEO 全量面(2026-10-18 补入,C 端公网)**:robots.txt + sitemap(hreflang/x-default)+ 流程页 noindex + 逐页 title。验收:dev/build/typecheck/lint 全绿,dark-only 语义 token 就位 | — | M |
| ✅ E6b 登录+画像流 | 微信/手机登录(消费 system 既有契约,不等 E0)、画像流(昵称必填/出生年月 8 档/未满 13 监护人硬闸/游戏偏好) | E6a,E0(画像流半片) | L |
| E6c 账号域收口(2026-10-23 终局对账立片) | SMS 骨架落地(README §1 已裁分层:capability+system provider)+密码登录/密码管理页(b3 §01 三路收齐)+绑定/合并流(b3 §02b:PHONE_EXIST/WECHAT_OPENID_EXIST 冲突拦截/自绑幂等/重绑解绑走客服/富吞贫一次性+审计)+account-security 聚合页;收编遗留清单 SMS 散行。验收=三登录路 e2e+合并反例(双方有数据不自动并)+冲突拦截反例。**10-25 联动注**:SMS 登录=手机浏览器支付腿(H5)前置;openid 绑定流=微信内 JSAPI 支付凭据(openid)供给者——E6c 优先级与 PAY 联动;微信 OAuth transport 注入缝(授权登录 dev 全链自动化,补 E6b 真人依赖课)随本片 | E6b,system provider 片,E0 增量(绑定契约) | M-L |
| E7 游戏框架+13 款 | 统一引擎框架(rAF+dt+refs 或 canvas,墙钟计时+暂停顺延单实现)、遥测采集(新契约)、参数化单流程路由、配置驱动运行时、真响应式横竖屏。三批:反应瞄准(red_dot/flash_aiming/multi_target/target_recognition)W3-W4;追踪技能(tracking_moving/moving_target/skill_trajectory/dual_zone)W5;决策资源(number_stars/offense_defense/character_counter/resource_match/economic_sim)W6。验收:每批 live 冒烟+遥测字段对齐(语义级) | E6,E3 | XL |
| E8 权益最小面 | entitlement 事件模型+激活码全生命周期(生成/批量/兑换原子守卫防超卖)+报告访问判定(policy open/membership/tenant/closed + scope 覆盖;D-AB)。R7=会员 31 天计/第 32 天过期+每日定时扫描(D-AL)。验收:兑换并发不超卖 e2e。**实况(10-25 E8-a 结构片✅)**:契约双面(ActivationCodeService=ACCESS_ADMIN 生成/列表;EntitlementService=BUILTIN 兑换/访问判定)+activation_code 表(migration 落库)+module/entitle 四件套自立(权益账本归位:EntitlementRP 实现自 pay 挪入,Grant 走类型别名,scope 词表进 shared kernel 单源);scope_kind 增 report_type 第三值(码兑 grant scope_ref=报告族名→唯一索引天然幂等);兑换仲裁=条件 UPDATE used_count<quantity(与 Settle 同族单仲裁者形,旧仓 XML 守卫同构核证)+同用户重兑幂等/跨用户竞态败者重查分流;码=AC+16 位去混淆字母表(crypto/rand,旧仓同形核证);访问判定默认策略落地(career open/cognitive 会员或码/school 仅码,config 化=b 片);错误码 6150-52+双客户端五语言文案;b(10-25✅):ENT_E2E=1 e2e 收验收锚(十人竞兑一码恰 1 成/9 EXHAUSTED/used_count=1/恰 1 行 grant;同用户并发双兑双幂等;胜者重兑不耗次;过期/无效码拒)+report_policy 表(每族一行 upsert,无行=出厂默认)+SetReportAccessPolicy(ADMIN)四值(D-AB 全量;tenant 值 M1 拒 6153=显式可见非静默闭)+判定序重构(码兑硬权益>policy>默认;closed 显式设=POLICY_CLOSED vs 默认闭=NO_ENTITLEMENT 两因可分);c(10-25✅):会员期内 trial 自动解锁=完成时刻事件钩子(test Submit 有效成绩落库后调 TrialGrantHook port→entitle 查会员活跃→grant trial 永久,source=membership;best-effort=成绩是已存真值,grant 失败仅 ERROR log 不毁成绩;不追溯入会前由钩子时点天然成立);**R7 修订建议(对 D-AL,待评审)**:31 天/第 32 天过期=已落地事实(grant 即 expires_at=paid_at+31d,e2e 31 天窗断言钉死)+判定全程惰性(expires_at>now 现算,E8-b e2e 证过期即拒)——**每日定时扫描不落码**:旧仓需扫描因其 status(active/expired)状态模型须搬运,新仓不可变账本把过期变成事实(expires_at 列)而非状态,扫描无消费者;触发器=通知/提醒系统出现(届时若需「到期前提醒」立延迟任务面)。**架构审查收口(10-25,整洁架构+DDD 三路对抗+用户两裁)**:①test→entitle 钩子违「cross-module write→domain event」→进程内事件面落地(shared.TrialCompleted 中立协议+bootstrap.EventSubscriptions 总线;test 发布/entitle 构造器订阅,**test↔entitle 模块边消失**);②pay settle→Grant 事务内跨模块写→**法条修订**(kratos §1 跨模块写=事件/事务内幂等账本写两合法形,已回灌);③operator 面拆形(activation.proto 拆 entitlement.proto C 端+activation_admin.proto 运营,SetReportAccessPolicy 归 EntitleAdminService,entitleadmin 模块自立+guard 增「entitle 禁回向」=删除闭操作);④Redeem 事务化(修「码扣权益未落零恢复」死局);⑤命名/tag 收敛(CodeRP→ActivationCodeRP/payOrderRP→orderRP/log tag 三形态,§2 扩条回灌) | E2 | M |
| E9 最小 admin 面 | admin client 挂 esports 模块:激活码+测试项/配置管理+报告访问策略(职业/认知 open/membership/closed)+题库种子查看;复用既有表格/搜索/动作列范式与 O-1/2/3 裁定。验收:开码→兑换→解锁闭环 e2e。**S4 增量锚(2026-10-23)**:题库 CRUD+Excel 导入、game-category 归类页(V19 自定义归类语义)、test-record 明细浏览(b3 §08⑤ 全 attempts)、批量导入、compliance-config(出生采集开关/付费年龄线/备案文案)、协议版本管理——剪裁随 P13 签认定 | E0 | M |
| E10 全链 e2e | 登录→画像→匹配→4 款→完成→单次报告→激活码解锁全链 live e2e;原型断言按语义级移植为验收基线(D-O) | 全部 | M |
| E-line-a 测试回位·esports 23 件(2026-10-27 审查役立片) | 骨架测试重建役欠账回位,法条=kratos §7 三家园:unit 17 件原位直回(match biz/engine 伴名+test 模块 15 伴名——全部为 biz/domain-companion 合法家);e2e 3 件迁 internal/e2e/(entitle/pay/report live_e2e,夹具前缀对齐,ENT_E2E 门);guards 3 件迁 internal/guards/(internal/shared 双守卫+ent mixin_test;sed 包名替换,mixin_order 集合扩含 ActivationCode 硬删表)。验收=esports:test 裸跑绿+ENT_E2E=1 三 e2e 活栈绿 | 骨架法条已立(§7) | M |
| E-line-b 测试回位·ctr 18 件(2026-10-27 审查役立片) | 同役 ctr 面,法条=tanstack §11 两家园:colocated vitest 原位直回(domains/error 1+test 模块根 4 件+13 games logic);基建已就位(vitest 5/jsdom 30/test target 全回)。验收=ctr:test 全绿+用例数对账 vs 剥离前 | — | M |

## hzy 外围启动线(2026-10-19 增补;W1-W2 桥接,定义归此、状态记 progress)

登录重构与 E0 余面为承重面留 lzx;hzy 切片圈定 ctr 非登录区,文件面与 lzx 线不相交(合流点仅 routeTree.gen 与未来一次提交适配器)。边界四律:①config 消费用 gen 真类型(已冻结 GetTest/GetTestConfig),mock 只落 transport 缝;②游戏运行时禁 import auth/services(插件契约准入);③不烤权威指标——客户端采集 rawDetail,评分归服务端(旧架构同构);④提交缝后置,E0 冻结后薄适配器接上,游戏零改动。任务分配判据(D-AJ,2026-10-23):不涉及骨架/底座(shared-ts/shared-go、proto 契约面、system、deploy)的任务可分配 hzy——不再限于本节圈定的 ctr 非登录区。

| 片 | 内容与验收要点 | 依赖 | 规模 |
|---|---|---|---|
| ✅ H1 ctr 系统态件 | error/empty/loading/退出确认+骨架屏+氛围布局(旧版 UserFlowAmbientLayout 语义级);兼 hzy onboarding 首片。验收:三链绿+走查 | E6a | S |
| ✅ H2 协议页+版本化承载 | agreement/age-required 页面+路由+版本参数;文案占位等 ljr 定稿;URL 形状开工前与 lzx 定约。验收:三链绿+走查 | H1 | S |
| ✅ H-A 引擎框架片 | E7 框架前拉:GameClock 单实现(墙钟+暂停顺延,fake timers 可测;旧版 13 处手搓计时收敛为 1)+参数化路由 /test/:code+GameModule 插件契约+overlay 族。验收:mock config 下 red_dot 可玩+时钟正确性测试 | H1 | M |
| ✅ H-B 反应瞄准批 | E7 首批前拉:red_dot/flash_aiming/multi_target/target_recognition;harness 免登录;rawDetail 采集反哺 E0 统一提交契约;玩法语义级对照旧版(D-O)。验收:每款 harness 可玩+明细字段与旧版 per-game SubmitData 语义对齐+mock 插桩净拔除(test-mock.ts 与其测试删除、transport.ts/env.ts 插桩还原、VITE_TEST_MOCK 零残留;拔除前插桩点全带 TODO(mock-removal) 标记——2026-10-20 用户裁定:临时插桩不常驻且必须 TODO 可寻) | H-A | M-L |
| ✅ H-C 判定回路闭环 | 边界四律④预定形态兑现:提交缝薄适配器(TestRunHost 终局采集——明细流+契约三计数 misses/missed_high/spawned_enemy+后台段;attempt_no E4 前恒 1,序数真律归 E4 会话)→SubmitTestAttempt 单发,游戏零改动;EndFace 增服务端判分回执卡(原始分/权威分/档位/valid/无效原因/retest_left,无效原因词表双语文案,权威分曲线待校准款按回执降级渲染,本地明细卡降为辅助;重试钮维持本地练习语义,真律归 E4)。验收=真人全链:登录→/match→一局 red_dot→服务端分数回执渲染+会话失效反例——esports 首条人工可 e2e 调试的完整测评闭环(mock 已于 H-B 拔净、读面已真,本片仅余判定回路) | H-B,E3-c | M |

## 支付片(D-B2:纳入承诺窗口,保持可插拔;2026-10-25 两步走裁定)

**两步走裁定(2026-10-25,用户)**:①在线支付通道(微信/支付宝/银联等多方式多平台)沉淀为骨架可复用资产——cyber-ecosystem 是通用开发基座,esports 只是首个真实业务验证;落 shared-go/capability/payment 家族(cache/mq/storage 同构:父接口+家族错误域+validate conformance+provider 子包),契约形状容纳多平台(已声明需求,非预测性抽象;但实现只做 wechat+mock 两 provider,空 provider 禁)。②开发环境不接真实通道,mock provider 模拟真实链路保支付闭环测试(下单→回调→结算幂等→退款);真实通道仅上线部署前后经 deploy 环境变量面切入,凭据值不入库。**官方实证(10-25 查证):APIv3 无沙箱环境,公众号测试号 appid 亦绑不了商户**——CP1 门=APIv3 验签握手通(零成本),0.01 真实闭环+全额退款归 CP3。

**归属判据(10-25 复议裁定)**:全局共享的是通道技术而非钱流——通道技术=capability(库级复用,全业务共享 provider 实现);订单/退款政策/权益授予=各业务聚合端到端;**不进 system**(身份宪法域无钱;scope 等业务形状禁入基座;订单/支付单跨服务分置=分布式事务陷阱;支付回调域名属 deploy 面,不耦合 system 发布)。同构短信分层先例(10-20 裁决:发送管道=capability,验证码登录=system provider);递延触发器=第二业务真实接入支付→立独立 pay 网关服务(消费 capability),仍非 system。

**Platform 装配形状(10-26 定稿;随 PAY1 首个消费者同片落码——kratos §5 无消费者不装配法条,不得先装)**:照 system platform_cache 双文件同构——`platform_payment.go`=conf 块(provider=mock|wechat 一字段切换+凭据字段,值走 deploy 环境变量面)→构造器分发(mock.New()/wechat.New())→`*payment.Payment` 容器+cleanup fn 进 wire;`platform_payment_handler.go`=`defaultPaymentError` 实例注入(五必填槽映射 infra 35xx 实例)+`HandlePaymentError` 包装=module 层错误转换唯一喉面;消费 module 只见容器类型与 handler。**多渠道并存(10-26 裁定,用户提)**:支付是首个多后端并存能力(cache/mq/storage 单后端足够,但容器模型天然支持多实例)——形状=每 backend 一容器实例并存(wechat.New()/alipay.New() 各一),路由归业务/platform:**provider(用户选的支付方式)×channel(接入模式)正交**(jsapi/h5/native 皆 wechat 之 channel),订单行记 provider,查单按 provider 取容器;契约层不加 mux(Gateway 查单入参无渠道,统一 mux=按单号轮询的假便利);platform 装配升级为按 conf providers 列表构造 map[provider]*payment.Payment;mock dev 单容器镜像 wechat 值集,alipay 落地且需双 mock 时值集参数化(递延)。**capability 家族质量门禁(10-26 定案)**:payment 家族语句覆盖率 **≥90%**(家族基线实测 cache 88.2/mq 95.2/storage 63.0 已知弱项;C0-a 首刀实测 97.2),C0 各片验收含实测报数(`go test -cover`),conformance 探针清单 100% 执行;storage 63% 为骨架面既有债登记在案,不立片待真实痛点触发。

**场景矩阵与契约锚(10-25 推演定案)**:访问环境决定支付方式——微信内=JSAPI;手机浏览器=H5(拉起微信 App 完成支付,终态必在 App 内);PC=Native 扫码(手机扫码+**前端轮询查单**,无浏览器回调);手机浏览器无微信=**显式不可支付降级态**(支付宝 H5 兜底递延,契约容纳 alipay 值)。支付凭据(openid)由登录供给:**手机浏览器腿完整链=SMS 登录→H5 支付(E6c 为 PAY-H5 前置)**;openid 绑定流(E6c)=SMS 账号在微信内 JSAPI 的凭据供给者——E6c 优先级与支付联动。契约锚:Channel=纯类型+命名约定 <provider>_<mode>,**值集与 Payer 必填归 provider 声明与执法(10-26 抽象无感裁定:父包不知任何渠道值,新 backend 注册零改父包;行为由各 backend conformance 钉)**+Credential{Kind,Value}中立(invoke/跳转 URL/二维码)+**AmountFen 命名类型**(全链 int64 分零换算,wire=APIv3 原生;不变式探针:金额>0/退款≤原单/累计≤原单/**重复通知不重复结算**/乱序容错);mock=状态机+铸造厂+投递器(可配延迟/**重复投递**——微信真实重发,回调幂等必须可测)+失败模式(超时/5xx);**mock 镜像 wechat 渠道值集**(同字符串声明三值,兄弟包不 import——消费方 channel 选择在 mock↔wechat 切换下零感知);对齐阶梯 L1 接口仿真/L2 线格式仿真/L3 协议时序仿真=第三方对接验收语言(**mock 对齐义务=接口边界:HTTP 原文进/事件出/失败模式同构/重复投递可测;渠道内部密码学字节非 mock 义务,mock 自持 mini 签名加密对**——10-26 分层裁定),对齐证明=共享 conformance 双跑(nats/pg 先例推广);v3crypto=微信协议镜像纯函数(签名/验签/AES-256-GCM/x509,**零第三方手写**,决策依据:窄面 7 端点+演进收敛两文件+父接口防腐可整体换官方 SDK,首版平台证书+微信支付公钥双模式齐)——**归 wechat 子包(父包纯契约,协议件随渠道走,alipay/unionpay 同构各归各包)**;wechat live conformance=env 门控(WECHAT_PAY_CONFORMANCE_*),零成本握手子集=CP1 门(判定=应答可验签即绿,公钥模式新商户的 404 亦是有效往返——10-26 修正)。PC 付费:capability 容纳(native 分支入 PAY-C1),前端付费墙形态候 P7/**ljr 拍板**;执行拆分=C0-a 契约层/C0-b v3crypto/C0-c mock provider/C0-d infra 35xx 错误段。**第三方对接 family 化平台原则**:第三方 API 边界一律 capability 家族化(双实现+共享 conformance+L1-L3 对齐),授权登录教训(E6b 真人依赖)由此制度化——conventions 升格(kratos 第三方边界律)随 C0 批落;微信 OAuth transport 注入缝(授权登录 dev 全链自动化)挂 E6c。

| 片 | 内容与验收 | 依赖 | 规模 |
|---|---|---|---|
| PAY0 pay 模块拆分(sku/order) | 结构债挂号(2026-10-27 外部基准评审定案):module/pay=双聚合根(SKU 目录/订单生命周期独立、按 ID 引用)+双 UC 双 service+biz 719 行,正中 kratos §3「unwieldy→aggregate boundary wrong」;esports dependency guard 已立 TestModuleSingleService 执法,豁免仅 pay 一条。拆分形:module/sku(SkuRP/SkuUC/SkuService)+module/order(OrderRP/PayChannelRP/OrderUC/OrderService),order 经 sku 的 repo port 读(跨模块读合法形);EntitlementRP 挪出先例=同向。**触发器=下一片触碰 module/pay 的切片开工前**(PAY1 五片已收,即 W7 PAY2 联调前)。验收=拆后 guard 豁免清空+PAY_E2E 全链绿+esports test/lint 绿 | kratos §3+K-STR-3 已立法 | S-M |
| PAY-C0 通道能力契约 | payment 家族:父接口(下单/查单/关单/退款/退款查/通知解析(验签+解密→领域事件)/对账下载)+家族错误域+mock provider(内存态全链)+conformance 探针(mock 全量常跑;wechat live=env 门控);契约以微信 APIv3 语义为主轴,支付宝/银联/租户侧简化对接可套入不绑死(骨架通用支持,见 M3 预判) | cache/mq 家族先例 | M |
| PAY-C1 微信 provider | APIv3 商户私钥签名/平台证书获取验签轮换/统一下单(JSAPI+H5 两 scene)/查单/关单/退款/回调解析;**零成本握手探针(GET /v3/certificates)=CP1 门实质**;真钱路径(0.01 闭环+退款)上线前联调。**实况(10-26 落地)**:四片收满——传输核(serial→公钥表归一双模式验签+防重放窗+惰性证书引导)/五法(三渠道下单+状态折叠,未知态保守拒)/回调解析(验签复用同表同窗,未知事件透传 ChannelType)/账单(签名票据+免签下载主机+SHA-256 hash 闸+gzip);公钥模式为主路径(官方 10-26 查证:**新商户仅有公钥模式,GET /v3/certificates 对新商户恒 404**);**退款契约增量:RefundInput+TotalAmountFen(微信 amount.total 必填,中立校验+mock 矛盾执法同步)**;家族覆盖率 payment 97.1/mock 90.0/wechat 91.3;探针清单=签名握手(verifiable-非401 即绿,含 404)+absent-order 折叠 NotFound | PAY-C0;类目补设(0.01 前置) | M ✅代码面(真握手待凭据) |
| PAY1 订单域 | 订单/幂等(同单复用+结算幂等)/状态机;订单行带 scope_kind/scope_ref(单次=trial,会员=period);SKU 读面(原价/现价结构,D-F 无立减展示);**消费 payment capability,dev 挂 mock 全链闭环 e2e,不被真实通道阻塞**。**实况(10-25 五片✅,20d61e2f→5b7e88a1)**:a 契约+两表+种子/b 渠道先行订单域+platform map 容器(50427bde)/c 回调面+结算事务+收敛器+ctr 边缘转发(1c88a4be)/d 对账+账单解析(家族覆盖 97.1/91.6/90.5,5b7e88a1)/e `PAY_E2E=1` 全链绿;PAY2 待 deploy 凭据面+行政两件(类目补设+H5 权限)。**资金安全件(10-26 职责分界定案,全在业务层)**:回调幂等=DB 唯一约束+状态条件更新(多实例下进程内去重非法);支付成功→权益授予=同库单事务;未决单定时查单收敛(渠道状态机为准,本地订单=投影,查单=崩溃恢复机制);每日对账任务(Reports 账单逐笔核订单表);outbox/mq **递延**(触发器=权益跨库或支付事件第二消费者,对齐 kratos §事件 MAY 法条)——capability 只供语义(幂等键/查单真相/账单运输/失败分类),重试策略/并发控制/调度属业务层。**合并前对抗审查(10-26,四视角+聚合二轮)**:P0=0(伪造回调双钥匙不可行/结算仲裁行锁自洽/IDOR 关闭/能力族机械面全守);A 档八修落地——PaidAt 用渠道 success_time(零值退本地钟,三视角独立命中同点)/对账日失败重试补取(NotFound 一次重试后推进,防无交易日卡死)/parseYuanFen 收紧+溢出守卫(双负号折叠正数等畸形拒)/证书信任表验后注册+失败回滚/proto+schema 两处注释改真(handled 不 ack 语义/period 并发双付事实)/边缘 64KiB 预检/公钥表键统一大写;**B 档递延(触发器在案)**:跨日退款对账盲区(账单 At=交易时间列,退款行当日覆盖/跨日不入窗→退款域设计输入)/公共回调面结算 grant 落 default 租户(TenantFromCtx 无主体回退;多租户面启用前须显式带 order 租户)/sku 丢 PeriodDays 时 settle 照付不 grant(sku 无写面,潜伏)/未知 serial 触发出站证书拉取无节流(公共面 1:1 放大)/e2e 对账用今天窗(生产昨天窗约定未覆盖) | **P1 已定(D-AB/D-AC,10-21)**,PAY-C0 | M |
| PAY2 通道接入与联调 | deploy 环境面切 wechat provider;0.01 真实闭环+全额退款(CP3 门);生产联调+对账;H5 支付权限(行政)就绪后开 H5 scene;**依赖注:JSAPI 要求商户绑定 appid 域 openid→上线前登录须切真实公众号 appid(测试号绑不了商户;登记上线清单+E6c)**;PC 端付费(Native 扫码)未注册,P4 裁决时定 | PAY1,PAY-C1,H5 权限 | M |
| PAY3 合规执法+付费墙 | 未成年拦截执法点(付费补填+18 岁禁付)/仅人工退款入口/会员态付费墙(P7 视觉稿,期内隐藏单次只留续费)。细锚(D-AL):年龄计算契约=满 N 岁为生日次日(仅大陆无时区);**S1 注:契约仅存年+月无日,「生日次日」须取月粒度近似(建议保守=次月起算满岁),本片定口径**;备案四落点=登录页底/个人中心/付费弹窗+订阅确认/报告页底;幂等参考旧库 openspec membership-pay(CREATED 同单复用+settleOrder 幂等) | PAY2,P7(12-02) | M |

## M2 运营与变现(核心组 W5-W7;余组 W8 可降级梯①)

- **核心组(承诺窗口内)**:会员/权益终态(按 D-AB)+SKU 多档;激活码全面(批次/导出/核销视图);订单+解锁记录;用户列表+档案核心 Tab。
- **余组(梯①候选,上线后首迭代)**:评级管理(P2)/增长裂变(分享卡+ref HMAC+归因)/消息中心/反馈客服/ToB 租户视图/仪表盘/报告配置面。

## M3 专版(承诺日外 buffer:12 月底-1 月中)

P4 裁决(11 月底前)→企业 App 授权登录→租户发码额度/功能差异配置;依赖 HL 文档(W1 跟进确认到位时点)。

**不分叉铁律(2026-10-25 裁定)**:专版不得与公版代码分叉——分叉即维护双路。登录/支付等对接差异一律落 capability provider 层或配置/模块注册表层(根规 §8 可插拔面最便宜层级优先);物理隔离需求递延至 delivery-time fork 最晚点。专版对接方式未定,本条为前置约束。**对接形态预判(10-25 用户)**:大概率租户 app 内完成必要对接、不穿透到公版 H5(第一租户=海量教育 e生涯 app,无微信级生态,大概率简化性对接);小概率类微信生态穿透。骨架侧按 family 化+provider 可扩展保持通用支持,具体对接随 P4 裁决。

## M4 存量迁移 cutover(预演 W8,切换 W9)

| 片 | 内容与验收要点 | 依赖 | 规模 |
|---|---|---|---|
| MIG1 迁移工具片(2026-10-23 终局对账立片) | 导入器落地 migration.md 全部口径:旧↔新 ID 映射持久化(覆盖软删行)、四档重算重放(A 明细/B 聚合/C 平移+legacy 标/D 作废)、身份翻译(密码 verify-then-rehash)、权益翻译(旧库已核实无支付网关无订单)、画像不完整态、报告快照保真搬运、种子重放、对账 SQL 族(行数+抽样逐字段+分布);W6-W7 建造、W8 生产副本脱敏预演(P6 日程+nginx 仿真)。验收=migration.md 血泪清单逐条过+预演对账全绿。**school→租户步时序=P14 裁后定** | E8(权益翻译目标面)、PAY1(订单域对齐)、E5(重算报告读面) | L |

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
| CP1 | 11-08 | E0/E1/E2 完成;E3 过半;E6a/E6b 完成(登录+画像可跑);商户号确认+APIv3 验签握手通(10-25 校准:官方无 v3 沙箱,原「沙箱通」字面永不达;门=GET /v3/certificates 商户签名+平台证书验签全链,零成本。**10-26 二次校准:新商户仅有公钥模式,证书端点对公钥商户恒 404——判定改「应答可验签即握手绿」(2xx 或 404 皆有效往返,唯 401/验签失败为红),探针已按此落地**。**✅10-25 实测达成:真渠道 404+Wechatpay 五头签名+我方公钥验签全链绿(应答体文案逐字印证公钥模式预判),absent-order 折叠 NotFound 同绿——支付门提前 14 天闭**) | 拉梯①(M2 余组移上线后) |
| CP2 | 11-22 | 游戏 ≥8 款(两批);E4/E8 完;E5 框架通;PAY1 完 | 拉梯②(游戏余款上线后两周 hot-add) |
| CP3 | 12-02 | M1 全链 e2e 绿;PAY2 沙箱全流程;M2 核心组可用 | 拉梯③(分批交付协商)+梯④(支付受阻激活码兜底、支付热补) |

检查点由三人 10 分钟站会核对,结果记 progress.md;拉闸即执行,不临场争论(D-M)。
