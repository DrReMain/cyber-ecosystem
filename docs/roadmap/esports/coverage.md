# 原型×新版覆盖矩阵 — 精修基线盘点

> 归属:esports 前沿档,2026-10-20 建立。三人+各自的 coding agent 共用同一基线:精修期每片开工前,对照本矩阵对应行的语义锚做逐条核对;谁落地谁把该行状态改为「已实现」并注差异。矩阵错漏当场修正,不另立流程。
> 语料真相源(旧仓 manage,只读):页面映射=manage 仓 docs 下的 DESIGN_DOC_MAP §二(交互基线↔视觉基线);**每页语义规则=ctr 交互原型的 `_check.js`(逐条引 PRD 节号)**;裁决台账=同目录 LOGIC_AUDIT(L1/L2/L3);PRD=block3/block4 各自的 PRD。FEATURE_LIST 自标部分过时,仅作检索辅助。
> 验收口径=D-O 语义级(流程/状态/数据完整),非像素级;UI 不承担高保真复原(D-O+§0.2 资产三级),业务语义必须对齐(§0.5 十一条口径)。
> 状态图例:✅ 已实现(语义级) · 🟡 骨架/占位 · ⬜ 未开工。

---

## 一、登录 / 资料 / 账号

| prototype 页(视觉锚) | 状态 | 承接片 | 差异与缺口注记 |
|---|---|---|---|
| `login`(block3/01+变体) | ✅ | E6b | 密码 tab 完整+协议双勾+微信环境门控+法务尾。**SMS tab=占位**(system 无契约);微信昵称一键(02-wechat 变体)未做——依赖绑定契约 |
| `agreement`(user-agreement+privacy-policy) | ⬜ | H2 | URL 形状已定约 `/agreement?p=user\|privacy`(2026-10-20 lzx);登录页协议链接现为占位 toast;consent 版本记录=契约缺口(progress 已登记) |
| `profile-nickname`(block3/02) | ✅ | E6b | 2-12 码点+规则卡;原型「禁 emoji」未显式实现(服务端校验兜底);昵称加权 24 冲突待 ljr 追认 |
| `profile-age`(block3/03) | ✅ | E6b | 8 档派生 ✓;**性别三选未做**(P8);次级年龄段口径=新版仅出生年月单口径(UpdateProfile 无 band 入参,已登记) |
| `profile-setup`(block3/04+custom) | ✅ | E6b | 分组多选+类型投票注+自定义 ≤3;服务端 cap=20 config 读面缺(已登记) |
| `bind-wechat` / `bind-phone`(含 conflict 变体) | ⬜ | P4 前置 | 互绑=SMS+绑定契约缺口;D-G/G 待裁;profile 安全区三行「未开放」占位已在 |
| `account-merge-*`(block3/confirm+success) | ⬜ | P4 前置 | 空号并入富号;双方有数据不自动合并 |
| `password`(三态合一 ?m=) | ⬜ | 依赖 SMS | 忘记/设置/修改三态;忘记密码链接占位已在 |
| `account-security`(原型自建页) | 🟡 | 随互绑 | 新版并入 profile 九区「账号与安全」三行(未开放态);独立聚合页形态待互绑片定 |

## 二、测评链路

| prototype 页(视觉锚) | 状态 | 承接片 | 差异与缺口注记 |
|---|---|---|---|
| `home`(future/03-home) | ✅ | E6b | 问候态机+双门控 CTA+TabBar 四槽两活;首页不在 block3/4 视觉范围,原型即权威 |
| `match-guide`(05/17,?genre=) | ⬜ | E4+E7 | steps=每能力 1 款按 render_order;D-E 匹配语义已冻 |
| `match-switch-confirm`(15) | ⬜ | E4+E7 | 会话生效不覆盖默认;已测不再编排 |
| `match-genre-pick`(16,?why=) | ⬜ | E4 | 平票/无偏好手动改选;空编排兜底 |
| `test-guide`(06/18+retest 变体) | ⬜ | E7 | 首次 3 次重测/二次 1 次;四提醒卡 |
| `test-intro`(8 张→?ab=) | ⬜ | E7/H-A | 四段进度条;「继续」进本体 |
| `test-running`(唯一横屏) | ⬜ | E7/H-A | GameClock+暂停顺延;遥测采集新契约 |
| `test-result`(8 张→?ab=) | ⬜ | E7 | 大分数+S/A/B/C+重测机会 N/3+取最高 |
| `exit-confirm`(31) | 🟡 | H1→H-A | **ExitConfirmDialog kit 件已交付零消费者**;原型为整页落地,kit 为弹层——形态以 D-O 由消费片定 |
| `test-done`(27c/27c-moba) | ⬜ | E7 | 亮眼结果页;P3 交互决策(ljr) |
| `scheme-b-*`(block4/16-19,渠道版 3 页) | ⬜ | M3/P4 | 6 能力直测+渠道报告 |

## 三、报告 / 付费 / 分享

| prototype 页(视觉锚) | 状态 | 承接片 | 差异与缺口注记 |
|---|---|---|---|
| `report-single`(29c 两态+moba-6dim) | ⬜ | E5 | 免费层 5 维+解锁态;MOBA 测满 6 维;评分双轨(D-§0.5) |
| `report-center`(block4/01) | ⬜ | E5 | 三类聚合入口+解锁态 |
| `report-cognitive`(04/05) | ⬜ | E5 | 8 维;**缺维度不渲染**七硬约束;「已测 5/6」完整度指示 |
| `report-career`(02/03) | ⬜ | E5/M2 余 | 已裁「暂未开放」 |
| `report-trend`(31+subscription-popup) | ⬜ | M2 余 | 依赖 practice_record(P2 数据源待定) |
| `share-invite`(25 两态) | ⬜ | M2 余(梯①) | D-F 立减已删;ref HMAC 归因 |
| `pay-confirm` / `pay-result`(20-22/26/27) | ⬜ | PAY1-3 | 单次/会员;微信 JSAPI+H5 |
| `age-required`(23) | ⬜ | H2+PAY3 | 弹层→整页形态差;付费补填硬闸 |
| `member-detail`(24) | ⬜ | PAY/E8 | 权益+价格+FAQ |

## 四、个人中心 / 消息 / 通用

| prototype 页(视觉锚) | 状态 | 承接片 | 差异与缺口注记 |
|---|---|---|---|
| `profile-overview`(06 三变体) | 🟡 | E6b 骨架→E5/E8/PAY 补 | 九区骨架✓(身份卡/资料/安全占位/更多虚线/关于/退出);**会员卡/能力概览/测评历史/数据统计/训练趋势未落**(各归后续片);member 变体随 PAY |
| `test-records`(09/10/11 三态 ?t=) | ⬜ | E5/E7 | 未完成续测(7 天)/已完成/历史最佳;TabBar「记录」槽占位已在 |
| `resume-remind`(12) | ⬜ | E7 | 续测弹框 |
| `message-center`(13 两 Tab) | ⬜ | M2 余(梯①) | TabBar「消息」槽占位已在 |
| `feedback-*` / `notification-detail`(14/15) | ⬜ | M2 余(梯①) | profile 意见反馈入口占位已在 |
| `loading`/`empty`/`error`(28/29/30) | ✅ | H1+E6a | 三族 kit 件在位;消费者随各片落 |
| `version-update`(block4) | ⬜ | **无承接片(缺口)** | 见下节缺口① |

---

## 五、盘点发现的缺口(切片池漏网/待归属)

1. **version-update 无承接片**:原型有页、admin 侧有 app-version 后台概念,但 E0-E10/H-*/PAY/M2 均未排。前置=system/esports 是否有版本契约;归属建议下次例会定(admin 面随 E9?ctr 面随 M2?)。
2. **每页语义锚的消费纪律**:`prototype/_check.js` 的 rules 面板(逐条引 PRD 节号)是防偷懒的现成抓手——承接片开工时把对应页的 rules 逐条搬进该片评审锚,逐条对上才算语义级完成(E0 测试读面片已示范此法)。
3. **迁移口径同步律**:本矩阵行所属聚合落地时,当日回 migration.md 对应表补行(不积压到 cutover)。

## 六、变更记录

- 2026-10-20 建立:44 页交互基线全量盘点;抓到缺口①(version-update 无承接片)。
