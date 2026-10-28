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
   - ctr 面 → `verify/ctr-prototype-matrix.md`(44 页×重建路由主表;真源=`lab/esports/manage/design-systems/ctr/prototype/`,活原型交互基线)
   - admin 面 → `verify/admin-gap-matrix.md`(27 行三路映射)+ `coverage.md` §五(41 页矩阵;真源=`design-systems/admin/outputs/prototype/`,2026-09-14 批)
   - 旧码逻辑参考 → `coverage.md` §一..四语义锚 + 片行已锚单源(metric.md 等);数值/口径另引 `design-systems/admin/` 的 METRIC_SPEC_*/PRODUCT_STRATEGY_DECK(verify README 断言纪律同源)
2. **读源**(只读映射命中件):ctr 原型是**活应用**(壳+`_state.js` 驱动 JS 渲染),文案与状态须渲染态读(chrome-mcp 开 file:// 取渲染后 DOM)或直读 `_state.js` 数据——静态剥标签只得壳文本;admin 原型静态页,可正则剥标签抽语义(E9-b 判例)。原型自带状态枚举(解锁/锁定/已测 N/6 等),每状态一条对齐用例。
3. **偏离白名单**(仅三形,逐条登记,此外不偏离):
   - 契约没有的面不造 → `verify/gap-ledger.md` 登红条;
   - 旧 schema 语义已被新裁定作废(如激活码租户绑定/双用途/停用)→ 收口注声明裁定号;
   - 仓法压倒原型(antd 规范/lucide 法等)→ 只记 `verify/pm-review-items.md`,不擅自二选一(原型固化律)。
4. **收口声明**:客户端面切片 commit 的 proven-by 须含「设计源:n 文件」清单与对齐判定;progress 收口注一行带计数。

## gap 消费环与已建面欠账

- `verify/gap-ledger.md` 红条=开发工单:归属面切片顺路消化,或立独立精修片;修复后验证会话复跑转绿(verify README 既定环,双向喂给)。
- 先行于本法的已建面不推倒重做——以红条工单回补;面检由验证会话按矩阵铺开(◐ 页)。

## 范围

新建或修改 `app/clients/ctr` / `app/clients/admin` 页面、组件、文案的切片。服务端逻辑对齐走既有单源(metric.md/prd.md/coverage 锚,D-AL 吸收法),不在本法范围。
