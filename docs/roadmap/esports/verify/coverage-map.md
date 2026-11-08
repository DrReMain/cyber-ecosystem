# 覆盖映射 — 需求→用例→代码 三方对照与缺口清单(2026-11-07 重构役三)

> 建立:2026-11-07,八域并行只读核验(D1 登录账号/D2 向导画像/D3 匹配测评/D4 报告族/D5 支付会员/D6 激活码权益/D7 消息反馈/D8 admin 运营面)+主会话复核与分母机核。
> 三层分母(全机提):需求=85 页原型×状态变体(原型源码+_state.js 实读)+blocks 补充面+已裁口径;代码=esports 85 RPC 逐方法+system 交叉(auth 14/user 3)+ctr 50 路由(1 root+9 布局壳+40 叶页,渲染门模型)+admin esports 17 面;用例=integration-matrix(设计层)。
> 判定规则:每页→≥1 用例 id 或显式处置(未建→片/窗外→D 号/底座→D-AY);每 RPC→≥1 用例触达、e2e/单测承载或显式豁免;每用例→预期锚可追溯。
> 消费:后续多轮走查按本档缺口清单逐轮收割,**修毕即删行**;计数随收口再生。

## 一) 总判定(2026-11-07)

| 分母 | 判定 | 计数(自证) |
|---|---|---|
| 需求·85 页 | **全归属零悬空** | 已建 45(ctr 29+admin 16)逐页有用例映射;未建/窗外/底座 40 全显式处置(demand-ledger §一/§二) |
| 代码·85 RPC | live 走查触达 **79** / e2e-only 承载 **5** / 零消费+零测试 **1**(ListOrders) | 79+5+1=85 ✓(R2 三批后:零执行 3 与单测-only 1 全转 live,e2e-only 11→5) |
| admin 17 面 | 逐面有用例(A 套+X 交叉腿) | 0 悬空(D8 逐页核) |
| 用例锚 | 全量核对:命中为主;**失真 9 处已随矩阵重写修正**(见 §四) | — |

e2e-only 承载 5 清单(R2 后):GetTestAttempt(测评 tab 只读面)/ListBanners/MarkAllMessagesRead/NotifyPayment(真通道刀)/PreviewMerge(merge_preview_e2e)。R2 已转 live:SetPayAgeLine/SetCompliancePrompts/SetRenewalReminder/UpdateAnnouncement/OfflineAnnouncement(A4-3+X2-2)/RefundPaymentOrder/TerminateMembership(A3-2/X1-4)/SwitchMatchGenre(单测→双承载)/GetTestConfig+UpdateTestConfig(A5-2 调参 modal)/SetReportAccessPolicy(C2-4 翻面)。
豁免台账(防误判):MockCashierService=dev-only(契约自证);MarkAllMessagesRead=原型无 oracle 面+e2e;ListBanners=消费面未建(遗留清单在案);NotifyPayment=真通道刀(微信真签名);GetWechatOpenId=pay 域下单富集(ctr 零消费);UpdateProfile=向导域承载;GetCurrentUser/system UserService 写=会话底座/admin 车道。

## 二) 走查缺口清单(收割优先序;修毕删行)

### P0 引擎与 MOBA 侧(结构性零触达)——**R2 轮 2026-11-07 全数收割**(毕证:matrix C1-7/C6-2/C1-4 行注)
8 款引擎首次活体执行+明细 schema 逐款收证;SwitchMatchGenre 真执行多轮;切换进度警告(done>0)/同款跨场再现/机会用尽置灰+mt-F1 双出口/须知二次面全 live;附带=Q1-⑥ 介绍文种子三款 live 首证+done 页三态变体补全+DB 双场互证(red_dot 0/C 判档,curvePending 全 raw-only)。

### P1 admin 写腿(代码在,零执行)——**R2 轮 2026-11-07 全数收割**(毕证:matrix X2-2/A4-3/A3-2/X1-1..X1-4 行注;日志双侧互证)
首批六写腿:SetRenewalReminder(7→14→7 往返+args=days:14 日志)、UpdateAnnouncement 编辑、OfflineAnnouncement 下线、公告查看 modal、SetPayAgeLine(18→19→18)、SetCompliancePrompts(minor_prompt 覆盖写+空串 SQL 净还原)。造态链批:P1-3 RefundPaymentOrder UI 退款腿(原因必填 Modal→受理 toast→RefundFacts 收敛;`order refunded by operator state=succeeded` 日志)+P1-4 TerminateMembership 真执行(detail Modal→头卡翻已终止;`membership terminated by operator` 日志),链承载=X1-2 mock 购单(Pdbn425tthm89nc9n3q2g)→X1-1 授 31d→X1-3 续费砌缝(顺带收)→A3-2 退款→X1-4 终止→ctr 回 pitch 态,全链零红。

### P2 C 例族——**R2 轮 2026-11-07 全数收官**(毕证:matrix C 节全行注;C1-3/C1-6 例外在案)
C1 冷启动=R1 承接在案;C2 报告族/C3 支付面/C4 账号族/C5 消息反馈/C6 变体扫尾全毕(C6-2 前批毕)。**发现入账**:B-10 高优(手机号+密码登录假承诺,plan 工单)+反馈 10 下界静默拦截(P4 轻缺,N-5 族)+UpdateProfile/POLICY 双「客户端拦截纵深防御形」判层。
**余=造态依赖(P3 批)**:C2-2 解锁态+C2-3 两点态/混类/N-1(cohort 造态)+C3-1 会员隐单次/X1-5 未成年红卡(R-minor)+C4-2 KEEP_*/C4-4 可合并(R-empty)+C4-1 NoPhoneCard+C4-6 双凭证(真微信)+C6-4 断网暖卡+P3-7 服务端直探族。

### P3 造态依赖与专项腿——R3 轮 2026-11-08 收割 P3-1/P3-2/P3-4 全删+P3-3/P3-7 主体(毕证:matrix C1-2/X1-5/C3-1/C4-1/C4-2/C4-4/C6-6 行注;wt-* 四号净清零残留)

**R4 轮 2026-11-08(同日第二场)**:X3 激活码全生命周期五例全收(X3-1..X3-5,拒绝序 INVALID→族幂等→EXPIRED→DISABLED 实证)+X2-1 反馈六腿全闭环+X2-3 半(feedback_reply 形+admin 信箱五列表毕,report/纯公告 CTA 形欠)+A2-1 详情六 Tab 顺路;毕证=matrix X2-1/X2-3/X3 族/A2-1 行注;wt-rich3-306/wt-fb-307 两号+批次 B-2611-HS8U+过期造态批+两票反馈+扇出信端到端净清零残留。走查坑三枚回灌 CONVENTIONS(§1.5 造号 id 须 20 字符 xid 形/§1.6 ant 表固定列克隆行/§2.3 跨端写后读须 reload)。
| # | 缺口 | 依据 |
|---|---|---|
| P3-3 余 | WechatCallback DISABLED 两支(换账号钮+密码 reason 词表族已毕:R2 词表+R3 LOGIN_SEAT_REQUIRED 红行/更换账号钮,零 R-multi 依赖=admin 号即可造态) | C4-1 行注 |
| P3-5 | 真微信腿族:登录授权腿/merge KEEP_OTHER 换边+切换尾页/bind 6005 竞态——随真通道窗标注 | C4-3/C4-4 例注 |
| P3-6 | pick 空池 why=empty 活体/切后台记无效面/较上局 delta 芯 | C1-4/C1-5/C6-2 补步 |
| P3-7 余 | GetMatchSession no-drawable 拒答造态支(UpdateProfile 6100+GENRE_UNDECIDED/GENRE_UNKNOWN 已直探毕@R3) | C6-6 行注 |

### P4 轻缺与变体(顺路收割)
昵称 emoji(**毕@R2**:「昵称暂不支持 emoji 表情」行内)/偏好重复 toast·满 3 态·平票字面(**毕@R3**:「已经添加过了」toast+「已添加 3 / 3」添加钮 disabled+「类型投票 FPS 1 / MOBA 1 · 默认类型：平票 → 需手动选类型」)/出生不完整拒(age.tsx complete=false 行内分支,R3 未走维持)/13-18 黄条(**毕@R3**:「未满 18 周岁——可以测评，但无法购买付费报告/会员（付费时会再次拦截）」无 checkbox)/profile 总览九区无主走例/digest 空兜底 toast/center covered≥6 full 支/趋势进阶 CTA(见 §三)/admin 码表(**复制未核销码毕@R4**:toast「已复制 4 个未核销码」;两筛 combobox 在场+分页汇总行毕@R4;筛选逐一走+空态=R5)/admin 反馈详情 closed 送框移除/公告时间列(§三)/X2-2 home 红点句复走证(毕@R2)/断网暖卡。反馈 10 下界=R2 入账后 11-08 复核**翻案销案**(行内红行「描述至少 10 个字,便于我们定位问题」真实渲染+SubmitFeedback 未发——R2「零反馈面」系 toast 单观测漏持久红行的误判,§2.2 立法实证案例)。

### R3 即修记录(2026-11-08 走查中发现,面级当场修毕;非偏离不入 §三)
①**D-AN⑤ 落地漏**:「命中偏好的能力池内权重 ×1.5」句 D-AN(10-25)已裁弃但两 locale 文案仍在+live 渲染在场——两 locale 删尾子句,HMR 上页亲证;②**merge 成功页文案错指**:security_merge_success_body/line_credential 静态写死「微信」按微信冲突升级场景而作,keepCallerSms(手机号吸收空号)腿下失真——四键两 locale 改中性形(「全新账号已注销 · 其登录方式已并入当前账号」/「登录方式已并入当前账号(手机号/微信均可用于登录)」),HMR 上页亲证。附带判层:SessionCounts 按 test_record 口径(仅 done 场 match_session 不计入),两空号对走 absorb-other 支方向不变。

## 三) 新登记偏离(本次核验发现,未入任何台账;全数转 prd 待裁区 R 系)
| # | 偏离 | 依据 |
|---|---|---|
| N-1 | 趋势页进阶分析模块+会员 CTA 缺失:原型 report-trend.html:72-84 有锁卡+开通 CTA;重建仅 advanced chip 无 CTA 模块/词条;Q2-E2 校对点明列但修复未含=审计漏项 | D4;trend-page.tsx;zh-CN.json:113-114 |
| N-2 | admin 码表搜索窄化:原型「搜索激活码/核销人」→实现 exact-only 码搜索 | D6-1;detail.html:57 |
| N-3 | profile 四格「进步次数」恒「—」(契约无序列读面) | D2;overview.ts:102-107 |
| N-4 | 公告列表时间列=发布时刻(草稿显-)vs 原型独立「创建」列 | D7(a) |
| N-5 | 空输入拒绝形态:原型 toast→现行禁用钮/antd 行内校验(ctr 追问框/admin 快捷回复/详情回复三处) | D7(b) |
| N-6 | age-required 保存后落 "/" vs 原型回 pay-confirm(UNVERIFIED,复走核) | D5;age-required/page.tsx:72 |

## 四) 锚与记录失真清单(矩阵重写已修,存档备查)
①C4-1 锚「服务端组成执法=lzx 队列在册」→已落 88ada92a(auth/biz.go:460);②C5-1 锚 contact 128→契约 128/输入 40(PM-14);③C2-5 锚「⑤裁」→Q2-⑦;④C2-1 锚 R-04 误并引(硬价=R-05);⑤C3-4 锚落点精确化(profile ③ 卡+admin 列);⑥X1-2 态列「未走」与 gap-ledger 11-07 活走矛盾(态列已清零重立);⑦X2-2「home 无消息入口徽」句与代码+demand-ledger 26 矛盾(删句,C5-2 复走证);⑧C1-2 步骤「≥3 款」残留→「至少 1 款」;⑨P11「在册票」措辞滞后于 D-AN 关闭。

## 五) 需求侧已证映射面(防重查;详证=matrix 例行+demand-ledger 行)
- ctr 34 已映射页全有例/锚;admin 16 已建页全有例(D8 15/15 锚核);85 RPC 账目见 §一。
- 清洁面实证:SKU 种子价=P1 冻结(seed.go:201-207)/SkuAdmin 四 RPC 走查+e2e 双覆盖/D-AB 判决序单序(kernel.go:88-115)+e2e 全臂/entitle 表模型六语义 e2e 承载(TakeUse 单语句/族幂等/ratchet/DoNothing/停用围栏/唯一键)/msg 19 RPC 服务端全实现+拒绝词表齐/error_reason 全域齐。
- 真源分母(S1 机核):旧 191 端点/36 业务表/原型 192(ctr)+41(admin),tsv 锚零差集;旧前端口径=ctr-game 65 tsx/admin-game 26 jsx+25 router path。
