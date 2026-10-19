# esports 进度档 — 认领状态与验收记录

> 归属:esports 前沿进度档,三人共写。认领即登记;里程碑收口时验收记录剪枝为一行收口注(提交 hash + 判定),长历史删除。检查点结果每站在下表登记。

## 切片认领表(目标窗口见 plan.md 时间轴)

| 片 | 认领人 | 分支 | 状态 | 备注 |
|---|---|---|---|---|
| E0 契约首刀 | lzx | esports/dev-lzx | 种子片完成(待评审);增量2 完成(待评审) | D-U 增量推进;2026-10-18 种子=画像域+游戏目录读面(proto 3 文件,生成/lint/breaking/ctr 链全绿,catalog manifest 已挂联邦树);性别不入选转 P8(ljr);错误码 61xx 段;2026-10-18 增量2=测试域读面冻结落盘:test.proto 三读(ListTests 水合源/GetTest 详情+introduction/GetTestConfig 引擎启动包 duration 镜像+background_invalid 机制阈值+engine_params KV);设计侧对比修正 5 处(ability 六词 camelCase 原形、档位百分位服务端判 GradeBand 撤回、unit_score_label 补、introduction per-test 超集、ListTests 角色重定位水合源);max_score 删除获设计侧先例(admin test-type 已删满分);ESPORTS_TEST_NOT_FOUND=6110(404 码系)+ctr 双语文案;全链绿(generate/breaking/esports build+test/ctr build→typecheck→check);target_recognition 能力冲突转 P9;题库两读出列随 E7 |
| E1 服务骨架 | lzx | esports/dev-lzx | 完成(待评审) | 2026-10-18 kratos-new-service 八步全走:profile schema 从 E2 前拉(ent 零 schema 不可生成,键形按 AGENTS §8=user_id+租户列);init 迁移已 apply;catalog 经 manifest 自动挂树;部署面全接(compose/traefik/02-databases/deploy 目标/admin vite 代理);shared-go 注解测试纳入 esports;会员链探针 e2e 绿(登录→自省座位→GET profile 200 全 null 惰性态,无 cookie 401)。GameService 注册随 E2(目录种子先行才有意义) |
| E2 schema+种子 | lzx | esports/dev-lzx | **已撤回(D-V 业务前置律)** | 2026-10-18 当日先落地后撤回:game 聚合+9 款种子+画像写面全规则曾全绿(探针 16 断言),用户裁决接通先行后整体撤回(备份 /tmp/e2-loop-backup/ 至重启,契约重建即据冻结 proto;探针脚本 /tmp/probe_loop.py);服务回到 E1 探针形态复验绿(login→401→惰性 GET;写面 501)。重启时机=接通闭环验证过 |
| E3 评分引擎 | — | — | 未认领 | XL,可分批(Tier A 前半/后半、Tier B、档位层) |
| E4 匹配引擎 | — | — | 未认领 | |
| E5 报告框架+能力报告 | — | — | 未认领 | 服务/客户端两半可并行 |
| E6a ctr 纯壳 | lzx | esports/dev-lzx | 完成(待评审) | 2026-10-18 先行开工(例会追认);技术栈 D-R;主题 dark-only + i18n zh/en(D-S);同日补入 SEO 全量面(public:robots/sitemap hreflang/noindex 流程页/逐页 title);对照 admin 全量审计:20/33 字节级一致,13 处偏离全部参数化有据(生成器前置标准),4 处漂移当日修正;收口:三链绿 + live 冒烟全绿(含 404 错误面 LocaleSelect);同日 D-T:收敛为仅根路由+空白落地页(能力配置为交付面),login/profile/test 路由/页面/消息目录随之移除(E6b 登录片按范式重建),home 消息留 title+SEO description |
| E6b 登录+画像流 | lzx 主刀 | esports/dev-lzx | **接通片完成(待评审)=前后端最小闭环打通** | 2026-10-19 微信真人链路全绿:手机点按钮→静默授权→回调→服务端 code 交换→自动注册(首次)→HttpOnly cookie→/profile 渲染惰性画像;日志+DB 双证(LoginByIdentity 200→0.3s 后 GetProfile 200;user wechat.{openid}@id.local+credential 落库)。关键修复:①回调路由 Outlet 缺失(login.tsx 降布局+login.index.tsx,点命名子路由的正典拆法)②system 落 config 门控自动注册(auto_register,合成邮箱/无密码/默认租户 user 人群,事务内 user+seat+credential)③vite 端口钉死 strictPort(回调域名承重)④ctr 对齐 admin:shared-transport 消费+Sentry 接入+机制面零差;测试号凭据入 dev config(用户裁定非敏感)。**D-V 接通验证达成——后续切片进入纯业务堆砌期**。骨架演进件已按分支纪律拆出:dev 4 笔(e9c006c0 shared-transport 抽取+admin 消费/4242f533 system 自动注册+测试/e739a081 env 追踪例外/2655b539 tanstack 点命名嵌套法)cherry-pick 回流本支(658ce098..7e261422);5 维 review 修出 autoRegister 零测试、注释纪律违律×2、测试夹具业务 gen 依赖(dev 态暴露) |
| E7 游戏框架+13 款 | — | — | 未认领 | XL 三批,批内可并行 |
| E8 权益最小面 | — | — | 未认领 | |
| E9 最小 admin 面 | — | — | 未认领 | |
| E10 全链 e2e | — | — | 未认领 | |
| H1 ctr 系统态件 | hzy | esports/dev-hzy | 未开工 | 2026-10-19 增补(hzy 外围启动线,分支自 dev-lzx 现尖拉):error/empty/loading/退出确认+骨架屏+氛围布局(旧版 UserFlowAmbientLayout 语义级,D-O);零 RPC;兼 onboarding 首片(走 tanstack 三链+切片节奏法) |
| H2 协议页+版本化承载 | hzy | esports/dev-hzy | 未开工 | agreement/age-required 页面+路由+版本参数;文案占位等 ljr 定稿;consent 裁决面归 E6b(lzx),开工前向 lzx 要一句 URL 形状约定 |
| H-A 引擎框架片 | hzy | esports/dev-hzy | 未开工 | E7 框架前拉:GameClock 单实现(墙钟+暂停顺延,fake timers 可测;旧版 13 处手搓计时收敛为 1)+参数化路由 /test/:code+GameModule 插件契约(运行时禁 import auth/services)+overlay 族;config 消费用 gen 真类型(GetTest/GetTestConfig 已冻结,mock 只落 transport 缝);验收=mock config 下 red_dot 可玩+时钟正确性测试 |
| H-B 反应瞄准批 | hzy | esports/dev-hzy | 未开工 | E7 首批前拉:red_dot/flash_aiming/multi_target/target_recognition;harness 免登录;每款=运行时+rawDetail 采集+本地结果页;不烤权威指标(评分归服务端,旧架构同构);判定阈值来自 config;rawDetail shape 反哺 E0 统一提交契约;玩法语义级对照旧版(D-O);提交适配器等 E0 冻结后薄接,游戏零改动 |
| PAY1 订单域 | — | — | 未认领 | 依赖 P1(10-28) |
| PAY2 微信支付接入 | — | — | 未认领 | 商户号 W1 确认 |
| PAY3 合规执法+付费墙 | — | — | 未认领 | 依赖 P7(12-02) |

## 检查点登记(D-M)

| 检查点 | 日期 | 判定 | 拉闸动作 |
|---|---|---|---|
| CP1 | 11-08 | 待检 | — |
| CP2 | 11-22 | 待检 | — |
| CP3 | 12-02 | 待检 | — |

## 里程碑验收记录

(空;收口时一行一注)

## 待裁项状态

见 README §2.2;分析进展记 prd.md 各条目。

## 节奏记录

- 2026-10-18 立项+目录化;M1 切片池就绪。
- 2026-10-18 grill 定稿(D-K..D-Q):60 天作战计划落档,承诺日 12-17(含支付);大框架冻结,变更走 README 台账。
- 2026-10-18 D-R:ctr UI 基座 = react-aria+tailwind(无 antd);E6 拆 E6a 纯壳/E6b 登录+画像流,lzx 先行建壳。
- 2026-10-19 hzy 外围启动线定形(H1/H2/H-A/H-B):登录重构与 E0 余面为承重面留 lzx,hzy 全部切片圈定 ctr 非登录区(文件面与 lzx 线不相交,合流点仅 routeTree.gen+未来一次提交适配器);E7 框架与首批经 H-A/H-B 前拉,rawDetail 采集面反哺 E0 统一提交契约。admin 只读面前拉提案否决:BUILTIN 会员面(ListTests 等)不得作运营消费面,一面一消费等级,运营读面随 E0 余面/E9 契约设计走。
