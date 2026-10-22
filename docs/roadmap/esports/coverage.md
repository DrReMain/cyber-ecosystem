# 原型×新版覆盖矩阵 — 精修基线盘点

> 归属:esports 前沿档,2026-10-20 建立。三人+各自的 coding agent 共用同一基线:精修期每片开工前,对照本矩阵对应行的语义锚做逐条核对;谁落地谁把该行状态改为「已实现」并注差异。矩阵错漏当场修正,不另立流程。
> 语料真相源(旧仓 manage,只读):页面映射=manage 仓 docs 下的 DESIGN_DOC_MAP §二(交互基线↔视觉基线);**每页语义规则=ctr 交互原型的 `_check.js`(逐条引 PRD 节号)**;裁决台账=同目录 LOGIC_AUDIT(L1/L2/L3);PRD=block3/block4 各自的 PRD。FEATURE_LIST 自标部分过时,仅作检索辅助。
> 2026-10-23 ljr 材料全量吸收(D-AL):新增可用锚=GAME_LOGIC_SPEC_DRAFT(match_config 种子/副能力 4 行 0.5/聚合 D1-D3)、USER_TRACEABILITY_PRD(user:*:view 分区权限点+手动解锁语义)、ADMIN_NAVIGATION_IA(9 模块)+ADMIN_INTERACTION_GUIDE(O-1/2/3)、db/README(36 表全景,V17-19 未发布)、旧仓 esport-server|esport-admin `openspec/changes/`(agreement-consent/membership-pay 等代码变更规格,clean-room 高价值参考);METRIC_SPEC 族仍按 §六.3 单源化消费。
> 验收口径=D-O 语义级 + **D-O2 四要素(2026-10-20)**:布局一样、细节接近、文案一样、视觉接近——原型为 AI 产出非专业 UI/UX 设计,且受新栈组件库(RAC+tailwind)约束,不追像素级;语义锚(rules 面板/PRD 规则)仍逐条对齐(§0.5 十一条口径)。
> 状态图例:✅ 已实现(语义级) · 🟡 骨架/占位 · ⬜ 未开工。

---

## 一、登录 / 资料 / 账号

| prototype 页(视觉锚) | 状态 | 承接片 | 差异与缺口注记 |
|---|---|---|---|
| `login`(block3/01+变体) | ✅ | E6b | 密码 tab 完整+协议双勾+微信环境门控+法务尾。**SMS tab=占位**(system 无契约);微信昵称一键(02-wechat 变体)未做——依赖绑定契约;双勾拦截 toast 指名缺哪项(§01)+备案四要素权威串核验=**B 批**(§七.5) |
| `agreement`(user-agreement+privacy-policy) | ✅ | H2 | 已落 `/agreement?p=user\|privacy`(宽容收窄非法值剥参回落 user;入 sitemap+可索引;head SSR 动态 title);登录页两处协议链接已切真链(H2 授权注记);文案=原型演示文本占位(ljr 定稿后换);版本参数/consent 版本记录=契约缺口随 E6b(progress 已登记) |
| `profile-nickname`(block3/02) | ✅ | E6b | 2-12 码点+规则卡;原型「禁 emoji」未显式实现(服务端校验兜底);昵称加权 24 冲突待 ljr 追认 |
| `profile-age`(block3/03) | ✅ | E6b | 8 档派生 ✓;**性别三选未做**(P8);次级年龄段口径=新版仅出生年月单口径(UpdateProfile 无 band 入参)——原型裁决 LOGIC_AUDIT ⑭=band-only 亦合法(guard 二选一,精确 birth 只在付费强制),与契约单轨相抵 → **P11 待裁** |
| `profile-setup`(block3/04+custom) | ✅ | E6b | 分组多选+类型投票注+自定义 ≤3;服务端 cap=20 config 读面缺(已登记) |
| `bind-wechat` / `bind-phone`(含 conflict 变体) | ⬜ | P4 前置 | 互绑=SMS+绑定契约缺口;D-G/G 待裁;profile 安全区三行「未开放」占位已在 |
| `account-merge-*`(block3/confirm+success) | ⬜ | P4 前置 | 空号并入富号;双方有数据不自动合并 |
| `password`(三态合一 ?m=) | ⬜ | 依赖 SMS | 忘记/设置/修改三态;忘记密码链接占位已在 |
| `account-security`(原型自建页) | 🟡 | 随互绑 | 新版并入 profile 九区「账号与安全」三行(未开放态);独立聚合页形态待互绑片定 |

## 二、测评链路

| prototype 页(视觉锚) | 状态 | 承接片 | 差异与缺口注记 |
|---|---|---|---|
| `home`(future/03-home) | ✅ | E6b | 问候态机+双门控 CTA+TabBar 四槽两活;首页不在 block3/4 视觉范围,原型即权威 |
| `match-guide`(05/17,?genre=) | ✅ | 前段 UI 片(2026-10-20)+E4 会话 | steps=每能力 1 款按 render_order;客户端临时编排器(arranger,E4 整体替换),类型 URL ?genre= 覆盖 |
| `match-switch-confirm`(15) | ✅ | 前段 UI 片+E4 | 会话生效不覆盖默认;进度警告面留 E4(done 恒 0) |
| `match-genre-pick`(16,?why=) | ✅ | 前段 UI 片 | tie/empty/normal 三面齐;平票由编排判定跳入 |
| `test-guide`(06/18+retest 变体) | 🟡 | 前段 UI 片(首测面)+E4 | 五规则+成绩口径已落;首次/二次切换与重测文案激活随 E4 会话 |
| `test-intro`(8 张→?ab=) | ⬜ | E7/H-A | 四段进度条;「继续」进本体 |
| `test-running`(唯一横屏) | ✅(引擎+三款+本地结果面) | E7/H-A/H-B | 引擎已立:/test/$code+GameClock(暂停顺延)+overlay 族+red_dot 可玩(mock config);真响应式横竖屏(无 rotate hack);遥测=rawDetail 事件累积+BackgroundSpan 后台段标记(服务端判);H-B 逐款长游戏+本地结果页 |
| `test-result`(8 张→?ab=) | ⬜ | E7 | 大分数+S/A/B/C+重测机会 N/3+取最高 |
| `exit-confirm`(31) | ✅ | H1→H-A | kit 弹层形态由 H-A 消费定形(/test 退出确认首用);续测文案半句留 E4 |
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
| `age-required`(23) | ✅(页面) | H2+PAY3 | 整页已立 `/age-required`(sessionGuard+noindex;保存走 UpdateProfile,<18 拦截;年范围 6-60 法则);当前仅直连 URL 可达——付费硬闸挂接与入口归 PAY3 |
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

## 五、admin 端矩阵(原型=admin/outputs/prototype 41 页;侧栏 9 组权威=_nav.js)

> admin 原型的「系统与合规」组大部分是**底座 IAM 既有面**(角色/权限/管理员/审计/菜单随 system 模块复用,不重做);esports 业务面按 M2 核心/余组与 E9 承接。原型自带「非本期」灰标四项(报告配置/渠道管理/落地页/归因)与新版 M2 余组(梯①)高度重合,映射照旧。

| 侧栏组(原型页) | 状态 | 承接片 | 注记 |
|---|---|---|---|
| 仪表盘(dashboard/user-portrait) | ⬜ | M2 余(梯①) | 运营总览+用户画像 |
| 用户与会员(01-user-list/01-member-list/02-\*-detail/09-user-import/announcement) | 🟡 | **用户列表 ✅**(esports users 模块,列表+6 Tab 详情骨架+封禁);会员管理→M2 核心;导入/公告→余组 | 会员详情 Tab 余面随 E8/PAY 补 |
| 题库与测试(question-bank/question-detail/test-type/test-record/08-test-record-detail/game-config/game-category) | ⬜ | E9(激活码+测试项/配置+题库种子查看);test-record 随 E2/E3 | 游戏与因子配置=配置驱动运行时面 |
| 定价与订单(sku-manage/04-order-record/04/05-order-detail/payment-channel) | ⬜ | PAY1+M2 核心 | SKU 原价/现价结构(D-F) |
| ToB 租户(tob-tenant-list/detail) | ⬜ | M3/P4 | 专版 buffer 窗 |
| 报告与内容(report-config,原型非本期) | ⬜ | M2 余(梯①) | 报告板块配置面 |
| 渠道与推广(activation-code(-detail)/channel-manage/landing-material/attribution,后三者原型非本期) | ⬜ | 激活码→E8+E9+M2 核心;渠道/落地/归因→M2 余(增长裂变) | 激活码=承诺窗口内(兑换防超卖 e2e) |
| 反馈与客服(feedback-manage/07-feedback-detail) | ⬜ | M2 余(梯①) | 与 ctr 消息中心同族 |
| 系统与合规(audit-log/role/permission/admin-user/menu-manage/system-config=**底座 IAM 既有面复用**;compliance-config/agreement-manage/agreement-consent/rating-manage→esports 域⬜ 随 M2 余+H2 consent 契约;app-version→**见缺口①**) | 🟡 | 底座面复用;esports 域随 M2 余 | 评级管理 P2;同意记录=consent 版本化契约缺口 |

## 六、语料版本律(防过时引用)

1. **引用先过权威表**:任何设计/需求文档作为对照锚前,先核 DESIGN_DOC_MAP §四权威等级——✅ 权威可用;⚠️ 过时(FEATURE_LIST/DESIGN.md 评分档位口径等)不得单独为据,须与 PRD/_check.js 交叉;🗄️ 归档只作历史证据。
2. **滞后窗口事实**:权威地图快照=2026-09-18,归档整理=09-26,其后 manage/docs(GAME_LOGIC_SPEC_DRAFT/SKILL_INTERACTION_SPEC 等)与 ctr 原型个别页仍在演进——**表外或快照后变动的文档,引用前先经 ljr 确认权威级**,不直接消费。
3. **数值档单独律**:评分/校准族(UNIT_SCORE/METRIC_SPEC 族/CALIBRATION_PLAN/TRACKING_DESIGN/FLASH_AIMING_TEST_STANDARD,位于 design-systems/admin/)=§0.2 经验数据层,E3 开工时按 README §3 收入评分规格单源档后才可依赖,原文档即失效(单源化)。

## 七、盘点发现的缺口(切片池漏网/待归属)

1. **version-update 无承接片**:原型有页、admin 侧有 app-version 页面(`app-version.html`,系统与合规组),但 E0-E10/H-*/PAY/M2 均未排。前置=system/esports 是否有版本契约;归属建议下次例会定(admin 面随 E9?ctr 面随 M2?)。
2. **每页语义锚的消费纪律**:ctr 交互原型的 `_check.js` rules 面板(逐条引 PRD 节号)是防偷懒的现成抓手——承接片开工时把对应页的 rules 逐条搬进该片评审锚,逐条对上才算语义级完成(E0 测试读面片已示范此法);admin 原型无同款面板,以对应 PRD 节+BLOCK5_ADMIN_REFORM_PRD 为锚。
3. **迁移口径同步律**:本矩阵行所属聚合落地时,当日回 migration.md 对应表补行(不积压到 cutover)。
4. **权威地图滞后**(2026-10-20 核查):DESIGN_DOC_MAP 快照(09-18)之后 manage/docs 与 ctr 原型仍有演进——处理见 §六语料版本律,不视为 ljr 侧缺陷,以流程消化。
5. **B 批修正清单(已建面×PRD 细则;2026-10-23 S1 校准后修订版,一片收)**:①genre-pick 补「推荐」角标(原型 §16 命中偏好类型带角标,现缺)。②profile-nickname 禁 emoji 显式拦截(现仅 trim+服务端兜底)。③age.tsx 改消费 age-required/birth.ts(现自带本地 years/ageOf 副本=遗留清单「上收」反例)。④向导 games 步完成去向改直连 /match(原型/PRD=04→05;现去首页——若 home 中枢是有意设计需 ljr 一句确认)。⑤profile-overview 模块集对核(PRD:训练趋势本期空态/报告中心内嵌 1 条+查看全部/账号安全=微信手机两行绑定)。⑥V4 视觉值对齐批:match 子页(pick/switch/guide)+agreement+age-required 仍用 rounded-lg/sm 旧词汇(V1-V3 未覆盖)。⑦admin 详情返回文案带目标名(O-1 断言形状「返回 XXX 列表」)。**S1 撤案两枚**:login「toast 指名缺哪项」(已在:agreeOK 拼名 toast)、setup「列表后端取」(已在:listGames 零硬编码)。横屏律=仅 test-running ✓ 已核无残留。归属=lzx。

## 八、变更记录

- 2026-10-20 建立:ctr 44 页交互基线全量盘点;抓到缺口①(version-update 无承接片)。
- 2026-10-20 增补:admin 端 41 页×9 组矩阵(用户指示双端覆盖);语料版本律(权威表前置+滞后窗口+数值档单源化);验收口径落 D-O2 四要素。
- 2026-10-23 D-AL 全量吸收:档头注+行注记 3 处(login/profile-age/setup 核验面)+§七.5 B 批清单;metric.md 核对零改动(七款单位分/槽位/档位套吻合)。
- 2026-10-23 S1 校准战役(语义层,代码证据):已建 14 面逐条过 _check.js/PRD/LOGIC_AUDIT 判据——大面全绿(登录簇/向导三步/首页/match 簇/协议/补填/hzy 系统态件);B 批清单修订(§七.5:两撤四增一 V4);新登记后续片锚=微信回填昵称头像(需 system identity 支撑)、PAY3 月粒度满岁口径。台账见 §九。

## 九、校准战役台账(2026-10-23 起;S1=ctr 语义,S2=admin 交互文案,S3=契约种子,S4=终局对账,随片续行)

> 判据四分级:✅对上 / ⚠小修(并入 B 批)/ ❌偏差待裁 / 📋后续片锚。S1 方法=逐面读实现代码×_check.js 规则逐条勾(视觉值层归 V4 批自带探针,不在本战役重复)。

| 面 | 判定 | 证据要点(代码位) |
|---|---|---|
| login+回调+守卫 | ✅ | 双勾三态+指名 toast(page.tsx agreeOK)/微信双环境置灰+提示/协议校验盖微信腿/?next= 回跳+开重定向卫生(after-login)/去向=资料完整性(guard.profileComplete 同口径)/备案四要素齐(messages login_footer_*) |
| profile-nickname | ✅(⚠emoji) | 2-12 码点+计数条+trim 拦+仅真实值预填;emoji 显式拦截缺=B批② |
| profile-age | ✅ | 6-60 年范围/双下拉+8 档派生提示/未满13 硬闸+guardianAccepted 独立落库/u18 提示;bands 码值=V18 词形;⚠age.tsx 未消费 birth.ts=B批③ |
| profile-setup(games) | ✅ | listGames 后端取零硬编码/未归类不计投票/平票注/cap 提示+删除×+换行;⚠完成去向=首页非 /match=B批④ |
| home | ✅ | 游客可浏览/点测评拦+next 回跳/hasGames 门控(profileFlowGuard 承真闸);续测卡/最近成绩=E4/E5 锚 |
| match 主页 | ✅ | 竖屏/tie+empty 路由/编排序=D-Z/?genre= 覆盖+switched 标/done=0 恒(已登记 E4) |
| match 子页×3 | ✅(⚠视觉) | switch=整页形态+会话生效;guide=五规则+成绩口径+首测面恒定;pick=tie/empty/normal 三面+整卡可点——缺推荐角标=B批①;rounded 旧词汇=V4 |
| agreement / age-required | ✅(⚠视觉) | 版本行=versions.ts 占位(已登记)/演示标注诚实;补填页双出路+<18 页内拦+6-60;rounded 旧词汇=V4 |
| hzy 系统态件(退出弹层/状态屏/HUD) | ✅ | 31 弹层形态+危险语义+cancel 焦点;空态 action 槽(裁决⑳);HUD fab/倒计时/低时红脉冲=旧仓法 |
| **S2** admin esports-users | ✅(⚠文案×1) | O-1/2/3 全合规(独立详情页/操作列收敛/敏感 Confirm 含审计提示且按真实执法语义);6-Tab=溯源 PRD F2 逐一对应;确认语/列名/空态文案对齐;⚠返回文案「返回」→ 带 O-1 断言形状目标名(并入 B 批);术语差记录:列名「会员」vs 原型「用户」=member plane 有意差不改;缺列/高级筛选/导出/批量导入=数据未到(E5/E8/system/A9)已在 §五注记;报告/溯源行内深链随 E5/E8 |
| **S3** 契约+种子+服务端 | ✅(❌一枚) | game 九款逐字保真(lol/dota2/honor/csgo/valorant/pubg/apex/overwatch/fortnite);13 款码名逐字+类别=D-Z 逐行吻合+池法全非空;contribution 次贡献 4 行=D-AA 递延在案;target_recognition=RS=D-AD 在案;**❌moving_target 主能力两源冲突**(交互权威 _state.js POOL=TS vs 数据权威 contribution 主行 AA+B2 次贡献注释)→ **P12 提请 ljr 裁**;满岁月粒度口径已锚 PAY3 |
| **S4** 终局对账(四列融合:旧后端 33 表端点级/旧前端 admin20+ctr60 路由/新原型 ctr43+admin41/新 PRD 全量) | 📋 | 判据=「计划执行完→终极目标」可达性。总判定:承诺窗(M1+M2 核心+PAY+M4)基本可达,前提=支付即启+H-C 落+未认领片认领;全量原型需 P13 豁免签认。三矛盾提裁:P13(窗外豁免)/P14(school→租户时序)/P15(报告簇窗口:本矩阵 report 簇承接=E5/M2 余,与 D-AB 默认 open 错位——该行即 P15 证据)。立片:MIG1/E5b/E6c+E9 增量锚(plan 已同步)。疑点解除:档位四档=BLOCK5 §7.7 逐字吻合(§7.3 六级=系统套称号并行词汇,非冲突);旧库复核=无支付网关(权益翻译口径不变)、基线 V15(b4「V16+wx_jsapi」=新版前向编号无存量面);8 维公式源已锚=旧仓 esport-server 仓内 OpenAPI_用户端接口文档.md |

**新登记后续片锚(不在 B 批)**:微信登录回填昵称/头像(LOGIC_AUDIT 2.5b-⑫;需 system identity 持久化+向导去向联动→E6b 余面小片);满岁口径=月粒度近似「生日次日」(契约仅年+月,PAY3 执法时定保守口径);home 内容模块(续测/最近成绩/数据统计)随 E4/E5。
