# esports 验证工作台 — 用例设计、覆盖映射与对账锚

> 重立:2026-11-07 重构役三。本目录=正式测试阶段的仪器集:用例设计执行真值(integration-matrix)、三方覆盖映射与缺口(coverage-map)、需求处置真值(demand-ledger)、断言锚源(q-line)、导航锚(ctr-prototype-matrix/admin-nav-anchor/admin-member-anchor)、保真 harness 重激活 SOP(HARNESS-SOP)。
> 体制(如实记档):2026-11-05 走查役起,走查与开发同会话执行(发现→当场修→记账同手)——因此**红→绿翻转与「毕」判定须第二证据**(11-07 三律:join 在案项/视觉轴标注/毕=动词全走);历史分权体制(独立验证会话)描述已废。
> 历史取证档(gap-ledger/render-coverage×2/calibration-log/pm-review-items/admin-gap-matrix/coverage)2026-11-07 退役,**全文=git 391ccc87**;残指一律按此读。

## 防失真协议(承前,继续生效)

1. **清单由脚本生成** — 计数类数字(demand-ledger §四 等)随收口脚本再生,禁手打快照。
2. **转写零手工** — 逐字内容一律脚本/渲染态提取,断言按路径引用。
3. **台账落仓** — 进度存文件不存上下文,跨会话续读文件。
4. **出处必填** — 每行标注事实源路径,可任意抽查对质。
5. **计数收口** — 每轮 N 文件→N 行,数目对不上不收口。

## 断言纪律(承前)

- **蒙眼规则**:期望值只从原型/规格档推导,禁止先读实现再写断言。
- **copy.ts 不是事实源**:断言「渲染文本==原型字面量」;文案库对原型的一致性是独立检查项。
- **分层断言**:文案层逐字全等;结构层语义断言;视觉层人工走查+几何测定(G-419 判例:布局判定须几何测定,目测不算)。
- **数值事实源**:metric.md 单源;战略口径=PRODUCT_STRATEGY_DECK。

## 原型语料(权威层级,⚠️ 待 PM 确认)

旧仓 `lab/esports/manage/design-systems/ctr/` 四层+导览页,共 **192 html**(2026-11-07 机核:44+45+80+22+1 导览;机器锚=`ctr-design-inventory.tsv`,与磁盘零差集):`prototype/` 44=**主 UI 事实源**;`outputs/block4` 45;`outputs/block3` 80(含 rejected);`outputs/future` 22=远期参考。admin 原型=`design-systems/admin/outputs/prototype/` 41 页;侧栏九组权威=_nav.js。
原型=活应用(壳+`_state.js`):逐字提取须渲染态读或直读 `_state.js`;每状态枚举=一条用例。

## 原型固化律(2026-10-28 裁定,承前)

原型在开发阶段不更新,是唯一还原真相:先严格还原,后细节调优。存疑不阻塞、不放宽断言——转 prd 偏离接洽单留档;与仓法冲突项同样只记账。范围口径(哪页何时建)走 plan.md,与保真口径分离。

## 文件

| 文件 | 内容 |
|---|---|
| `integration-matrix.md` | 集成走查矩阵:**用例设计+执行真值**(58 例 X13/C30/A15;态列 R2 起清零重走) |
| `coverage-map.md` | 覆盖映射:需求→用例→代码三方对照、缺口清单(P0-P4 收割序)、豁免台账、新偏离登记 |
| `demand-ledger.md` | 需求总账:85 页×处置(进度对账唯一真值)+活锚清单 |
| `q-line.md` | Q 线校对锚:页级三源断言清单(矩阵预期锚的主锚源;态列=11-04 Q1-Q6 役快照) |
| `ctr-prototype-matrix.md` | ctr 44 页→路由纯映射(开工导航) |
| `admin-nav-anchor.md` | admin 菜单真值(_nav.js 九组) |
| `admin-member-anchor.md` | 会员两页走查 oracle(渲染态机提) |
| `cutover-watch.md` | 上线走查+值守清单(hzy 座位,W6 执行稿基) |
| `HARNESS-SOP.md` | 保真 harness 重激活 SOP(封存档,重激活只许照抄) |

## 执行节奏

- 正式测试阶段=多轮走查:每轮开工对账 coverage-map 缺口清单(收割序 P0→P4)+plan 缺陷工单;毕例记「毕@R<n>」;发现按四路处置(面级即修/缺陷→plan/偏离→prd/裁定→prd)。
- harness 重激活(U2/保真轮)按 HARNESS-SOP 照抄落盘,禁凭记忆重写。
- 退役档取证:`git show 391ccc87:docs/roadmap/esports/verify/<档名>`。
