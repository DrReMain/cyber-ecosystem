# esports 进度档 — 认领状态与验收记录

> 归属:esports 前沿进度档,三人共写。**剪枝律(2026-10-23 立,2026-10-28 补执法细则,约束三个 agent;全文见 README §5)**:片收口即剪枝为一行收口注(状态+日期+提交 hash+关键判定);节奏行=索引形(日期+切片+hash+一短语,≤150 字),细节归提交信息 proven-by 尾;活档区(快照/认领表/遗留清单)为唯一可变区,收口时重写而非追加;历史节奏行随里程碑收口压实迁 progress-archive.md;里程碑收口做一次活档漂移对账——档是给 agent 读的,读档成本即团队成本。检查点每站在下表登记。

## 现状快照(2026-10-28,W2 中段)

| 人 | 已交付(均已并 esports/main) | 当前 → 下一片 |
|---|---|---|
| lzx | E0-E2、E3 评分(13 款收满,71c33f4f;§指针清扫 43fd4b3d)、E4 匹配全量(704c3388/07b87349)、E6a/E6b、E7-a..e 13 款引擎收满(fa866414/dcce4fc6/1ce16289)、V1-V4+B 批(6b98c886/865511c1)、会员读面+admin 用户管理(88bcd9ba/3f272b28)、组合缝(a8b9f0a4)、校准战役 S1-S4+D-AL(9e0478e6/fe73a63a/1c1a8424)、支付 PAY-C0 四片+PAY-C1 四片+PAY1 五片(CP1 支付门 10-25 提前达成;APIv3 密钥 33 字符待核=回调解密/CP3 前必须修)、流转批三件(ICP 备字+种子国服重播;偏好上限终态=常量写死,10-26 二裁推翻 config 化)、E8 权益三片(10-25 收满,含架构审查收口)、E5-0 定案+a/b1/b/c(10-26,二轮审查 15 项修复批同日)、MR!9 吸收+wire-stamp 修复笔(e62ce8c6,10-27)、E5-d+E5-f 读回面族(E5-d 69721f93/f 族 0ae81a28,10-28)、E9 最小 admin 面(c1ca75d3+5895f1a8,10-28)、D-AQ 设计对齐前置法立档(abacc542)、E10 全链 e2e(login 九红条消化+Playwright 全链,10-28) | **M1 全片收满**→**▶ 下一片待定**(候选:E5b 已解锁/E6c/PAY2 行政件就绪后/MIG1 W6 起) |
| hzy | H1/H2/H-A/H-B、10-22 精修批、10-23 H-C 判定回路(真人链路验收;撞面合流正身=E4-c 版)、10-23/24 设计保真批(MR!6 于 10-25 并入 main,599dddce)、10-26 走查修批九条(93f67de3..179b1af2,见归档;vitest 112→115)、10-27 验收收口(tm/home-D5 亲验)+性别卡(6beff1c1,销 D-AN 流转半边)+走查下午批(MR!9 已并 main) | **下一片:E-line-b ctr 测试回位**(10-28 派定,18 件两家园直回,机械面;顺手项平均决策由 MR 内 12dc76a5 销) |
| ljr | P1 计费(D-AB/AC)、P3 完成页(D-AI)、P9 能力归属(D-AD)、D-AG/AH 同批、P15 报告簇(D-AP,10-28) | 待裁:**P13 窗外豁免签认(CP2 前)**、P16 文案键半边(待标定款回执)、P4(11 月底;HL 问卷回复红线 11-15)、P7(12-02)、P10(随 P4)、P2/P5/P6;~~P15/P8/P11/P12 已裁(D-AN/D-AM/D-AP,10-28 刷新)~~ |

**M1 燃尽**:E0✅ E1✅ E2✅ E6a/E6b✅;E3 ✅(13 款收满,报告分 e2e 挂 E5);E4 ✅(a-d,e2e 三验收);E7 ✅(13 款收满);E8 ✅(三片收满,10-25,含架构审查两裁落地);E5 ✅ 全族收满(E5-0..f3,10-26..28;E5-e 溶片=锚并 E-line-a/E10,10-28 拍板);E9 ✅(10-28)+E10 ✅(10-28);M2 用户列表+档案首面已前拉;V1-V4+B 批落,admin 对齐批待排;支付 PAY-C0✅+PAY-C1✅+PAY1✅五片(**CP1 门 10-25 提前 14 天实测达成**:真渠道验签握手绿);PAY2-3 未启动(H5 闸实测红=产品权限未开通,行政件)。**CP1(11-08)前瞻:门全项实质达成;余=E-line-b 回位(hzy 进行中)。**

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
| MIG1 迁移工具片 | — | 未认领 | 2026-10-23 终局对账立片;W6-W7 建/W8 预演;school 步时序=P14;依赖 E8/PAY1/E5 |
| E5b 报告簇扩片 | — | 未认领 | 2026-10-23 终局对账立片;P15 裁后开工;8 维公式源已锚 plan 行 |
| E6c 账号域收口 | — | 未认领 | 2026-10-23 终局对账立片;SMS+密码+绑定合并;收编 SMS 散遗留行;**10-25 联动注:SMS=手机浏览器支付腿(H5)前置,openid 绑定流=微信内 JSAPI 凭据供给,E6c 优先级与 PAY 联动;微信 OAuth transport 注入缝随本片** |

## 开放遗留清单(防丢网;完成即删行)

| 项 | 归属 | 触发/时点 |
|---|---|---|
| 微信回调 next 丢失(redirect_uri 不能带 query,保 next 须 state 中转) | 后续微信链路片 | 设计时定(review A6) |
| conventions guard 正则精度修正回流 dev 兄弟服务(esports 侧已 AST 化:顶层声明 doc 判定,体内 var 上方叙事不再误报;组内 spec doc 存量登记下行) | lzx | dev 车道下次触碰同步 system 双子守卫 |
| 组内 spec doc 存量约 40 处(test 模块 Frozen 出处系+E5-b const 组 doc;旧文本守卫对 const 组内注释天然漏检,AST 升法照见后刻意不扩执法面) | lzx | 独立机械片按家法行尾化(触碰 test 模块时顺路或立片) |
| SMS 骨架片(system provider+capability) | lzx | E6c(2026-10-23 立片收口) |
| 协议 consent 版本记录契约 | lzx | H2 versions.ts 真源 GET /agreements/current |
| **mt-F1 重测耗尽死局**:match 阵容步「剩 0 次」进 intro,「继续 · 开始测试」被花费闸拦死——无提示、无出路(不能重测/跳过/换);活体证据在 10-26 走查会话档 | ljr | 裁出口三选一:提示文案/留观/恢复一次 |
| skill_trajectory 30s 走查体感偏紧,操作时间不够 | lzx | 评估加时 |
| mt-D5 intro 按款文案:getTest.introduction 字段在但种子未喂(seed.go 注「metric-sheet data」),locale 方案待定;客户端 per-game 键 B 方案被否(单一事实源) | lzx | 随种子/国际化面(用户 10-26 裁 A=后置转此) |
| mt-D6 手机横屏真机未验(桌面 812×375 仿真预验绿,10-26:布局/控件锚位/目标合法) | hzy | 手机在手时顺手过一眼 |
| D-W population 余件(provider 行 member_population/座位迁移/guard 注释清理) | lzx | E6c 账号域收口(原锚 E8 已收未动,10-28 重定) |
| 走查墙 4:运营测试账号口径——pg2-probe 无 user 座位登不进 admin 端;种子超管=admin/admin123456 | 用户侧 | 即办(SQL 补座位或文档化) |
| digest 六能力 best_percentile 恒缺席(f3 换源后 AbilityBests 不算百分位=诚实 null;薄 cohort 下 A-② free 座空缺同源) | lzx | cohort 常模服务化立片 |
| age-required 逃生按钮(先看免费报告;原型有面,现仓无) | lzx | pay 客户端片(付费墙落地时) |
| ~~P11 已裁(D-AN:维持单轨 birth)~~ | — | 2026-10-25 关闭 |
| ICP 备案号权威串(LOGIC_AUDIT 2.5b「浙ICP2026028293号」vs MEMBERSHIP §九「浙ICP备2026028293号」) | ljr | 随 ljr 下轮文案定稿 |
| 微信登录回填昵称/头像(2.5b-⑫;需 system identity 持久化+向导去向联动) | lzx | E6b 余面小片(S1 登记) |
| 版本检查页+app-version/内容块管理(coverage 缺口①;S4 终局对账低优) | 待排 | P13 签认后随 M2 余组/E9 增量 |
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
| CP2 | 11-22 | 待检 | — |
| CP3 | 12-02 | 待检 | — |

## 里程碑验收记录

(空;收口时一行一注)

## 待裁项状态

见 README §2.2;分析进展记 prd.md 对应条目。

## 节奏记录(一行一事;证据与细节归 git log)

> 剪枝律执法复位(2026-10-28,M1 收口):历史节奏行(10-18..10-28,75 条)已压实为索引行迁 progress-archive.md;全文叙事=提交信息 proven-by 尾与收口报告。新条目按索引形写入本节(日期+切片+hash+一短语,≤150 字);里程碑收口时本节清空压实回归档,并做活档区漂移对账。
