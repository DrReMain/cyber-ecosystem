# esports 开发计划 — 60 天作战计划(大框架定稿)

> 归属:esports 前沿开发档。总则/架构定案/裁决台账见 README;切片状态与认领记 progress.md。
> **D0 = 2026-10-18,承诺日 = 12-17 新系统上线切换(范围含支付)**;专版 buffer 至 12 月底-1 月中(D-K)。
> 执行模式(D-L):AI 并行作战——细切片、契约先行、双线全并行、每日集成;工程师职责 = 指挥 coding agent + 审查 + 集成;完成门禁与协作规约见 README §5。客户端面切片另执行设计对齐前置法(D-AQ,见 design-alignment.md):开工先查映射读命中原型源,收口 proven-by 带设计源清单。
> 规模:S(≤1 天)/ M(2-4 天)/ L(≈1 周)/ XL(多人多批)。

---

## 时间轴总览(检查点与降级梯见文末)

> **实况(2026-10-30,W2 尾,M1 全片收满+质量批收口)**:E0-E10+H 线全收(E5 全族+E5b 全簇 10-28;E-line-a 回位收口,E-line-b hzy 进行中);支付 PAY-C0/C1 代码面+PAY1 五片收,**CP1 支付门 10-25 提前 14 天实测达成**(**CP2 门 10-30 前瞻全项实质达成,提前 23 天**;行政件=类目补设+H5 权限+APIv3 密钥 32 位核对——用户侧即办);hzy 线全并 main(MR!6/!9)+新派 P10 语言设置行小片(D-AT);verify 工作台冻结(7ec87dea,重激活=开发完成后);ljr:P7 视觉稿提前出稿确认(三态 HTML 落库,原 12-02 门提前,PAY3 实现源就绪),待裁=P4(11-15 HL 红线)/P2/P5/P6/P16;E6c 全簇收口(10-30)+微信回填片(G-201)+**质量批三件**(真机回归 7 用例 6 族修复/G-208 浮层首刀 BottomConfirmSheet 单源/规范大审查 20 条三波全修+i18n 前缀清算 105 键+config 显式化——99c5bb06);MergePreviewer 判例升格 kratos §1 verdict-port 法条;D-AS 召回入窗:消息中心+反馈、分享裂变(+2 切片,建议 W7-W8);**候选下一片(10-30 二次校准,用户裁定)=M2 核心组拆片+首刀(CP3 门三构成项唯一未动者,原 W5-W7 提前开工:会员/权益终态+SKU 多档→订单+解锁记录→激活码全面→用户详情余 Tab 挂内容)/PAY0 pay 模块拆分(S-M 结构刀)/PAY3 合规执法(mock 通道零依赖行政件,P7 视觉稿就绪)/D-AS 消息中心+反馈+分享裂变(W7-W8 窗);MIG1 回原窗 W6-W7 建造(修正:M2 SKU 若动权益表/PAY3 落执法分支都会返工迁移口径——迁移等功能面定型,依赖绿≠该现在建;软依赖 P14 school 步时序);PAY2 行政件=即办但不占开发序(只闸 0.01 真钱闭环验证,代码面全走 mock)**。逐片状态见 progress.md 认领表。

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
| ✅ E0 契约首刀 | esports v1 proto 增量推进冻结(D-U:种子片→测试读面→统一提交→三计数反哺,193a9b7c);entitlement→E8、contribution→E3(D-AA);nullable 显式+判定阈值随配置下发 | README §0 | M |
| ✅ E1 服务骨架 | kratos-new-service 复制集:双面 accessLadder(ADMIN/MEMBER)+catalog 注册+deploy compose 面+system 消费;会员链探针 e2e 绿(兼 hzy onboarding 切片) | E0 | M |
| ✅ E2 schema+种子 | ent 首批九表+种子重放(修坑:dual_zone 枚举统一/multi_target 补行 D-Z/skill_prediction 不入/枚举约束防回归);D-V 撤回后按冻结契约 clean-room 重建;缩围补登 D-AA | E1 | M |
| ✅ E3 评分引擎(13 款收满) | 三层:指标单源(metric.md)/门槛百分位/评分双轨(线性+权威分位);Tier A 10/10 后 Tier B 三款升格消亡(71c33f4f);ScorerRegistry 插件+提交缝(4c5f0617);题库种子增量=旧库经验数据(D-AL);§指针存量清扫同日(43fd4b3d) | E0,E2 | XL |
| ✅ E4 匹配引擎 | match_config+类型轴多数投票+两层编排+当日 seed+session 锁死落库+重测 3|1;细锚(D-AL:续测 7 天反推/超窗重排/切后台 is_valid 入库)全落;e2e 三验收全绿(704c3388/07b87349/43fd4b3d) | E2 | M |
| ✅ E5 报告框架+能力报告(10-28 全族收满) | 板块插件化(supports)+缺维不渲染+同龄百分位+档位四色;细锚(D-AL:同龄同性别/档位快照固化/较上局=上一场/完整度 X/8)全落;报告文案五区后台可配;E5-e 溶片处置 10-28(服务端锚=E-line-a,客户端断言=E10) | E3 | L |
| ✅ E5b 报告簇扩片(10-28 全簇收口 a-e;认领 lzx) | 范围(D-AP)=报告中心四卡+认知 8 维策略渲染面(ToC 默认 closed 零入口)+成长趋势基础层(b4 免费;进阶=跨报告对比/训练处方/分位漂移随会员)+职业聚合只读页(D-AB 默认 open);称号映射随 P2。GL D1-D3 均有载体(D1=metric §七冻结/D2=综合分等权律/D3=cohort 键降级律)。**拆片**:a 契约三读面(S)→b 认知+趋势聚合(M)→c 职业聚合(M)→d ctr 四页(M-L,D-AQ 渲染态前置)→e 策略翻转 e2e(S)。**语料三裁**:①历史最佳百分位现无数据源(AbilityBests 诚实 null;ability_score 百分位列从未写)→read-time 复用 percentileFor(输入=best,同 cohort 键),快照仍唯一冻结源;digest best_percentile 随 b 回填销遗留。②职业页原型仅关闭态(「暂未开放·后台开关控制」),开放态=旧仓语义级(careerPotentials/RFS §5.4/OpenAPI 认知报告区)=D-AQ 白名单①。③旧 A/B/C/D 等级表废止,新档位 S/A/B/C 唯一;认知综合分=已渲染维度等权均值(缺维不计 0,与能力报告同律);趋势混类=FPS/MOBA 能力集不同综合分不可横比,序列按能力出+带 genre。策略执法=聚合响应内嵌判定(判定序复用);报告中心=组合既有面(digest+双 access)不立新 RPC。验收=缺维渲染收拢+策略翻转 e2e(片 e) | E5,P15(D-AP);称号随 P2 | L |
| ✅ E6a ctr 纯壳 | TanStack Start+react-aria/tailwind(D-R)+dark-only(D-S)+路由骨架(D-T)+微信 WebView 适配基线(safe-area/键盘 vh/无 hover);paraglide zh-CN/en-US;SEO 全量面(robots/sitemap/noindex/title);四包 UI-free 接线 | — | M |
| ✅ E6b 登录+画像流 | 微信/手机登录(消费 system 既有契约,不等 E0)+画像流(昵称必填/出生年月 8 档/未满 13 监护人硬闸/游戏偏好);微信真人链路 e2e 绿;遗留随对应片(SMS=E6c/consent=H2 真源/上限口径已裁写死) | E6a,E0(画像流半片) | L |
| E6c 账号域收口(2026-10-23 终局对账立片) | SMS 骨架落地(README §1 已裁分层:capability+system provider)+密码登录/密码管理页(b3 §01 三路收齐)+绑定/合并流(b3 §02b:PHONE_EXIST/WECHAT_OPENID_EXIST 冲突拦截/自绑幂等/重绑解绑走客服/富吞贫一次性+审计)+account-security 聚合页;收编遗留清单 SMS 散行。**E6c-0 定案(2026-10-28)**:拆片 a SMS 能力+发码契约+SMS 登录(S-M,**✅ 10-28**)→b 绑定+合并(M)→c ctr 客户端余面(M-L,D-AQ 渲染态前置;登录面已随 a 提前,**✅ 10-30 dd83d3d6**)→d 三路 e2e+微信 transport 注入缝+散行收编(M,**✅ 10-30**——AuthorizeURL 覆写补缝/微信路+合并反例 ENT_E2E/identity-routes Playwright 三锚/D-W 台账收口);语料裁定:①capability/sms 家族照 10-20 分层(payment 家族模板,dysmsapi+log 双实现+conformance),验证码存 cache 家族(TTL+counter 限速),provider 住 auth 模块与微信同层;②SMS 无跳转可铸——发码=新 PUBLIC RPC(限速),登录=LoginByIdentity oneof 加 sms{phone,code}(breaking 复核);③BindCredential 加 sms payload+契约自点名的 state→发起会话缺口必补;冲突拦截错误码族+自绑幂等;重绑解绑=客服语义无自助;④合并=富吞贫一次性+审计,双方有数据不自动并(e2e 反例锚);⑤密码三态:修改=既有 ChangePassword,设置(无密首登)+忘记(SMS 验证码)为契约增量;⑥account-security 独立页形态本片定。验收=三登录路 e2e+合并反例(双方有数据不自动并)+冲突拦截反例。**10-25 联动注**:SMS 登录=手机浏览器支付腿(H5)前置;openid 绑定流=微信内 JSAPI 支付凭据(openid)供给者——E6c 优先级与 PAY 联动;微信 OAuth transport 注入缝(授权登录 dev 全链自动化,补 E6b 真人依赖课)随本片 | E6b,E0 增量(绑定契约) | M-L |
| ✅ E7 游戏框架+13 款(10-25 收满) | 统一引擎框架(rAF+墙钟计时+暂停顺延单实现)+遥测新契约+参数化单流程路由+配置驱动运行时+真响应式横竖屏;三批收满(反应瞄准/追踪技能/决策资源);定律族立档(采样网格锚/自然终局/出生间距/开场布局;nil 让渡律) | E6,E3 | XL |
| ✅ E8 权益最小面(10-25 三片收满) | a 结构(entitle 模块自立+激活码全结构+兑换仲裁=条件 UPDATE+scope 词表单源)/b 验收锚(十人竞兑不超卖 e2e+report_policy 面+判定序:码兑>policy>默认)/c 会员衍生解锁(事件钩子 TrialGrantHook);架构审查收口两裁落地(进程内事件面+跨模块写两合法形法条回灌 kratos §1;operator 面拆形);**R7 修订:过期=事实(expires_at)非状态,每日扫描不落码(触发器=提醒/通知系统出现)** | E2 | M |
| ✅ E9 最小 admin 面(10-28) | admin 挂 esports 模块:激活码+测试项/配置+报告策略+题库查看(c1ca75d3+5895f1a8);闭环 e2e(开码→兑换→CODE_REDEEMED);O-1/2/3 裁定照用;**S4 增量锚剪裁已定(D-AS⑥,10-28):协议版本管理+compliance-config 随 E9 增量入窗;题库 CRUD/Excel/批量导入/测试记录明细随 M2 余组;报告配置面梯①** | E0 | M |
| ✅ E10 全链 e2e(10-28) | login 九红条消化(D-AQ 首片)+ctr 首 Playwright 全链 spec(CTR_E2E 门控;两 fixture 法=反应池收窄+cohort 样本,finally 全净清);原型断言语义级移植为验收基线(D-O)兑现 | 全部 | M |
| ✅ E-line-a 测试回位·esports 23 件(10-28 收口) | 三家园回位:unit 17 原位直回/e2e 3 迁 internal/e2e(ENT_E2E 门)/guards 3 迁 internal/guards(法条=kratos §7);含 E5-e 溶片锚(五幕+读回+composite 一致性) | 骨架法条已立(§7) | M |
| E-line-b 测试回位·ctr 18 件(2026-10-27 审查役立片) | 同役 ctr 面,法条=tanstack §11 两家园:colocated vitest 原位直回(domains/error 1+test 模块根 4 件+13 games logic);基建已就位(vitest 5/jsdom 30/test target 全回)。验收=ctr:test 全绿+用例数对账 vs 剥离前 | — | M |

## hzy 外围启动线(2026-10-19 增补;W1-W2 桥接,定义归此、状态记 progress)

登录重构与 E0 余面为承重面留 lzx;hzy 切片圈定 ctr 非登录区,文件面与 lzx 线不相交(合流点仅 routeTree.gen 与未来一次提交适配器)。边界四律:①config 消费用 gen 真类型(已冻结 GetTest/GetTestConfig),mock 只落 transport 缝;②游戏运行时禁 import auth/services(插件契约准入);③不烤权威指标——客户端采集 rawDetail,评分归服务端(旧架构同构);④提交缝后置,E0 冻结后薄适配器接上,游戏零改动。任务分配判据(D-AJ,2026-10-23):不涉及骨架/底座(shared-ts/shared-go、proto 契约面、system、deploy)的任务可分配 hzy——不再限于本节圈定的 ctr 非登录区。

| 片 | 内容与验收要点 | 依赖 | 规模 |
|---|---|---|---|
| ✅ H1 ctr 系统态件 | error/empty/loading/退出确认+骨架屏+氛围布局(旧版语义级);兼 hzy onboarding 首片;三链绿+走查 | E6a | S |
| ✅ H2 协议页+版本化承载 | agreement/age-required+路由+版本参数;文案占位等 ljr 定稿后整目录换值;三链绿+走查 | H1 | S |
| ✅ H-A 引擎框架片 | GameClock 单实现(墙钟+暂停顺延,旧版 13 处手搓计时收敛为 1)+参数化路由 /test/:code+GameModule 插件契约+overlay 族;mock config 下 red_dot 可玩 | H1 | M |
| ✅ H-B 反应瞄准批 | 四款引擎语义锚钉死+本地结果面;rawDetail 采集反哺 E0 统一提交契约;mock 插桩净拔除(TODO 法,零残留) | H-A | M-L |
| ✅ H-C 判定回路闭环 | 提交缝薄适配器+服务端判分回执卡(权威分待校准降级渲染);真人全链验收=esports 首条人工可走通闭环;撞面合流取 E4-c/E7-a 版为正身(详见 progress 认领表) | H-B,E3-c | M |

## 支付片(D-B2:纳入承诺窗口,保持可插拔;2026-10-25 两步走裁定)

**两步走裁定(2026-10-25,用户)**:①在线支付通道(微信/支付宝/银联等多方式多平台)沉淀为骨架可复用资产——cyber-ecosystem 是通用开发基座,esports 只是首个真实业务验证;落 shared-go/capability/payment 家族(cache/mq/storage 同构:父接口+家族错误域+validate conformance+provider 子包),契约形状容纳多平台(已声明需求,非预测性抽象;但实现只做 wechat+mock 两 provider,空 provider 禁)。②开发环境不接真实通道,mock provider 模拟真实链路保支付闭环测试(下单→回调→结算幂等→退款);真实通道仅上线部署前后经 deploy 环境变量面切入,凭据值不入库。**官方实证(10-25 查证):APIv3 无沙箱环境,公众号测试号 appid 亦绑不了商户**——CP1 门=APIv3 验签握手通(零成本),0.01 真实闭环+全额退款归 CP3。

**归属判据(10-25 复议裁定)**:全局共享的是通道技术而非钱流——通道技术=capability(库级复用,全业务共享 provider 实现);订单/退款政策/权益授予=各业务聚合端到端;**不进 system**(身份宪法域无钱;scope 等业务形状禁入基座;订单/支付单跨服务分置=分布式事务陷阱;支付回调域名属 deploy 面,不耦合 system 发布)。同构短信分层先例(10-20 裁决:发送管道=capability,验证码登录=system provider);递延触发器=第二业务真实接入支付→立独立 pay 网关服务(消费 capability),仍非 system。

**Platform 装配形状(10-26 定稿;随 PAY1 首个消费者同片落码——kratos §5 无消费者不装配法条,不得先装)**:照 system platform_cache 双文件同构——`platform_payment.go`=conf 块(provider=mock|wechat 一字段切换+凭据字段,值走 deploy 环境变量面)→构造器分发(mock.New()/wechat.New())→`*payment.Payment` 容器+cleanup fn 进 wire;`platform_payment_handler.go`=`defaultPaymentError` 实例注入(五必填槽映射 infra 35xx 实例)+`HandlePaymentError` 包装=module 层错误转换唯一喉面;消费 module 只见容器类型与 handler。**多渠道并存(10-26 裁定,用户提)**:支付是首个多后端并存能力(cache/mq/storage 单后端足够,但容器模型天然支持多实例)——形状=每 backend 一容器实例并存(wechat.New()/alipay.New() 各一),路由归业务/platform:**provider(用户选的支付方式)×channel(接入模式)正交**(jsapi/h5/native 皆 wechat 之 channel),订单行记 provider,查单按 provider 取容器;契约层不加 mux(Gateway 查单入参无渠道,统一 mux=按单号轮询的假便利);platform 装配升级为按 conf providers 列表构造 map[provider]*payment.Payment;mock dev 单容器镜像 wechat 值集,alipay 落地且需双 mock 时值集参数化(递延)。**capability 家族质量门禁(10-26 定案)**:payment 家族语句覆盖率 **≥90%**(家族基线实测 cache 88.2/mq 95.2/storage 63.0 已知弱项;C0-a 首刀实测 97.2),C0 各片验收含实测报数(`go test -cover`),conformance 探针清单 100% 执行;storage 63% 为骨架面既有债登记在案,不立片待真实痛点触发。

**场景矩阵与契约锚(10-25 推演定案)**:访问环境决定支付方式——微信内=JSAPI;手机浏览器=H5(拉起微信 App 完成支付,终态必在 App 内);PC=Native 扫码(手机扫码+**前端轮询查单**,无浏览器回调);手机浏览器无微信=**显式不可支付降级态**(支付宝 H5 兜底递延,契约容纳 alipay 值)。支付凭据(openid)由登录供给:**手机浏览器腿完整链=SMS 登录→H5 支付(E6c 为 PAY-H5 前置)**;openid 绑定流(E6c)=SMS 账号在微信内 JSAPI 的凭据供给者——E6c 优先级与支付联动。契约锚:Channel=纯类型+命名约定 <provider>_<mode>,**值集与 Payer 必填归 provider 声明与执法(10-26 抽象无感裁定:父包不知任何渠道值,新 backend 注册零改父包;行为由各 backend conformance 钉)**+Credential{Kind,Value}中立(invoke/跳转 URL/二维码)+**AmountFen 命名类型**(全链 int64 分零换算,wire=APIv3 原生;不变式探针:金额>0/退款≤原单/累计≤原单/**重复通知不重复结算**/乱序容错);mock=状态机+铸造厂+投递器(可配延迟/**重复投递**——微信真实重发,回调幂等必须可测)+失败模式(超时/5xx);**mock 镜像 wechat 渠道值集**(同字符串声明三值,兄弟包不 import——消费方 channel 选择在 mock↔wechat 切换下零感知);对齐阶梯 L1 接口仿真/L2 线格式仿真/L3 协议时序仿真=第三方对接验收语言(**mock 对齐义务=接口边界:HTTP 原文进/事件出/失败模式同构/重复投递可测;渠道内部密码学字节非 mock 义务,mock 自持 mini 签名加密对**——10-26 分层裁定),对齐证明=共享 conformance 双跑(nats/pg 先例推广);v3crypto=微信协议镜像纯函数(签名/验签/AES-256-GCM/x509,**零第三方手写**,决策依据:窄面 7 端点+演进收敛两文件+父接口防腐可整体换官方 SDK,首版平台证书+微信支付公钥双模式齐)——**归 wechat 子包(父包纯契约,协议件随渠道走,alipay/unionpay 同构各归各包)**;wechat live conformance=env 门控(WECHAT_PAY_CONFORMANCE_*),零成本握手子集=CP1 门(判定=应答可验签即绿,公钥模式新商户的 404 亦是有效往返——10-26 修正)。PC 付费:capability 容纳(native 分支入 PAY-C1),前端付费墙形态候 P7/**ljr 拍板**;执行拆分=C0-a 契约层/C0-b v3crypto/C0-c mock provider/C0-d infra 35xx 错误段。**第三方对接 family 化平台原则**:第三方 API 边界一律 capability 家族化(双实现+共享 conformance+L1-L3 对齐),授权登录教训(E6b 真人依赖)由此制度化——conventions 升格(kratos 第三方边界律)随 C0 批落;微信 OAuth transport 注入缝(授权登录 dev 全链自动化)挂 E6c。

| 片 | 内容与验收 | 依赖 | 规模 |
|---|---|---|---|
| PAY0 pay 模块拆分(sku/order) | 结构债挂号(2026-10-27 外部基准评审定案):module/pay=双聚合根(SKU 目录/订单生命周期独立、按 ID 引用)+双 UC 双 service+biz 719 行,正中 kratos §3「unwieldy→aggregate boundary wrong」;esports dependency guard 已立 TestModuleSingleService 执法,豁免仅 pay 一条。拆分形:module/sku(SkuRP/SkuUC/SkuService)+module/order(OrderRP/PayChannelRP/OrderUC/OrderService),order 经 sku 的 repo port 读(跨模块读合法形);EntitlementRP 挪出先例=同向。**触发器=下一片触碰 module/pay 的切片开工前**(PAY1 五片已收,即 W7 PAY2 联调前)。验收=拆后 guard 豁免清空+PAY_E2E 全链绿+esports test/lint 绿 | kratos §3+K-STR-3 已立法 | S-M |
| ✅ PAY-C0 通道能力契约(四片收) | payment 家族:父接口(下单/查单/关单/退款/退款查/通知解析/对账下载)+家族错误域+mock provider(镜像 wechat 渠道值集,兄弟包不 import)+conformance 探针(mock 常跑;wechat live=env 门控);v3crypto 归 wechat 子包(协议件随渠道走);infra 35xx 八枚;契约以微信 APIv3 语义为主轴,多平台可套入 | cache/mq 家族先例 | M |
| ✅ PAY-C1 微信 provider(10-26 代码面完工;10-25 真握手绿) | 传输核(serial→公钥表双模式验签+防重放窗+惰性证书引导)/五法(三渠道下单+状态折叠,未知态保守拒)/回调解析/账单(票据+SHA-256 hash 闸+gzip);公钥模式为主路径(新商户证书端点恒 404);退款契约增量 RefundInput+TotalAmountFen;覆盖率 payment 97.1/mock 90.0/wechat 91.3;握手判定=应答可验签即绿(见 CP1 行) | PAY-C0 | M |
| ✅ PAY1 订单域(10-25 五片,20d61e2f→5b7e88a1) | 订单/幂等(同单复用+结算幂等)/状态机;订单行带 scope_kind/scope_ref;SKU 读面(原价/现价,D-F);dev=mock 全链 PAY_E2E 绿,不被真实通道阻塞;**资金安全=业务层职责(10-26 定案:回调幂等 DB 唯一约束+条件更新/成功→权益同库单事务/未决单定时收敛/每日对账;outbox 递延=权益跨库或第二消费者)**;合并前对抗审查 P0=0+A 档八修落地;B 档递延五项登记 progress.md 遗留清单 | **P1 已定(D-AB/D-AC)**,PAY-C0 | M |
| PAY2 通道接入与联调 | deploy 环境面切 wechat provider;0.01 真实闭环+全额退款(CP3 门);生产联调+对账;H5 支付权限(行政)就绪后开 H5 scene;**依赖注:JSAPI 要求商户绑定 appid 域 openid→上线前登录须切真实公众号 appid(测试号绑不了商户;登记上线清单+E6c)**;PC 端付费(Native 扫码)未注册,P4 裁决时定 | PAY1,PAY-C1,H5 权限 | M |
| PAY3 合规执法+付费墙 | 未成年拦截执法点(付费补填+18 岁禁付)/仅人工退款入口/会员态付费墙(P7 视觉稿,期内隐藏单次只留续费)。细锚(D-AL):年龄计算契约=满 N 岁为生日次日(仅大陆无时区);**S1 注:契约仅存年+月无日,「生日次日」须取月粒度近似(建议保守=次月起算满岁),本片定口径**;备案四落点=登录页底/个人中心/付费弹窗+订阅确认/报告页底;幂等参考旧库 openspec membership-pay(CREATED 同单复用+settleOrder 幂等) | PAY2,P7(12-02) | M |

## M2 运营与变现(核心组 W5-W7;余组 W8 可降级梯①)

- **核心组(承诺窗口内)**:会员/权益终态(按 D-AB)+SKU 多档;激活码全面(批次/导出/核销视图);订单+解锁记录;用户列表+档案核心 Tab(首面已前拉 10-20);**D-AS(10-28)召回入窗**:消息中心/反馈客服、增长裂变(分享卡+ref HMAC+归因)——+2 切片,建议 W7-W8,未完走梯③。
- **余组(梯①候选,上线后首迭代)**:评级管理(P2)/ToB 租户视图/仪表盘/报告配置面。

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
| CP1 | 11-08 | E0/E1/E2 完成;E3 过半;E6a/E6b 完成(登录+画像可跑);商户号确认+APIv3 验签握手通(校准史:官方无 v3 沙箱+新商户仅公钥模式→判定=应答可验签即绿,公钥商户 404 亦有效往返;**✅10-25 实测达成,提前 14 天**——真渠道 404+签名+公钥验签全链绿,absent-order 折叠 NotFound 同绿) | 拉梯①(M2 余组移上线后) |
| CP2 | 11-22 | 游戏 ≥8 款(两批);E4/E8 完;E5 框架通;PAY1 完 | 拉梯②(游戏余款上线后两周 hot-add) |
| CP3 | 12-02 | M1 全链 e2e 绿;PAY2 沙箱全流程;M2 核心组可用 | 拉梯③(分批交付协商)+梯④(支付受阻激活码兜底、支付热补) |

检查点由三人 10 分钟站会核对,结果记 progress.md;拉闸即执行,不临场争论(D-M)。
