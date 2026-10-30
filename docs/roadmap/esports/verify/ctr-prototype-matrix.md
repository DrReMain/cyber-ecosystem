# V-0 · ctr 原型 44 页 × 重建映射矩阵

> **2026-10-30:本表状态列=10-28 快照,已滞后(report 族/账号族/绑定族其后落地);现行处置唯一真值=demand-ledger.md §一。本表继续承重=原型页→重建路由的映射关系与面检工作项。**
> 左列 = `prototype/` 全部 44 个 HTML(脚本枚举,零手工);重建列 = ctr 路由(24 个路由文件)。
> 状态:✅已实现(面检待做) / ◐部分实现或形态差异待面检 / ❌未实现 / 📄壳或通用态 / ⏸待裁。
> 「待面检」= V-1/V-2 逐字+结构比对的工作项;❌ = 红条 gap,开发会话的后续验收契约。

## 主表(44 行)

| # | 原型页 | 屏名 | 重建落点 | 状态 | 说明 |
|---|---|---|---|---|---|
| 1 | index.html | 原型入口/导航壳 | — | 📄 | 状态直达选择器,非产品屏 |
| 2 | home.html | 首页 | `/` | ◐ | 续测卡已有;最近成绩/数据统计/报告中心入口待面检 |
| 3 | login.html | 登录 | `/login` | ✅ | E5 走查过密码登录;微信面待面检 |
| 4 | profile-nickname.html | 完善资料·昵称 | `/profile/setup/nickname` | ✅ | 三步向导之一 |
| 5 | profile-age.html | 完善资料·年龄 | `/profile/setup/age` | ✅ | 三步向导之二 |
| 6 | profile-setup.html | 完善资料·游戏偏好 | `/profile/setup/games` | ✅ | 三步向导之三(名称差异:setup→games) |
| 7 | profile-overview.html | 个人中心 | `/profile` | ◐ | 页面在;内容面(会员卡/消息入口/设置项)待面检 |
| 8 | password.html | 密码管理 | — | ❌ | 无独立页;block4 另有 set/change-password 设计 |
| 9 | match-genre-pick.html | 选择测评类型 | `/match/pick` | ✅ | |
| 10 | match-guide.html | 匹配引导·编排列表 | `/match` | ◐ | 名字错位:重建把编排列表独立成 `/match`;形态待面检 |
| 11 | match-switch-confirm.html | 切换类型确认 | `/match/switch` | ✅ | |
| 12 | test-guide.html | 测评前须知 | `/match/guide` | ◐ | 名字错位:重建 `/match/guide` 是须知;逐字待面检 |
| 13 | test-intro.html | 单项测试介绍 | `/test/$code`(intro 态) | ✅ | |
| 14 | test-running.html | 测试中 | `/test/$code`(running 态) | ✅ | |
| 15 | test-result.html | 单项测试结果 | `/test/$code`(end/review 态) | ✅ | E5-f2 三标签+收据卡 |
| 16 | test-done.html | 测评完成 | `/report/session/$sessionId` | ✅ | E5-f2 按原型落地 |
| 17 | test-records.html | 我的测试记录 | — | ❌ | block4 有未完成/完成锁定/最佳三态设计 |
| 18 | report-single.html | 能力测评报告 | `/report/$recordId` | ✅ | E5-d 七区块 |
| 19 | report-center.html | 报告中心 | — | ❌ | 四卡入口+解锁态(已测 N/6 门槛文案) |
| 20 | report-career.html | 电竞职业指导 | — | ❌ | 原型标「暂未开放·后台开关控制」 |
| 21 | report-cognitive.html | 认知与潜力评估报告 | — | ❌ | 8 维认知;测满 6 项才完整 |
| 22 | report-trend.html | 能力成长趋势 | — | ❌ | 基础免费/进阶随会员 |
| 23 | member-detail.html | 会员权益 | — | ❌ | 报告页内有会员卡(E5-d),独立权益页无 |
| 24 | pay-confirm.html | 确认支付 | — | ❌ | pay parked;契约(OrderService)已在 |
| 25 | pay-result.html | 支付结果 | — | ❌ | block4 有成功/失败/取消三态 |
| 26 | message-center.html | 消息中心 | — | ❌ | |
| 27 | feedback-submit.html | 意见反馈 | — | ❌ | |
| 28 | feedback-detail.html | 反馈详情 | — | ❌ | |
| 29 | agreement.html | 用户协议与隐私政策 | `/agreement/user` + `/agreement/privacy` | ◐ | 原型一壳两协议 vs 重建两路由;全文逐字待对 |
| 30 | age-required.html | 补填出生年月 | `/age-required` | ✅ | |
| 31 | bind-phone.html | 绑定手机号 | — | ❌ | |
| 32 | bind-wechat.html | 绑定微信 | — | ❌ | `/login/wechat/callback` 仅回调,非绑定管理 |
| 33 | account-security.html | 账号与安全 | — | ❌ | |
| 34 | account-merge-confirm.html | 账号合并 | — | ❌ | |
| 35 | resume-remind.html | 继续上次进度 | `/`(home 续测卡) | ◐ | 卡片 vs 原型提醒形态待面检 |
| 36 | share-invite.html | 分享测试结果 | — | ❌ | |
| 37 | scheme-b-entry.html | 渠道版入口 | — | ❌ | 渠道 B 方案三件套 |
| 38 | scheme-b-guide.html | 渠道版直测说明 | — | ❌ | |
| 39 | scheme-b-report.html | 渠道版报告 | — | ❌ | |
| 40 | exit-confirm.html | 已退出测评 | `/test/$code`(弹层?) | ◐ | 是否有退出确认弹层待面检 |
| 41 | version-update.html | 版本更新 | — | ⏸ | H5 形态下是否适用待裁 |
| 42 | loading.html | 加载中 | 全局 | 📄 | 通用态,待面检 |
| 43 | empty.html | 空态 | 全局 | 📄 | 通用态,待面检 |
| 44 | error.html | 网络异常 | 全局 | 📄 | 通用态,待面检 |

**计数**:44 = ✅13 + ◐6 + ❌20 + ⏸1 + 📄4(壳 1 + 通用态 3)。

## blocks/future 补充面(原型 44 页之外、语料内存在的屏或口径)

| 出处 | 内容 | 处置 |
|---|---|---|
| block4/notification-detail.html | 通知详情 | 未实现,随 message-center 族 |
| block4/account-set/change-password.html | 设置/修改登录密码 | 未实现,随 password 族 |
| block4/subscription-popup.html · theme-hero.html | 订阅弹窗 · 主题英雄位 | 未实现,待裁是否在 scope |
| block3/moba-6dim-report.html(+unlocked) | moba 6 维报告形态 | ⚠️ 与现行 4 款×genre 形态不同,产品口径待裁 |
| block3/forgot-password.html · bind-*-guide 系 · account-merge-success | 登录/绑定族补充态 | 未实现 |
| block4/PRODUCT_STRATEGY_DECK-20260827 | **分数双轨**(无上限分 + 百分位/评级) | ⚠️ 数值断言的战略口径,用例设计前须确认现行重建采哪轨 |
| future/ 22 页 | 学校报告/训练模式/装扮商店/青训自荐/训练营等 | 远期范围参考,不入本期 gap |

## 待裁清单(用例铺开前须 PM 拍板)

1. 权威层级假设(prototype > block4 > block3 > future)是否成立。
2. 分数双轨口径(block4 战略 deck)对现行数值断言的约束。
3. moba 6 维报告与 8 维认知报告的形态是否在 scope。
4. version-update / scheme-B 渠道版是否入本期。
