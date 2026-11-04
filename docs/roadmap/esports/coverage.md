# 原型×新版 — 语义锚索引(2026-11-05 压缩版)

> **本档 2026-11-05 瘦身**:§一..五逐页状态表+§九校准台账全数退役——处置(建没建)=verify/demand-ledger.md;映射(页→路由)=verify/ctr-prototype-matrix.md;面检(建得像不像)=verify/q-line.md(Q1-Q6 全毕)+gap-ledger;历史全文=git。本档继续承重=**验收口径+语料版本律+活锚清单**三件。
> 归属:esports 前沿档,三人共写(2026-10-20 建立)。

## 验收口径(D-O2,2026-10-20)

UI 对照四要素=**布局一样、细节接近、文案一样、视觉接近**。依据:原型为 AI 产出非专业 UI/UX 设计,且受新栈组件库(ctr=RAC+tailwind)约束,不追像素级;语义锚(rules 面板/PRD 规则)仍逐条对齐。ctr 侧已定义部分高度对齐(值对齐),未定义可增强;admin 侧对齐对象=**交互流+文案**(antd+皮肤系统,不做像素对齐,D-AK)。

## 语料真相源与版本律(防过时引用)

1. 语料入口:旧仓 manage 仓 `docs/DESIGN_DOC_MAP`(导览)+`DOC_INVENTORY`;页面语义规则=ctr 交互原型 `_check.js`(逐条引 PRD 节号);裁决台账=同目录 LOGIC_AUDIT;admin 交互=BLOCK5_ADMIN_REFORM_PRD+ADMIN_INTERACTION_GUIDE(O-1/2/3)。**ctr 活原型=渲染态读**(`_state.js` 驱动,静态解析只得壳);admin 原型=JS 渲染活原型(admin-member-anchor 实证)。
2. **引用先过权威表**:任何设计/需求文档作对照锚前,先核 DESIGN_DOC_MAP §四权威等级——✅ 权威可用;⚠️ 过时(FEATURE_LIST/DESIGN.md 评分档位口径等)不得单独为据,须与 PRD/_check.js 交叉;🗄️ 归档只作历史证据。
3. **滞后窗口事实**:权威地图快照=2026-09-18,归档整理=09-26;其后 manage/docs 与 ctr 原型个别页仍可能有演进(manage 仓 git log 最近一次 design-systems 变动=入库五笔,2026-11-05 核零演进)——表外或快照后变动的文档,引用前先经 ljr 确认权威级。
4. **数值档单独律**:评分/校准族(UNIT_SCORE/METRIC_SPEC 族/CALIBRATION_PLAN/TRACKING_DESIGN 等)=经验数据层,已按 README §3 收入 metric.md 单源档,原文档废止引用。

## 活锚清单(仍开放的语义锚;其余随对应切片开工时从原型/PRD 现读)

| 锚 | 内容 | 归属 |
|---|---|---|
| test-records 三态 | block4 未完成/完成锁定/历史最佳三态设计(?t=) | M2 余组记录读面 |
| message 族 | message-center 双 Tab/feedback 六类枚举/对话式详情;notification-detail 补充面 | ctr 腿 ✅(2026-11-05,MSG-1);admin 三页=W4 |
| share-invite 两态 | 免费层 5 维卡+动态三档文案+未成年代称+保存图片 | MSG-2(P23) |
| E4 三配置模块 | 学科优势匹配/职业发展潜能/提升计划(契约无字段,后台可配置+空态不造假) | P4(ljr)/M3 |
| scheme-b 三页 | 渠道版 6 能力直测+渠道报告 | M3/P4 窗外 |
| admin 余组 13 页 | 逐页开工时以 _nav.js+对应 PRD 节为锚(D-AQ admin 面流程) | M2 余组梯① |

## 迁移口径同步律

本矩阵原行所属聚合落地时,当日回 migration.md 对应表补行(不积压到 cutover)。

## 变更记录

- 2026-10-20 建立:ctr 44 页+admin 41 页双端矩阵;语料版本律;D-O2 四要素。
- 2026-10-23 D-AL 全量吸收;S1 校准战役(B 批清单);10-23..25 S2/S3/S4 毕(台账原 §九)。
- 2026-10-30 状态列退役(处置真值=demand-ledger)。
- 2026-11-05 瘦身(Q1-Q6 面检役全毕+PAY3 四片收口后):逐页表退役为三档分工(demand-ledger/matrix/q-line),本档留口径+版本律+活锚;B 批⑦ 项全闭(R1-R5 收口);S1-S4/P12/P15/P19 待裁全闭。
