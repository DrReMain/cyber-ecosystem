# ctr 原型 44 页 × 重建路由 — 纯映射表

> **本表唯一职责=原型页→重建落点的映射**(设计对齐前置法 D-AQ 的开发侧导航:客户端面切片开工先查此表读命中原型源,渲染态读 `_state.js`)。处置/建没建=demand-ledger §一;像不像=gap-ledger+q-line;历史状态快照=git(2026-10-28 版)。
> 2026-11-05 瘦身重推导:映射列按现行 43 路由重核,44 行零手工(路由树 `ls routes/*.tsx` 对账)。

## 主表(44 行)

| # | 原型页 | 屏名 | 重建落点 |
|---|---|---|---|
| 1 | index.html | 原型入口/导航壳 | —(评审工具页,非产品屏) |
| 2 | home.html | 首页 | `/` |
| 3 | login.html | 登录 | `/login`(+`/login/forgot`、`/login/wechat/callback`) |
| 4 | profile-nickname.html | 完善资料·昵称 | `/profile/setup/nickname` |
| 5 | profile-age.html | 完善资料·年龄 | `/profile/setup/age` |
| 6 | profile-setup.html | 完善资料·游戏偏好 | `/profile/setup/games`(名称差异:setup→games) |
| 7 | profile-overview.html | 个人中心 | `/profile` |
| 8 | password.html | 密码管理 | `/profile/security/password`(三态 ?m=,忘记=/login/forgot) |
| 9 | match-genre-pick.html | 选择测评类型 | `/match/pick` |
| 10 | match-guide.html | 匹配引导·编排列表 | `/match`(名称错位:重建以编排列表为 match 主页) |
| 11 | match-switch-confirm.html | 切换类型确认 | `/match/switch` |
| 12 | test-guide.html | 测评前须知 | `/match/guide`(名称错位:须知在 match 域) |
| 13 | test-intro.html | 单项测试介绍 | `/test/$code`(intro 态) |
| 14 | test-running.html | 测试中 | `/test/$code`(running 态,唯一横屏) |
| 15 | test-result.html | 单项测试结果 | `/test/$code`(end/review 态) |
| 16 | test-done.html | 测评完成 | `/report/session/$sessionId` |
| 17 | test-records.html | 我的测试记录 | 未建(→M2 余组,block4 三态 ?t= 设计在) |
| 18 | report-single.html | 能力测评报告 | `/report/$recordId` |
| 19 | report-center.html | 报告中心 | `/report/center` |
| 20 | report-career.html | 电竞职业指导 | `/report/career` |
| 21 | report-cognitive.html | 认知与潜力评估报告 | `/report/cognitive` |
| 22 | report-trend.html | 能力成长趋势 | `/report/trend` |
| 23 | member-detail.html | 会员权益 | `/member`(PAY3 片4) |
| 24 | pay-confirm.html | 确认支付 | `/pay`(PAY3 片2;oracle=P7 稿 p7-member-paywall.html) |
| 25 | pay-result.html | 支付结果 | `/pay/result`(PAY3 片3,四态) |
| 26 | message-center.html | 消息中心 | 未建(→MSG-1 余面) |
| 27 | feedback-submit.html | 意见反馈 | 未建(→MSG-1 余面) |
| 28 | feedback-detail.html | 反馈详情 | 未建(→MSG-1 余面) |
| 29 | agreement.html | 用户协议与隐私政策 | `/agreement/user`+`/agreement/privacy`(一壳两协议 vs 两路由,在册分歧) |
| 30 | age-required.html | 补填出生年月 | `/age-required`(逃生口=遗留清单) |
| 31 | bind-phone.html | 绑定手机号 | `/profile/security/bind-phone` |
| 32 | bind-wechat.html | 绑定微信 | `/profile/security/bind-wechat` |
| 33 | account-security.html | 账号与安全 | `/profile/security` |
| 34 | account-merge-confirm.html | 账号合并 | `/profile/security/merge`(+success) |
| 35 | resume-remind.html | 继续上次进度 | `/` 续测卡(形态差=卡片 vs 整页,G-208 在册) |
| 36 | share-invite.html | 分享测试结果 | 未建(→MSG-2) |
| 37 | scheme-b-entry.html | 渠道版入口 | 未建(窗外随 M3/P4,D-AS③) |
| 38 | scheme-b-guide.html | 渠道版直测说明 | 未建(同上) |
| 39 | scheme-b-report.html | 渠道版报告 | 未建(同上;P14 已解散) |
| 40 | exit-confirm.html | 已退出测评 | BottomConfirmSheet(全局单源组件) |
| 41 | version-update.html | 版本更新 | 未建(已裁 B=D-BC:ctr 轻量提示随 M2 余组) |
| 42 | loading.html | 加载中 | 全局 StateScreen |
| 43 | empty.html | 空态 | 全局 StateScreen(域内变体随消费面) |
| 44 | error.html | 网络异常 | 全局 ErrorPage/ErrorFallback(network kind=暖卡,Q1-③) |

**计数:44 = 已映射 31 + 未建 9 + 壳/通用 4**(与 demand-ledger §一 计数互通:未建 9=❌8+⏸1)。

## blocks/future 补充面(44 页之外)

| 出处 | 内容 | 处置 |
|---|---|---|
| block4/notification-detail.html | 通知详情 | 随 message-center 族(MSG-1 余面定形态) |
| block4/account-set/change-password.html | 设置/修改登录密码 | ✅ 已随 password 族 |
| block4/subscription-popup.html · theme-hero.html | 订阅弹窗 · 主题英雄位 | ⏸ 不入(P13 已裁) |
| block3/moba-6dim-report.html | moba 6 维报告形态 | ⏸ 不入(P15/D-AP 已裁:窗内 4×genre 形态) |
| block3/forgot-password.html · bind-*-guide 系 · account-merge-success | 登录/绑定族补充态 | forgot ✅/merge-success ✅;guide 系=设计注记不迁 |
| block4/PRODUCT_STRATEGY_DECK | 分数双轨 | 已定案=总则 §0.5(6) 双轨并存(PM-6 闭) |
| future/ 22 页 | 学校报告/训练模式/装扮商店等 | 远期参考,不入账 |

## 待裁清单 — 已全闭(2026-11-05 校准确认)

1. 权威层级(prototype > block4 > block3 > future)——事实采纳,D-AK/D-AQ 法条落地。
2. 分数双轨——已裁:双轨并存(总则 §0.5(6));PM-10 明细行两半边已落。
3. moba 6 维——已裁不入(P15/D-AP)。
4. version-update / scheme-B——已裁:D-BC(B 轻量版)/D-AS③(窗外随 M3)。
