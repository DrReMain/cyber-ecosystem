# 设计对齐前置法 — 客户端面切片的开工纪律(D-AQ)

> 立法:2026-10-28 用户裁定,D-AK 双轨标准的执行法。**开发阶段即对齐设计,不等验证会话事后精修**——开工读源的成本低于事后返工。
> 动因:语料量大(ctr 侧 192 HTML+admin 侧 41 页+旧仓双前端)曾致开发会话凭契约反推 UI、跳过原型阅读,产出偏离设计(V-1 login 9 条 gap 与 E9-b admin 初版为实证)。对策不是全读语料,是**查映射、只读命中件**——映射表已由验证会话建成。

## 两级标准(承 D-AK,重申)

| 端 | 对齐义务 |
|---|---|
| ctr | **高保真还原原型**:UI/UX+文案+逻辑三层(D-O2 四要素:布局一样/细节接近/文案一样/视觉接近) |
| admin | 高度还原**需求逻辑+文案**,交互尽量对齐;视觉走 antd/skin 规范,不逐像素(E9-b 判例) |

## 前置动作(认领后、写第一行页面代码前)

1. **查映射**(面→源,查表即得,禁漫游语料):
   - **归属与处置**(该页归哪个切片/建没建)→ `verify/demand-ledger.md`(2026-10-30 起唯一真值;未建页先看归属片是否已被 P 票阻塞)
   - ctr 面 → `verify/ctr-prototype-matrix.md`(44 页×重建路由主表;真源=`lab/esports/manage/design-systems/ctr/prototype/`,活原型交互基线)
   - admin 面 → `verify/demand-ledger.md` §二(41 页分桶;真源=`design-systems/admin/outputs/prototype/`,2026-09-14 批);**菜单落位/命名 → `verify/admin-nav-anchor.md`**(_nav.js 九组真值;admin 原型两页验证为 JS 渲染活原型,同 ctr 须渲染态读——member 两页实证 10-30;会员两页逐字 oracle=`verify/admin-member-anchor.md`)
   - 旧码逻辑参考 → 片行已锚单源(metric.md/prd.md 裁决族);数值/口径另引 `design-systems/admin/` 的 METRIC_SPEC_*/PRODUCT_STRATEGY_DECK(verify README 断言纪律同源)
2. **读源**(只读映射命中件):ctr 原型是**活应用**(壳+`_state.js` 驱动 JS 渲染),文案与状态须渲染态读(chrome-mcp 开 file:// 取渲染后 DOM)或直读 `_state.js` 数据——静态剥标签只得壳文本;admin 原型静态页,可正则剥标签抽语义(E9-b 判例)。原型自带状态枚举(解锁/锁定/已测 N/6 等),每状态一条对齐用例。
3. **偏离白名单**(仅三形,逐条登记,此外不偏离):
   - 契约没有的面不造 → `prd.md` 偏离接洽单登记(N/PM 系);
   - 旧 schema 语义已被新裁定作废(如激活码租户绑定/双用途/停用)→ 收口注声明裁定号;
   - 仓法压倒原型(antd 规范/lucide 法等)→ 同记 prd 偏离接洽单,不擅自二选一(原型固化律)。
4. **收口声明**:客户端面切片 commit 的 proven-by 须含「设计源:n 文件」清单与对齐判定;progress 收口注一行带计数。

## gap 消费环与已建面欠账(2026-11-07 役三重锚)

- 偏离工单=prd 偏离接洽单(R/N/PM 系);缺陷=plan 缺陷工单(B 系);走查覆盖缺口=verify/coverage-map.md——三台账在案项,开发与走查不得另下新判。
- 先行于本法的已建面不推倒重做——以工单回补;面检由走查轮按 integration-matrix 铺开(◐ 页)。

## 范围

新建或修改 `app/clients/ctr` / `app/clients/admin` 页面、组件、文案的切片。服务端逻辑对齐走既有单源(metric.md/prd.md 裁决族,D-AL 吸收法),不在本法范围。

## 语料真相源与版本律(2026-11-07 自 coverage.md 迁入;防过时引用)

1. 语料入口:旧仓 manage 仓 `docs/DESIGN_DOC_MAP`(导览)+`DOC_INVENTORY`;页面语义规则=ctr 交互原型 `_check.js`(逐条引 PRD 节号);裁决台账=同目录 LOGIC_AUDIT;admin 交互=BLOCK5_ADMIN_REFORM_PRD+ADMIN_INTERACTION_GUIDE(O-1/2/3)。**ctr 活原型=渲染态读**(`_state.js` 驱动,静态解析只得壳);admin 原型=JS 渲染活原型(admin-member-anchor 实证)。
2. **引用先过权威表**:任何设计/需求文档作对照锚前,先核 DESIGN_DOC_MAP §四权威等级——✅ 权威可用;⚠️ 过时(FEATURE_LIST/DESIGN.md 评分档位口径等)不得单独为据,须与 PRD/_check.js 交叉;🗄️ 归档只作历史证据。
3. **滞后窗口事实**:权威地图快照=2026-09-18,归档整理=09-26;其后 manage/docs 与 ctr 原型个别页仍可能有演进(2026-11-05 核零演进)——表外或快照后变动的文档,引用前先经 ljr 确认权威级。
4. **数值档单独律**:评分/校准族(UNIT_SCORE/METRIC_SPEC 族/CALIBRATION_PLAN/TRACKING_DESIGN 等)=经验数据层,唯一事实源=metric.md(README §3),原文档废止引用。
