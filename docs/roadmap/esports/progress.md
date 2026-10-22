# esports 进度档 — 认领状态与验收记录

> 归属:esports 前沿进度档,三人共写。**剪枝律(2026-10-23 重申,约束三个 agent)**:片收口即剪枝为一行收口注(状态+日期+提交 hash+关键判定);节奏行一行一事;验证证据一律归提交信息的 proven-by 尾,本档不存叙述——档是给 agent 读的,读档成本即团队成本。长历史随剪枝删除,git log 即存档。检查点每站在下表登记。

## 现状快照(2026-10-23,W1 末)

| 人 | 已交付(均已并 esports/main) | 当前 → 下一片 |
|---|---|---|
| lzx | E0 契约(种子+增量2/3/4)、E1 骨架、E2 schema+种子、E3 评分(框架+Tier A 六款+提交缝,ce763107)、E6a/E6b、会员读面+admin 用户管理(88bcd9ba/3f272b28)、system 组合缝 ListUsersByPopulation(a8b9f0a4) | **▶ 视觉对齐批 V1(装饰件基座)→ V2(登录+画像)→ V3(编排+个人中心)**;后排:Tier A 余款四款(W3,首片 number_stars 补 CP1「过半」)→ E4(W4)→ PAY1+E8(W5) |
| hzy | H1 系统态件、H2 协议页、H-A 引擎框架、H-B 反应瞄准批(四款引擎,mock 净拔除)、10-22 精修批(控制台三红根治/profile 写缝/D-AG 落地/C3) | **H-C 判定回路闭环(已分配 10-23,D-AJ 首例,冒烟唯一硬断点)** → E7 追踪技能批(W2-3 起);挂起:D-AG/C3 实机走查随下次登录 |
| ljr | P1 计费(D-AB/AC)、P3 完成页(D-AI)、P9 能力归属(D-AD)、D-AG/AH 同批 | 待裁:**性别字段(原型必选 vs 契约无,W2)**、P8(E5 前,W4 末;含协议正文联动)、P4(11 月底;HL 问卷回复红线 11-15)、P7(12-02)、P10(随 P4)、P2/P5/P6 |

**M1 燃尽**:E0✅ E1✅ E2✅ E6a/E6b✅;E3 ⏳(Tier A 6/10);E7 引擎 4/13(判定回路待 H-C);E4/E5/E8/E9/E10 未认领;M2 用户列表+档案首面已前拉落地;支付 PAY1-3 全未启动。**CP1(11-08)前瞻:达标面基本在手,风险面=支付沙箱零启动。**

## 切片认领表(窗口与定义见 plan.md)

| 片 | 认领人 | 状态 | 收口注 |
|---|---|---|---|
| E0 契约首刀 | lzx | ✅ 10-21 | 种子(画像+目录)+增量2(测试读面)+增量3(统一提交,4 决策经裁)+增量4(三计数反哺,193a9b7c);余面已裁去向:报告读面随 E5/E9、entitlement 随 E8(D-AA) |
| E1 服务骨架 | lzx | ✅ 10-18 | kratos-new-service 八步;双面 accessLadder+catalog+部署面;会员链探针 e2e 绿 |
| E2 schema+种子 | lzx | ✅ 10-20 | D-V 撤回后按冻结契约 clean-room 重建;五表+三读+13 款种子骨架(校准值不杜撰,随 E3 单源);multi_target 补行=D-Z |
| E3 评分引擎 | lzx | ⏳ | 增量a metric.md 单源档、增量b Registry+纯函数+20 黄金 ✅;增量c 提交缝(4c5f0617)、增量d MT/TR 定稿(5b0160fb)✅,双轴评审两实错已吸收;增量e DZ/ST 定稿 ✅(净钟+哨兵双射闸,metric §6.7/6.8);**余:Tier A 四款(number_stars/offense_defense/resource_match/character_counter)+Tier B 三款闸门(moving_target/tracking_moving/economic_sim)**;「报告分 e2e」验收面挂 E5,H-C 前移回执层;guard 正则修正随下次镜像批回流 dev 兄弟服务 |
| E4 匹配引擎 | — | 未认领 | 重测 3/1 真律与难度模式轴归属地 |
| E5 报告框架+能力报告 | — | 未认领 | 服务/客户端两半可并行 |
| E6a ctr 纯壳 | lzx | ✅ 10-18 | D-R/S/T 全落;SEO 全量面同日收口 |
| E6b 登录+画像流 | lzx | ✅ 10-20 | 接通片(微信真人链路绿)+完整形态(登录/三步向导/个人中心/首页)+机制修复+R1R2+流程前段四页;遗留契约缺口随对应片:SMS 骨架(system provider)、consent 版本记录、自定义游戏上限 3v20 口径 |
| E7 游戏框架+13 款 | — | ⏳(引擎先行) | 反应瞄准批四款已由 H-A/H-B 前拉交付,批片余面=H-C 后 live 冒烟;追踪技能批/决策资源批未动 |
| E8 权益最小面 | — | 未认领 | D-W population 余件触发器=E8 |
| E9 最小 admin 面 | — | 未认领 | 用户管理子面已另落(下行);本片=激活码+测试项/配置管理+报告策略 |
| E10 全链 e2e | — | 未认领 | 全部依赖收口 |
| 会员读面+admin 用户管理 | lzx | ✅ 10-20 | MemberService 读面(88bcd9ba)+admin esports-users 列表/详情 Tab 骨架(3f272b28)+ListUsersByPopulation 组合缝(a8b9f0a4,D-X 落位);裁:一面一消费等级,BUILTIN 会员面不作运营消费面;详情余 Tab 随 E2/E5/E8 挂内容 |
| H1-H2(系统态件/协议页) | hzy | ✅ 10-20 | MR!2 双轴评审合并;EN 协议译文 ljr 定稿后整目录换值 |
| H-A 引擎框架片 | hzy | ✅ 10-20 | GameClock+GameModule 契约+TestRunHost+/test/$code+red_dot;biome 插件契约执法 |
| H-B 反应瞄准批 | hzy | ✅ 10-21 | 三款引擎语义锚钉死+本地结果面;mock 插桩净拔除零残留;E0 增量4 反哺闭环 |
| H-C 判定回路闭环 | hzy | ▶ 已分配 10-23 | D-AJ 首例;提交缝适配器+服务端判分回执;验收=真人全链(登录→/match→red_dot→分数回执) |
| PAY1-3 支付片 | — | 未认领 | P1 已定(D-AB/AC);商户资质就绪;遗留=经营类目补设+H5 支付权限申请(建议即办);**CP1 门含支付沙箱通=当前风险项** |

## 开放遗留清单(防丢网;完成即删行)

| 项 | 归属 | 触发/时点 |
|---|---|---|
| 微信回调 next 丢失(redirect_uri 不能带 query,保 next 须 state 中转) | 后续微信链路片 | 设计时定(review A6) |
| favicon 404(ctr 无 public/ 目录)资产口径 | hzy | 待裁,随手片 |
| birth.ts 上收共享层(原拟随 H-B,未随落) | 触文件者 | 顺手清 |
| conventions guard 正则精度修正回流 dev 兄弟服务 | lzx | 下次定期镜像批 |
| D-AG/C3 实机视觉走查 | hzy+用户 | 下次微信登录同场 |
| SMS 骨架片(system provider+capability) | lzx | E6b 余面,排期待定 |
| 协议 consent 版本记录契约 | lzx | H2 versions.ts 真源 GET /agreements/current |
| 自定义游戏偏好上限:服务端 20/产品口径 3 | ljr | config 读面缺口,随 E9 或裁 |
| D-W population 余件(provider 行 member_population/座位迁移/guard 注释清理) | lzx | E8 权益面触发 |
| 走查墙 3:match 编排含无引擎款(tracking/skill 排进四项,流程 4/4 走不完) | hzy(H-C 顺带)或 lzx 小修 | W2 |
| 走查墙 4:运营测试账号口径——pg2-probe 无 user 座位登不进 admin 端;种子超管=admin/admin123456 | 用户侧 | 即办(SQL 补座位或文档化) |
| 走查注:admin 工作台 ERR·NETWORK(dev 栈缺数据源,非 esports 面) | lzx | pre 部署前查 |
| hzy 台账评审标记销项(H1/H-B 行曾滞「待评审」,本表已并) | — | 已随本档剪枝闭 |

## Review 遗留转交(2026-10-20 ctr 全项目 review;开放项已并入上表,此处仅存注记)

接受不改注记:B4 `user-scalable=no`(游戏 H5 防双击缩放,产品定位接受);D8 stores 空壳(D-T 能力配置面,机制待命)。

## 检查点登记(D-M)

| 检查点 | 日期 | 判定 | 拉闸动作 |
|---|---|---|---|
| CP1 | 11-08 | 待检(前瞻:支付沙箱=风险项) | — |
| CP2 | 11-22 | 待检 | — |
| CP3 | 12-02 | 待检 | — |

## 里程碑验收记录

(空;收口时一行一注)

## 待裁项状态

见 README §2.2;分析进展记 prd.md 对应条目。

## 节奏记录(一行一事;证据与细节归 git log)

- 10-18 立项+目录化(D-J);五路旧仓分析+三路审计入档;grill 定稿(D-K..D-Q,12-17 承诺);D-R/S/T 落;E6a 壳开工同日收口;E0 种子片冻结;E1 八步落地;E2 首轮落地。
- 10-19 E6b 接通片:微信真人链路全绿(D-V 兑现);E2 按 D-V 撤回;D-W/X/Y 裁;组合缝 a8b9f0a4;admin 只读面前拉提案否决(一面一消费等级)。
- 10-20 E0 增量2;E2 增量a 重建;E3 增量a(metric.md 单源);E6b 完整形态+机制修复+R1R2+流程前段 UI;H1/H2 交付;MR!2(H2+H-A)评审合并;会员读面+admin 用户管理落地;D-Z/D-AA/D-O2;批量追审+修复片;mock 插桩 TODO 法裁。
- 10-21 E0 增量3/4;E3 增量b;H-A/H-B 交付+走查;P1(D-AB/AC)/P9(D-AD)定案;D-AE/D-AF;小程序栈初步讨论(plan.md M5)。
- 10-22 E3 增量c/d(提交缝+MT/TR 定稿,双轴评审吸收两实错);hzy 精修批 MR!4 回流(三红根治/profile 写缝/D-AG 落地/C3);D-AG/AH/AI 裁;让步节制+环境分裂律立(README §5);注释精简批;dev-hzy 全量并入 main。
- 10-23 D-AJ 任务分配判据+H-C 分配+hzy 线盘点(27f87fb8);本档整体剪枝重梳:状态对齐 git 事实(全部「待评审」滞标改已并 main)、补登会员读面/admin 用户管理行、修切片表结构断裂、开放遗留清单建网。
- 10-23 E3-e:dual_zone/skill_trajectory scorer 定稿(净钟单向化+配置锚+终局对齐;哨兵双射闸;无效词表增 ABSURD_TIME_SPAN),metric.md §6.7/6.8 入档;自起栈三探针全绿(dz 90/st 521.6/反例 DETAIL_MISMATCH)。
