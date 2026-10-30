# esports 进度档 — 认领状态与验收记录

> 归属:esports 前沿进度档,三人共写。**剪枝律(2026-10-23 立,2026-10-28 补执法细则,约束三个 agent;全文见 README §5)**:片收口即剪枝为一行收口注(状态+日期+提交 hash+关键判定);节奏行=索引形(日期+切片+hash+一短语,≤150 字),细节归提交信息 proven-by 尾;活档区(快照/认领表/遗留清单)为唯一可变区,收口时重写而非追加;历史节奏行随里程碑收口压实迁 progress-archive.md;里程碑收口做一次活档漂移对账——档是给 agent 读的,读档成本即团队成本。检查点每站在下表登记。

## 现状快照(2026-10-30,W2 尾)

| 人 | 已交付(均已并 esports/main) | 当前 → 下一片 |
|---|---|---|
| lzx | E0-E2、E3 评分(13 款收满,71c33f4f;§指针清扫 43fd4b3d)、E4 匹配全量(704c3388/07b87349)、E6a/E6b、E7-a..e 13 款引擎收满(fa866414/dcce4fc6/1ce16289)、V1-V4+B 批(6b98c886/865511c1)、会员读面+admin 用户管理(88bcd9ba/3f272b28)、组合缝(a8b9f0a4)、校准战役 S1-S4+D-AL(9e0478e6/fe73a63a/1c1a8424)、支付 PAY-C0 四片+PAY-C1 四片+PAY1 五片(CP1 支付门 10-25 提前达成;APIv3 密钥 33 字符待核=回调解密/CP3 前必须修)、流转批三件(ICP 备字+种子国服重播;偏好上限终态=常量写死,10-26 二裁推翻 config 化)、E8 权益三片(10-25 收满,含架构审查收口)、E5-0 定案+a/b1/b/c(10-26,二轮审查 15 项修复批同日)、MR!9 吸收+wire-stamp 修复笔(e62ce8c6,10-27)、E5-d+E5-f 读回面族(E5-d 69721f93/f 族 0ae81a28,10-28)、E9 最小 admin 面(c1ca75d3+5895f1a8,10-28)、D-AQ 设计对齐前置法立档(abacc542)、E10 全链 e2e(login 九红条消化+Playwright 全链,10-28)、E6c 全簇 d 片+E5b 全簇(10-28..30)、微信回填片 G-201(10-30)、质量批三件:真机回归 7 用例 6 族修复+G-208 浮层首刀 BottomConfirmSheet 单源+规范大审查 20 条三波全修与 i18n 前缀清算(105 键)+config 显式化(SMS backend 必填枚举+boot 横幅+APP_ 前缀法;99c5bb06,10-30) | **E6c+质量批全收(10-30)**→**候选下一片(10-30 二次校准,用户裁定):M2 核心组拆片+首刀(CP3 门唯一未动构成项:会员/权益终态+SKU→订单+解锁记录→激活码全面→用户详情余 Tab)/PAY0 拆分(S-M)/PAY3 合规执法(mock 通道不等行政件,P7 视觉稿就绪)/D-AS(W7-W8);MIG1 回 W6-W7 原窗;PAY2 行政件即办不占开发序(只闸 0.01 真钱验证)** |
| hzy | H1/H2/H-A/H-B、10-22 精修批、10-23 H-C 判定回路(真人链路验收;撞面合流正身=E4-c 版)、10-23/24 设计保真批(MR!6 于 10-25 并入 main,599dddce)、10-26 走查修批九条(93f67de3..179b1af2,见归档;vitest 112→115)、10-27 验收收口(tm/home-D5 亲验)+性别卡(6beff1c1,销 D-AN 流转半边)+走查下午批(MR!9 已并 main) | **E-line-b ctr 测试回位**(10-28 派定,进行中)→**P10 语言设置行小片**(D-AT 已裁,现成 LocaleSelect 挂 profile 设置区) |
| ljr | P1 计费(D-AB/AC)、P3 完成页(D-AI)、P9 能力归属(D-AD)、D-AG/AH 同批、P15 报告簇(D-AP,10-28)、P10 已裁(D-AT)、P7 四参数(D-AU)+**P7 视觉稿提前出稿并经确认**(三态 HTML 落库 p7-member-paywall.html,10-30——原 12-02 门提前,待交 PAY3 实现) | 待裁:P16 文案键半边(待标定款回执)、P4(11 月底;HL 问卷回复红线 11-15)、P2/P5/P6(P6=迁移演练日程,12-07 前);~~P7 视觉已出~~;~~P13 已签毕(D-AS)~~;~~P15/P8/P11/P12 已裁(D-AN/D-AM/D-AP,10-28 刷新)~~ |

**M1 燃尽**:E0✅ E1✅ E2✅ E6a/E6b✅;E3 ✅(13 款收满,报告分 e2e 挂 E5);E4 ✅(a-d,e2e 三验收);E7 ✅(13 款收满);E8 ✅(三片收满,10-25,含架构审查两裁落地);E5 ✅ 全族收满(E5-0..f3,10-26..28;E5-e 溶片=锚并 E-line-a/E10,10-28 拍板)+E5b 全簇✅(10-28);E9 ✅(10-28)+E10 ✅(10-28);**E6c ✅ 全簇(0..d,10-28..30,含微信回填片与真机回归收尾)**;M2 用户列表+档案首面已前拉;V1-V4+B 批落,admin 对齐批待排;支付 PAY-C0✅+PAY-C1✅+PAY1✅五片(**CP1 门 10-25 提前 14 天实测达成**:真渠道验签握手绿);PAY2-3 未启动(卡行政件:经营类目补设+H5 支付权限申请——**用户侧即办**;APIv3 密钥 32 位核对=CP3 前必须)。**CP1(11-08)前瞻:门全项实质达成;余=E-line-b 回位(hzy 进行中)。CP2(11-22)前瞻:全项实质达成(游戏 13 款≥8/E4/E8/E5 框架+收满/PAY1)——提前 23 天。**

## 切片认领表(窗口与定义见 plan.md)

| 片 | 认领人 | 状态 | 收口注 |
|---|---|---|---|
| E0 契约首刀 | lzx | ✅ 10-21 | 种子(画像+目录)+增量2(测试读面)+增量3(统一提交,4 决策经裁)+增量4(三计数反哺,193a9b7c);余面已裁去向:报告读面随 E5/E9、entitlement 随 E8(D-AA) |
| E1 服务骨架 | lzx | ✅ 10-18 | kratos-new-service 八步;双面 accessLadder+catalog+部署面;会员链探针 e2e 绿 |
| E2 schema+种子 | lzx | ✅ 10-20 | D-V 撤回后按冻结契约 clean-room 重建;五表+三读+13 款种子骨架(校准值不杜撰,随 E3 单源);multi_target 补行=D-Z |
| E3 评分引擎 | lzx | ✅ 10-23 | 13 款全量定稿收满:metric.md 单源+Registry 插件+提交缝(4c5f0617),Tier A 10/10(ce763107/f9ff14d7/5bb5db3f 等)后 Tier B 三款升格消亡(71c33f4f,j/k/l);双轴评审实错吸收族在案;§指针存量清扫同日独立片清零(43fd4b3d);余面=报告分 e2e(挂 E5)+guard 正则回流(随 E-line-a) |
| E4 匹配引擎 | lzx | ✅ 10-23 | a 聚合出生+b 提交缝(704c3388,双轴评审十项吸收:序数部分唯一索引原子化/先验再关/竞态读回)+c ctr 切服务端编排+attempt_no 清号(07b87349)+d e2e 三验收全绿(43fd4b3d:同日稳定/跨日变化/重排换 seed);×1.5 偏好权重裁不落;观察=test 目录第二排序键(入开放清单) |
| E5 报告框架+能力报告 | lzx | E5-0..f3✅(10-28 全族收满);E5-e 溶片(10-28 拍板) | E5-a=第一手三源综合(旧仓 PeerComparison 四字段逐字采纳/原型 tierOf 与档位表同源核对/compScore 缺维不计 0 进契约注释);E5-c=双保险幂等快照(事件+lazy 兜底,unjudged/none 诚实哨兵,ENT_E2E 五幕锚);E5-d=ctr report-single 页七段面+双入口+match 锚增量(anchor_record_id 派生自 scored records);E5-f=成绩读回面(原型 44 页对照盘点出的系统性欠账,见 10-28 节奏行);E5-e 处置(10-28 拍板)=不立独立片:服务端锚(E5-c 五幕+f 族读回+composite 去重一致性)并入 E-line-a internal/e2e/report,缺维/诚实面客户端断言归 E10 |
| E6a ctr 纯壳 | lzx | ✅ 10-18 | D-R/S/T 全落;SEO 全量面同日收口 |
| E6b 登录+画像流 | lzx | ✅ 10-20 | 接通片(微信真人链路绿)+完整形态(登录/三步向导/个人中心/首页)+机制修复+R1R2+流程前段四页;遗留契约缺口随对应片:SMS 骨架(system provider)、consent 版本记录、自定义游戏上限 3v20 口径 |
| E7 游戏框架+13 款 | lzx(E7-a..e) | ✅ 10-25 收满 | 反应瞄准批 E7-a、追踪技能批 E7-b/c、决策资源批 E7-d/e(offense_defense+resource_match 10-25,number_stars+character_counter+economic_sim 10-25 三款收 13);引擎先行;种子 engine_params 全量物化(nil 让渡律)+缩放守卫族+采样网格锚定律+自然终局定律+出生间距定律(E7-d)+开场布局守卫族(E7-d/e)均立;详各节奏行 |
| E8 权益最小面 | lzx | ✅ 10-25 三片收满 | a 结构(activation_code+仲裁链+scope 词表单源)/b 验收锚(十人竞兑不超卖 e2e+report_policy 面)/c 会员衍生解锁(事件钩子);架构审查收口(事件面落地+跨模块写两合法形法条+operator 面拆形);D-W population 余件触发器=E8 面已开工,余件仍待 |
| E9 最小 admin 面 | lzx | ✅ 10-28 全收(a=c1ca75d3 服务端,b=5895f1a8 客户端) | a:TestAdminService 六 RPC+策略读面+admin_readback 五面 e2e;b:admin 三页(激活码批次/核销进度、报告策略 set+回显、测试项+调参+题库抽屉)原型逻辑/文案/基本视觉对齐(码表明细/停用/导出=契约无不造,gap 在档);闭环 e2e 实走:UI 开码→member 兑换→access=CODE_REDEEMED;走查抓 E8 存量 bug=codePB 域词查 proto 名表恒 UNSPECIFIED(修复验证);6163 admin 文案漏批补;S4 增量锚(题库 CRUD/Excel/compliance-config)随 P13 剪裁 |
| E10 全链 e2e | lzx | ✅ 10-28 收口 | a=login 九红条消化(D-AQ 首片,渲染态读原型含密码 tab 态;修复 3/依赖 E6c 1/已裁压倒 1/随 PM 2/态比对 1/dev-only 1,gap-ledger 逐条处置);b=ctr 首 Playwright 全链 spec(CTR_E2E=1 门控):登录→向导→match→4 款→完成页→单次报告→激活码兑换 CODE_REDEEMED;两 fixture 法(反应池收窄+cohort 9 样本)finally 全净清;E5-e 客户端断言落位(免费座/缺维不渲染/付费墙) |
| E-line-a 测试回位·esports 23 件(2026-10-27 审查役立片) | lzx | ✅ 10-28 收口 | 三家园回位:unit 17 原位直回(match biz/engine+test 模块 15 伴名)/e2e 3 迁 internal/e2e/(entitle/pay/report live_e2e+夹具前缀对齐)/guards 3 迁 internal/guards/(internal/shared 双守卫+ent mixin_test;sed 包名替换法,mixin_order 集合扩含 ActivationCode 硬删表;conventions guard 正则修正随本片)。验收=esports:test 裸跑绿+ENT_E2E=1 三 e2e 活栈绿;**E5-e 并入(10-28 拍板)**:report live_e2e 扩为 E5-c 五幕锚+f 族读回锚(GetSessionReport/GetAbilityDigest fixture 断言)+composite 一致性(重测 board 两面同数) |
| E-line-b 测试回位·ctr 18 件(2026-10-27 审查役立片) | hzy | 10-28 派定 | 两家园回位:colocated vitest 原位直回(domains/error 1+test 模块根 4+13 games logic);基建已就位(vitest/jsdom/test target 全回)。验收=ctr:test 全绿+用例数对账 vs 剥离前 |
| 会员读面+admin 用户管理 | lzx | ✅ 10-20 | MemberService 读面(88bcd9ba)+admin esports-users 列表/详情 Tab 骨架(3f272b28)+ListUsersByPopulation 组合缝(a8b9f0a4,D-X 落位);裁:一面一消费等级,BUILTIN 会员面不作运营消费面;详情余 Tab 随 E2/E5/E8 挂内容 |
| H1-H2(系统态件/协议页) | hzy | ✅ 10-20 | MR!2 双轴评审合并;EN 协议译文 ljr 定稿后整目录换值 |
| H-A 引擎框架片 | hzy | ✅ 10-20 | GameClock+GameModule 契约+TestRunHost+/test/$code+red_dot;biome 插件契约执法 |
| H-B 反应瞄准批 | hzy | ✅ 10-21 | 三款引擎语义锚钉死+本地结果面;mock 插桩净拔除零残留;E0 增量4 反哺闭环 |
| H-C 判定回路闭环 | hzy | ✅ 10-23 | 判定回路=esports 首条人工可走通全链,已闭——**与 lzx E4-c 同日并行交付同一地皮(切片定义重叠的计划期裂缝),合流取 E4-c/E7-a 服务端真值版为正身**(提交适配器/判决卡/会话面/attempt_no 清号,详见 E4-c 节奏行);H-C 侧净贡献=**真人链路验收**(chrome-devtools 驱动+用户在场:阵容过滤/三入口直达/正例 raw 550+S+权威 96.5/无效反例 116ms 机器人反应判 ABSURD_REACTION_TIME 红章双语/待校准降级 ×2/257 事件逐字段过线/单发+锚点满程/6132 探针 404)+P16 登记(转 ljr)+D-AG/C3 走查销项;H-C 自有代码孤件随合流净删,撞面复盘见节奏行 |
| PAY-C0/C1+PAY1-3 支付片 | PAY-C0✅+PAY-C1✅全收含真握手(lzx,10-25);PAY1✅五片(lzx,10-25);PAY2-3 未认领 | C0 契约/v3crypto/mock/35xx 四片落地;C1 传输核/五法/回调/账单/live 探针落地(公钥模式主路径);**CP1 门✅10-25 实测:真渠道握手 404+验签绿+absent-order 折叠 NotFound(凭据经授权从 GitLab wiki 取,/tmp 隔离不入库)**;P1 已定(D-AB/AC);商户资质就绪;**APIv3 密钥 wiki 记录 33 字符(应 32)待核对——握手/查单不受影响(不解密),回调解密与 CP3 前必须修正**;遗留=经营类目补设+H5 支付权限申请(建议即办);新依赖=JSAPI 需商户绑定 appid 域 openid,公众号测试号绑不了商户,上线前登录切真实公众号 appid |
| MIG1 迁移工具片 | — | 未认领 | 2026-10-23 终局对账立片;W6-W7 建/W8 预演;school 步时序=P14;依赖 E8/PAY1/E5;**10-30 校准修正:依赖全绿≠现在建——M2 SKU 若动权益表/PAY3 落执法分支都会返工迁移口径,功能面定型后原窗开工** |
| E5b 报告簇扩片 | lzx | E5b-0/a/b/c/d/e ✅ 10-28 全簇收口 | E5b-0=设计定案+拆片+语料三裁(详见 plan 行):历史最佳百分位=read-time 复用 percentileFor(现无数据源);职业开放态=旧仓语义级(原型仅关闭态);旧 A/B/C/D 等级表废止;策略执法内嵌聚合响应。a=契约三读面(report.proto 增 GetCognitive/GetTrend/GetCareerReport,全 BUILTIN,ReportAccessVerdict 判定内嵌三响应,趋势点带 genre,职业=推荐行+差距处方单数组,零新错误码;gen 11 文件纯增量,breaking/build/test/ctr:build 真码全绿)。b=认知+趋势聚合落地+digest 百分位回填(**B 裁定:判定序上移 shared.ReportAccessVerdict 单源双消费,entitle GetReportAccess 重构为消费方;AccessJudge UC→UC 注入形违 kratos §1 红线被用户叫停撤形**;退化池黄金锚=cohort 全员同分→侧席恒 50,八维全 50/tier C/缺维法 4 能力喂 2 维;趋势=ScoredSessions 时序+快照优先 live 兜底;ENT_E2E 活栈全绿含 guards)。c=职业聚合落地(careerAreas 四域种子库:旧仓门控 75/70+min 匹配 transcribe,维 key 经 metric.md 映射表译至冻结矩阵拼写;处方=已渲染维升序最低 4 行,target=系统 tierA 85 单源零私设线,gap 负值禁出;双 helper reportAccessOf/renderedDims 重构认知面同源零复制;库事实=行业级统计借挂(career_statistics 青训 95%/日均 12-14 小时),分域数据待 ljr;真空白用户空页断言;ENT_E2E 活栈全量绿)。e=翻转+零值锚 e2e(career/cognitive 置 closed:显式行压过 membership 而兑换码压过行,covered 仍公开,趋势 advanced 免疫一并钉界;独立 cohort 键 2009 造确定性 100→gap=0 已达线行=protojson 零值回归线;partial 趋势恰 4 序列收拢缺维;净清 8 探针 0) |
| E6c 账号域收口 | lzx | 全簇 ✅ 10-30(0+a bf0a99a0;b 2e0baf09;c dd83d3d6;d 本日) | 2026-10-23 终局对账立片;SMS+密码+绑定合并;收编 SMS 散遗留行;**10-25 联动注:SMS=手机浏览器支付腿(H5)前置,openid 绑定流=微信内 JSAPI 凭据供给,E6c 优先级与 PAY 联动;微信 OAuth transport 注入缝随本片**;E6c-0(10f67ab2)=拆片 a-d+语料六裁(详见 plan 行:capability/sms 家族照 payment 模板/发码新 PUBLIC RPC/BindCredential 补 state 系会话/合并=富吞贫+审计/密码三态增量/account-security 形态本片定);**a(10-28,bf0a99a0)**:capability/sms 家族落地(payment 模板:Message/Sender+哨兵 36xx 映射+dysmsapi 手写 RPC 签名客户端(协议强制 HMAC-SHA1,G505 行内 #nosec 特例——法条唯一例外类,优于 config 级豁免;openssl 金向量+httptest 协议真验签探针)+mock 后端 New/Open/ControllerOf+双 conformance);system 契约=SendSmsCode PUBLIC RPC(发码即铸同款 one-shot state,LoginByIdentity required state 零改动滑入既有 cashState)+payload oneof 加 sms{phone,code}(breaking 绿)+SYSTEM_SMS_RATE_LIMITED/CODE_INVALID 6006/6007;auth 模块=registry 重构(init 注册→wire 注入 map,redirectProvider/codeIssuer 接口隔离)+smsProvider(IssueCode 限速铸码发码+Exchange GetDel 烧码+大陆手机号 canonical);conf sms 行加显式 backend 选择(dysmsapi=强制真实缺凭据 boot 错/mock=固定不拨号/空=按凭据自动,与 payment 平台同款开关——真实通道验一次全链后开发固定 mock 省成本)+compose 镜像;双端 copy 7 locale;ENT_E2E=发码→窗口限速→错码烧码→state 重放拒→全链落会员净清;**审查修(10-28 三轴对审)**:NewSms 补 mock/dysmsapi 显式分支+unknown 值拒 boot(装配表测 100%)+container/config 镜像 sms 行+e2e 全腿穿 service 层+日上限恰 10/11 边界+失败发码覆盖旧码生命周期锚+非 issuer 拒镜像+dysmsapi 传输三形态(坏端点/短体/net.Error 超时)+Timestamp ISO 断言+兄弟熵失败统一返错(panic 消)+state 铸造去重;覆盖率实测 sms 家族 93%/platform_sms 100%;**ctr 接入+双轮拟真(10-28 晚,原 c 段登录面提前)**:登录页 SMS tab 转真(码 tab 默认/useSmsLogin hook+ui/sms-fields 拆分/手机号双闸实时校验/60s 倒计时换号清窗——服务端才是真限速器/发送也过协议门/infra 族 reason 兜底文案不再静默);mock 轮+真实通道轮(Aliyun dysmsapi,APP_SMS_ 前缀 env 注入)Chrome 拟人操作全链均绿:发码→真手机收码→登录→新用户向导→落 member 座→净清;live conformance 真通道 PASS(签名/凭据/模板被真 edge 接受);kratos env 注入唯一入口=APP_ 前缀(裸变量静默忽略);**b(10-29,2e0baf09)**:契约=BindCredential oneof 加 sms(4)+state 系会话缺口闭合(identityStateValue{provider,bind_user};BUILTIN 孪生面 BeginBindCredential/SendBindSmsCode 铸绑定态——rung 即安全语义,PUBLIC 面匿名铸;匿名/异会话态兑换前拒)+MergeAccounts(形状一:调用者=保留方,凭证证明=防接管闸;场景二当前会话=空号由客户端分解=先登富号再并,服务端零方向输入)+6008 NOT_MERGEABLE/6009 NOT_BOUND;合并=富吞贫一次性(account_merges 台账 poor_id unique=one-shot race backstop,记 provider/kind/identifier)+被并号 enabled=false+RevokeSessions+NotifyChanged 镜像禁用路径;空号判据两层:system 侧硬闸(无密码+单凭证+member 座+在册)本片执法,esports 侧(测评/权益/邀请)归 E6c-c offer 面——system 不见跨服务数据;自绑幂等/6005 拦截存量已在;重绑解绑零建设(客服);偏差=语料 IP 列不落台账(边缘拓扑/XFF=deploy 面,biz 层可伪造;audit 中间件 WRITE 行已记);验证=unit 三件(会话绑定律重写/合并全律/绑定铸态)+ENT_E2E 六锚(匿名态拒/自绑幂等/冲突拦/富吞贫正例+台账+被并会话亡/密码重量反例)+live 探针(401×2+validator 400+线上 redis 绑定态 bind_user 实证);双端 copy 7 locale+copy.ts 联盟键;**c(10-30,待提交)**:契约增量=system 三 RPC(ResolveCredential BUILTIN=占用事实{owner,mergeable,credentials 掩码}——bind-phone 发码前预检「占用无需发码」原型法/ResetPassword PUBLIC=忘记+设置合一,短信码到绑定手机=证明;设置页只持掩码→客户端 typed phone input 补全,契约零改动/GetSecurityStatus BUILTIN=security 页三行真状态)+esports PreviewMerge BUILTIN(合并 offer 面:sms 分支裸 phone 查询/wechat 分支携 OAuth 证明兑换即烧——微信腿三次授权形态)+MergePreviewStatus 五值(FREE/SELF/KEEP_CALLER/KEEP_OTHER/BLOCKED);**方向感知重量闸**(MergeUC 住 report 模块实现 profile.MergePreviewer 接口倒置+wire.Bind):先算双方 emptiness(esports 侧 sessions=0∧无会员∧无解锁,邀请表未建空真)再按方向查被并方 system 重量(KEEP_CALLER→other.Mergeable/KEEP_OTHER→caller 密码∨多凭证——保留方密码=正常属性不拦;b 段 esports 缺口闭+方向 bug 修正:原无条件 `!other.Mergeable` 把保留方密码当重量);ctr 全页面族=11 路由+security.fn 层+merge-flow 内存一次性 offer(不可 URL 化)+i18n 113 键×2;**server fn 边界 toJson DTO 防腐**(枚举过界=proto 名字符串/WKT Timestamp=字符串,页端禁与数字枚举比对——走查网络面板实锤);走查(D-AQ 渲染态读原型 n31/32/33/34+block3/4):登录后绑定提示/security 三行/bind-phone 冲突卡+BLOCKED 说明 CTA/bind-wechat 三态/密码三路/merge 四场景+双闸 sheet+success;**场景一短信腿全链**(M1 富+M2 空:发码→合并→success→DB 三锚)+**场景一微信腿真人全链**(用户手机三次 OAuth:6005 冲突→fresh leg preview→KEEP_CALLER→执行→openid 迁 M1+W 注销+台账)+**KEEP_OTHER 真实渲染**(M4/M5 对:方向翻转+微信切换腿 CTA)+**BLOCKED 链**(M3 富化对:pane+soon toast);走查抓修 4 bug=①路由 leaf 吞子路由(merge/bind-wechat 无 Outlet 吞 success 与 callback——/profile/security 同课重演,改 layout+index)②clearMergeFlow 先清后导航=闪退竞态(后置)③BLOCKED 判定无触达入口(BlockedPane 死码→冲突卡 BLOCKED CTA+preview-callback 送进 merge 页;原型 34 右 pane=同页 seg 演示态无流转定义,入口 CTA 文案=设计自由度)④确认卡样式(外卡 padding 缺失+掩码行「手机号」前缀折行,原型 34 逐值对齐 16/12/10px);两裁定偏离=解绑走客服(白名单#2)/设置密码需输手机号;掩码观察=多凭证保留卡显示第二号非登录主号(策略记 gap);门禁=proto generate+breaking+check-drift 0/system+esports build/test 0(esports 修正后复跑)/ctr 三链 0×2 轮;净清=DB 6 用户族+2 台账行+redis 105 键+页面+栈;**d(10-30)**:微信 OAuth 注入缝补全——AuthorizeURL 硬编码官方域改 endpoint 覆写同走 base(Exchange 半边早已铺,config.yaml WECHAT_ENDPOINT 注释自证;conf.proto endpoint 注释语义扩展=authorize+token 双腿)+unit 覆写锚;微信登录路 ENT_E2E(httptest stand-in 双端点:authorize 302 回显 state+token 兑换→auto register 落 member 座;E6b 真人依赖课闭;#nosec G710 行内特例=mock 镜像微信回显契约,协议强制原语类)+合并反例 ENT_E2E esports 侧(双方重量→BLOCKED+空 other→KEEP_CALLER 双锚;fake SystemRP 立缝=被测面是本服务的 emptiness bar 与 verdict,system 自答归其家)+ctr Playwright identity-routes 三锚(SMS 路 redis 读码全链含向导落座/wechat 路 spec 内 18443 stand-in+UA MicroMessenger 过环境门+REST 预检决定 skip/占用号冲突卡=双方皆空→BLOCKED CTA+零发码断言;栈起法=APP_SMS_BACKEND=mock+APP_SMS_TEMPLATE_CODE 非空+APP_WECHAT_ENDPOINT=http://127.0.0.1:18443)+fullchain 回归修复(E6c-a 默认 tab 翻 code 后密码路未切 tab——既有回归本片首暴露)+D-W 收口(a0abb1a4 population 机制片已全落:provider 行 population 语义/人群行 boot ensure/登录惰性落座=存量迁移/guard 测试自含+组合缝 a8b9f0a4,零代码台账滞后消解)+散行收编核验(SMS 散行 10-28 对账役已销账);门禁=system+esports build 0/ENT_E2E 双家全量 0(GOFLAGS=-count=1 禁 Go 测试缓存——env 门控测试的缓存污染坑)/ctr check+typecheck 0/全套 e2e 5/5 含 fullchain 2.6m/lint+format 0;净清=探针硬删+软删行既有形状+redis 键+栈 |

| 微信昵称/头像回填(G-201,E6b 余面小片) | lzx | ✅ 10-30 | snsapi_userinfo 全链:authorize scope 升 userinfo+Exchange 拉 /sns/userinfo(失败容忍=profile nil 不炸登录);契约 WeChatProfile{nickname,avatar_url} 入 LoginByIdentity/BindCredential 两响应(breaking 绿;10-30 规范批改 WKT StringValue 全链);biz 透传 id.Profile(自绑幂等分支也带——pill 误点已绑 openid 优雅回填);ctr 昵称步预填(截断/剔除清洗按冻结契约 12 码点)+from 行+琥珀调整行+pill(微信环境+未带出;点击=BeginBindCredential 绑定流,openid 占用=6005 诚实拒绝 toast)+头像同笔落 profile.avatar_url(E2 预埋列,存储 lives here);systemCredential 不持久化(裁定:每次登录重 OAuth 重新带出;绑定用户中途离开且回流=罕见边缘,gap 登记触发器=「已绑用户重同步头像」产品面);identity-routes spec 扩锚(预填三断言+avatar 落库 DB 锚+pill 成功绑定路新 test);chrome 三态走查(微信登录预填全链含 DB 断言/sms+微信 UA pill→6005 服务端实锤/非微信 UA pill 不显示);门禁全绿+全套 e2e 6/6;G-206 已裁修复(改文案去承诺)/G-207 白名单②+头像显示面递延;**10-30 质量批随片(99c5bb06)**:真机回归 7 用例 6 族修复(SMS 幽灵码时序/park 三点清/双空合并放行 KEEP_CALLER/pending 文案错位×3+无 code 死页×3/G-206 必选/G-208 首刀)+G-208 浮层首刀(BottomConfirmSheet 单源:exit-confirm/logout/merge 三迁)+规范审查 20 条三波全修+i18n 前缀清算(105 键×136 引用,跨域 login_* 归零)+config 显式化(SMS backend 必填枚举+boot 双横幅+APP_ 前缀法写明)+MergePreviewer=kratos §1 verdict-port 法条 |
| P10 语言设置行(hzy 小片,D-AT) | hzy | 待派 | 现成 LocaleSelect 挂 profile 设置区一行;组件零改动,只加挂载点;三链绿+走查 |

## 开放遗留清单(防丢网;完成即删行)

| 项 | 归属 | 触发/时点 |
|---|---|---|
| 微信回调 next 丢失(redirect_uri 不能带 query,保 next 须 state 中转) | 后续微信链路片 | 设计时定(review A6) |
| conventions guard 正则精度修正回流 dev 兄弟服务(esports 侧已 AST 化:顶层声明 doc 判定,体内 var 上方叙事不再误报;组内 spec doc 存量登记下行) | lzx | dev 车道下次触碰同步 system 双子守卫 |
| 组内 spec doc 存量约 40 处(test 模块 Frozen 出处系+E5-b const 组 doc;旧文本守卫对 const 组内注释天然漏检,AST 升法照见后刻意不扩执法面) | lzx | 独立机械片按家法行尾化(触碰 test 模块时顺路或立片) |
| 协议 consent 版本记录契约 | lzx | H2 versions.ts 真源 GET /agreements/current |
| ~~mt-F1 重测耗尽死局~~ **已裁(D-AR,2026-10-28)**:提示文案出口——双出口文案「本款机会已用完,成绩已取最高 —— 返回阵容继续下一款,或切换类型开新的一场」 | ~~ljr~~→hzy | 改 test_intro_blocked zh/en 两键,小片落地 |
| skill_trajectory 30s 走查体感偏紧,操作时间不够 | lzx | 评估加时 |
| mt-D5 intro 按款文案:getTest.introduction 字段在但种子未喂(seed.go 注「metric-sheet data」),locale 方案待定;客户端 per-game 键 B 方案被否(单一事实源) | lzx | 随种子/国际化面(用户 10-26 裁 A=后置转此) |
| mt-D6 手机横屏真机未验(桌面 812×375 仿真预验绿,10-26:布局/控件锚位/目标合法) | hzy | 手机在手时顺手过一眼 |
| ~~D-W population 余件(provider 行 member_population/座位迁移/guard 注释清理)~~ **已收(10-30 E6c-d 核验)**:a0abb1a4 population 机制片已全落(provider 行 landing population 语义+人群行 boot ensure 声明制+登录进场惰性落座=存量迁移答案+guard 测试自含)+组合缝 a8b9f0a4;挂账=台账滞后非真欠款 |
| 走查墙 4:运营测试账号口径——pg2-probe 无 user 座位登不进 admin 端;种子超管=admin/admin123456 | 用户侧 | 即办(SQL 补座位或文档化) |
| age-required 逃生按钮(先看免费报告;原型有面,现仓无) | lzx | pay 客户端片(付费墙落地时) |
| ~~P11 已裁(D-AN:维持单轨 birth)~~ | — | 2026-10-25 关闭 |
| ~~ICP 备案号权威串~~ **已定稿+已落地(D-AO)**:代码四键(login/report × zh/en)已全为「浙ICP备2026028293号」带「备」版,与 D-AO 一致,无改键动作;仅剩上线前 beian.miit.gov.cn 复核一次(随上线清单) | ~~ljr~~→已落 | 关闭 |
| G-208 浮层族余面:首刀(exit-confirm/logout/merge)已收 10-30;余族=pay-confirm/match-switch-confirm/resume-remind/share-invite 待对应页面片 | lzx | 各页面片落地时 |
| 规范递延族(10-30 审查批):security→login 跨模块引(beginBindWechatFn/wechat.fn/inputClass 三点)+fetchProfile 降级零遥测(provider 无 log 注入链)+report/biz.go Private 节内部函数方法交错(存量) | lzx | 跨模块依赖清算批/下次触碰对应模块 |
| 行政件催办:经营类目补设+H5 支付权限申请(PAY2 阻塞件,建议即办)+APIv3 密钥 32 位核对(CP3 前必须)+走查墙 4 运营账号口径 | 用户侧 | 即办 |
| 版本检查页+app-version/内容块管理(coverage 缺口①;S4 终局对账低优) | 待排 | P13 已签(D-AS⑥):协议版本管理+compliance-config 随 E9 增量入窗;版本检查页/内容块随 M2 余组 |
| 走查注:admin 工作台 ERR·NETWORK(dev 栈缺数据源,非 esports 面) | lzx | pre 部署前查 |
| test 目录 List 单键 created_at 排序,同秒并列池序靠 PG 返回序(E4-d 观察) | lzx | 下次 schema 触碰带上 |
| ~~Nx 冒号目标名解析(ctr:dev:down 被 Nx 落到 dev target,曾以 devserver.cjs 直调兜底)~~ | — | 2026-10-28 关闭(devserver 退役,dev-up/down target 已删,手滑冒号形态从静默退化为显式报错) |
| admin system-policies params.ts PolicyWire 断言(wire seam 收窄法适用面;system 属地不落本车道) | dev 骨架通道 | 随 dev 车道下次触碰 |
| PAY1 B 档递延五项(10-26 对抗审查):①跨日退款对账盲区(账单 At=交易时间列,退款跨日不入窗)②公共回调面结算 grant 落 default 租户(多租户启用前须显式带 order 租户)③sku 丢 PeriodDays 时 settle 照付不 grant(sku 无写面,潜伏)④未知 serial 出站证书拉取无节流(公共面 1:1 放大)⑤e2e 对账用今天窗(生产=昨天窗约定未覆盖) | lzx | 各自触发器:①退款域设计②多租户面③sku 写面④节流面⑤PAY2 联调 |

## Review 遗留转交(2026-10-20 ctr 全项目 review;开放项已并入上表,此处仅存注记)

接受不改注记:B4 `user-scalable=no`(游戏 H5 防双击缩放,产品定位接受);D8 stores 空壳(D-T 能力配置面,机制待命)。

## 检查点登记(D-M)

| 检查点 | 日期 | 判定 | 拉闸动作 |
|---|---|---|---|
| CP1 | 11-08 | 待检(前瞻:全项实质达成——支付门 10-25 实测绿;余=E-line-b 回位) | — |
| CP2 | 11-22 | 待检(前瞻 10-30:全项实质达成——游戏 13 款≥8/E4/E8/E5 全族/E5b/PAY1;提前 23 天) | — |
| CP3 | 12-02 | 待检 | — |

## 里程碑验收记录

(空;收口时一行一注)

## 待裁项状态

见 README §2.2;分析进展记 prd.md 对应条目。

## 节奏记录(一行一事;证据与细节归 git log)

> 剪枝律执法复位(2026-10-28,M1 收口):历史节奏行(10-18..10-28,75 条)已压实为索引行迁 progress-archive.md;全文叙事=提交信息 proven-by 尾与收口报告。新条目按索引形写入本节(日期+切片+hash+一短语,≤150 字);里程碑收口时本节清空压实回归档,并做活档区漂移对账。
- 10-28 D-AR(ljr):mt-F1 重测耗尽=提示文案出口,双出口文案转 hzy 改双语键;重测经济不动(3ee06f51)。
- 10-28 E5b-0 报告簇设计定案+拆片(1792dffe):语料三裁(历史最佳百分位=read-time cohort/职业开放态=旧仓语义级/旧 A-B-C-D 档位废止);拆片 a 契约→b 认知趋势→c 职业→d ctr 四页→e 翻转 e2e 入 plan 行。
- 10-28 E5b-a 契约三读面落地(d619454e):认知/趋势/职业聚合入 ReportService(判定内嵌响应,趋势点带 genre,零新错误码);proto 链+esports build/test+ctr:build 真码全绿。
- 10-28 E5b-b 认知+趋势聚合落地(5a760a60):判定序上移 shared 单源双消费(B 裁定;AccessJudge UC→UC 违红线被叫停撤形);退化池黄金锚(八维全 50)+digest 百分位回填销遗留;ENT_E2E 活栈全绿。
- 10-28 E5b-c 职业聚合落地(e73f5553):四域门控库(旧仓 75/70 阈值+min 匹配,key 译至冻结矩阵拼写)+处方=最低 4 维 vs tierA 85 单源;双 helper 重构认知面同源;库事实借行业统计待 ljr;ENT_E2E 全绿
- 10-28 E5b-d ctr 四页(56e7f0a6):中心(组合 digest+双 access 四卡)/认知(双计数+雷达≥3 维+缺维不画)/趋势(genre 折线+进阶徽章)/职业(关闭态整版+库行);D-AQ 渲染态读原型,gap 二批 6 条,三链 0
- 10-28 E5b-e 翻转+零值锚 e2e(99b4032a):closed 行压 membership/兑换码压行/covered 公开/趋势 advanced 免疫;独立 cohort 造 gap=0 已达线行=protojson 零值回归线;partial 趋势 4 序列;净清 8 探针 0
- 10-28 E6c-0 账号域定案+拆片(10f67ab2):SMS 按 10-20 分层(capability 家族+auth provider+cache 验证码);发码新 RPC+LoginByIdentity 加 sms;BindCredential 补 state 系会话;合并=富吞贫+审计;拆 a-d
- 10-28 E6c-a SMS 能力+发码+登录(bf0a99a0):capability/sms 双后端(dysmsapi 手写签名+金向量/mock 可显式切)+conformance;发码 RPC 铸 state;LoginByIdentity 加 sms;烧码防爆破;ENT_E2E 全链绿
- 10-28 E6c-a 续 ctr SMS 登录接入+双轮拟真验收(bf0a99a0):登录页 SMS tab 转真(hook/ui 拆分);mock 轮+Aliyun 真发码轮 Chrome 拟人全链均绿落 member 净清;热修=infra reason 静默/倒计时换号清窗/APP_ 前缀
- 10-29 E6c-b 绑定+合并流(2e0baf09):BindCredential 加 sms+state 系发起会话(BUILTIN 孪生面铸绑定态);MergeAccounts 富吞贫一次性+台账;空号判据 system 硬闸本片+esports offer 面随 c;ENT_E2E 六锚+live 探针(redis 绑定态实证)
- 10-28 D-AS(ljr):P13 六项签毕——消息中心+反馈/分享裂变面召回入窗(W7-W8 落点,梯①减员),scheme-b 随 M3,旧 rank/pk·upload 废弃,学校报告随 P14(默认锁+后台手动解锁),admin 折中(协议版本管理+compliance 随 E9 增量,报告配置面梯①)。(d32ba170)
- 10-30 E6c-c ctr 账号安全客户端面+合并 offer 面(dd83d3d6):system 三 RPC(ResolveCredential/ResetPassword/GetSecurityStatus)+esports PreviewMerge 五态 offer;方向感知重量闸(接口倒置住 report);ctr 11 路由+fn 层 DTO 防腐(枚举过界=字符串);四场景全态走查含微信真人三次 OAuth 合并全链;抓修 4 bug(路由吞子页/闪退竞态/BLOCKED 无入口/确认卡样式);门禁全绿+净清
- 10-30 E6c-d 三路 e2e+微信 OAuth 注入缝+D-W 收口:E6c 全簇毕。AuthorizeURL 覆写(缝双半边齐)+微信路/合并反例 ENT_E2E+identity-routes Playwright 三锚(SMS/wechat/冲突卡 BLOCKED);fullchain tab 回归修复;GOFLAGS=-count=1 防 Go 测试缓存污染 env 门控;D-W 零代码收口(a0abb1a4 早已全落);门禁全绿+全套 e2e 5/5+净清
- 10-30 D-AT(ljr):P10 定案 语言切换=个人中心设置行(现成 LocaleSelect 挂设置区,hzy 小片);P10 关闭。
- 10-30 D-AU(ljr):P7 四参数定案(提醒可配置/到期回双选项/成功态复用 27 结构+标题双键/AI 出稿);视觉稿 12-02 前定稿交 PAY3。
- 10-30 P7 视觉稿出稿+用户确认(ljr,AI 出稿三态 HTML 落库 p7-member-paywall.html;续费态=绿 CTA/单 CTA/状态头到期日/详情出口);待交 PAY3 实现。
- 10-30 微信昵称/头像回填(G-201 修复,E6b 余面小片):scope 升 snsapi_userinfo+userinfo 腿(失败容忍);WeChatProfile 入两响应;ctr 预填+pill(绑定流复用,占用=6005 诚实拒绝)+头像同笔落库;identity-routes 扩锚+pill 新 test(4/4)+全套 6/6;chrome 三态走查含 DB 断言;G-206/G-207 登记
- 10-30 微信全链真机回归(用户协作,7 用例全过):R1 登录预填真链路(thirdwx.qlogo.cn 头像落库铁证)/R2 二登进绑定页=设计内/R3 授权确认页正面+取消不回跳/R4 pill 6005 拒绝/R5-6 合并全链(双凭证归属+openid 转移 DB 铁证)/R7 pill 成功闭环。抓修 6 族:SMS 幽灵码+窗口死锁+模板空兜底(服务端先写后拨改「先轮换→拨→失败烧」+强断言)/park 残影跨账号身份泄漏(登出+消费+拒绑三点清)/双空合并误 BLOCKED+文案撒谎(放行 KEEP_CALLER)/「正在登录」文案错位三处+无 code 死页三处(同构修)/G-206 性别文案改「必选」(用户裁定 sex 不回填)/G-208 登出 sheet 六处差(浮层族批首条)。另立 config 显式化工单方向(用户裁定:空值禁静默兜底链);净清三库归零+三栈停
- 10-30 浮层样式对齐+规范大审查(用户裁定「都按建议处理」):第一刀 BottomConfirmSheet 单源化(390px/上圆角 18px/三色族 primary·hot·ghost,红色只留真危险;登出=主 CTA/退出测评主按钮=继续测试),exit-confirm 重写+logout-sheet 薄壳+merge sheet 迁移,原型逐字文案+chrome 双侧 7 项几何实测全同,G-208 收口;第二刀双 agent 全读审查 20 条实证(工作树 8+E6c 三笔 12),三波全修:波 1 小项(config 镜像/bind_wechat_pending 键名/token×2/wechat-profile 迁 features/app)+波 2 proto WKT 全链(WeChatProfile→StringValue+wrapperspb+Private 节序+e2e 断言改形)+波 3 E6c 批(MergeUC 归位 UC/Method/Private 三节+log tag 去 report_merge/ProviderSet 分层序/profile 值类型归 DO/client const 上移跨节区/bind+merge 回调一文件多组件拆 5 文件+pending-landing 独立/注释清洗 11 处去切片号中文场景标/proto 响应标量 carve-out justification 6 处)+B 批 i18n 清算(security 域 105 键改 security_ 前缀×2 locale×136 引用点+footer 三键迁 common+SMS/错误族 security 域自持 12 键,跨域 login_* 引用归零,paraglide 强制重生成);门禁 proto 三链+esports/system build/test+ctr build/typecheck/check+admin typecheck 全绿;递延=MergePreviewer 接口倒置架构裁定(材料已备)+fetchProfile 遥测+beginBindWechatFn 跨引族
- 10-30 MergePreviewer 架构裁定落定(用户裁定选项 A,判例升格):kratos §1 UC→UC 禁令补「verdict port」carve-out 法条——跨聚合裁决(调用方持 RPC 面+事实在对方模块)唯一合法形=消费方 biz 定义接口+提供方 UC 实现(wire.Bind),与 repo port 同向(消费方持接口,适配方满足),DAG 无环由「仅提供方 import 消费方」保证;E6c-c 形状由此合法化,无代码改动
- 10-30 config 显式化切片(方案 B,用户裁定「按你建议继续处理」):SMS 通道拆自动推断——backend 必填显式枚举(dysmsapi 需凭据对|mock 只记录;空/未知/半凭据 boot 即炸,错误带字段级 missing/set 诊断),按凭据自动分支删除;boot 解析横幅两落点(sms channel resolved: provider/backend/mode + auth provider resolved: key/kind/enabled/auto_register/endpoint official-or-override),裸名 env 陷阱实证并显式化(yaml 占位符 ${VAR} 只被 APP_VAR 覆盖——env source 只认 APP_ 前缀,SMS_BACKEND=typo 实测静默忽略;configs/ 两服务头注释写明规则,container/ 已有;deploy compose 全 APP_ 前缀核实干净);wechat endpoint 空=官方基址语义合规不动(e2e 注入缝依赖);测试表翻转锁新语义(空=炸×2+半凭据=炸);真机四验:正例横幅三行+服务三 transport 就绪+typo 炸+dysmsapi 缺凭据炸;门禁 proto 三链+system/esports build/test+go:lint/format 全绿
