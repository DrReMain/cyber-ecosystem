# Q 线·业务码校对精修役台账

> 2026-11-02 D-BB 立役。法条=plan.md Q 线节:三源对照(原型 D-AQ 渲染态读+prd/demand-ledger+proto 契约),作者盲独立复核;发现项**全部上线前修完**——面级即时走链落码,契约/schema 级入 lzx 队列同窗修完;e2e 全套=回归门。射面=esports 业务面/模块(IAM 底座六面=骨架车道不在射面)。
> 每片开工细化逐项验收清单(本档);收口行迁 progress 认领表。

## Q1 ctr 测评主链(11-03 开工;射面=登录/注册/向导→match→games→成绩+pending 明细)

清单基线=ctr-prototype-matrix.md 44 行映射(状态列滞后以 demand-ledger §一为准)。校对序=先 ◐ 待面检(未被任何走查覆盖),后近期已走查面(独立复扫)。已毕走查的消费面不重复登记存量判定,只记新发现。

### 逐项清单

| 段 | 面 | 原型→重建 | 校对点(三源断言) | 态 |
|---|---|---|---|---|
| A1 | 登录 | login.html→/login | tab 序/默认验证码;微信腿 UA 态(✅G-005 已证);协议勾选门槛+全选;错误行;忘记密码腿 | 待 |
| A2 | 昵称步 | profile-nickname→/profile/setup/nickname | 文案逐字;校验口径;跳过态 | 待 |
| A3 | 年龄步 | profile-age→/profile/setup/age | 文案逐字;年龄线(合规配置真源);补填闭环 | 待 |
| A4 | 偏好步 | profile-setup→/profile/setup/games | 文案逐字;三步编排序;完成去向(D-BD=/match) | 待 |
| A5 | 年龄门槛 | age-required→/age-required | 触发条件;文案;去向 | 待 |
| A6 | 协议壳 | agreement→/agreement/{user,privacy} | 版本行真源(E9 已落);面布局 vs 原型壳 | 待 |
| B1 | 首页 | home.html→/ | 续测卡(✅R2);最近成绩/数据统计/报告中心入口逐字全过;会员 greet=Q1-① | 毕(①) |
| B2 | 编排列表 | match-guide→/match | 形态+字面全过(PM-9/12 裁已吸收;剩N次收据=E7 增强) | 毕 |
| B3 | 类型选择 | match-genre-pick→/match/pick | tie/normal/empty 三态+toast 逐字全过;推荐角标=登记增强 | 毕 |
| B4 | 切换确认 | match-switch-confirm→/match/switch | 逐字全过(rule2 分歧=G-208 N1 已注册) | 毕 |
| B5 | 前须知 | test-guide→/match/guide | 16/17 逐字全过;rule5 过期语义=Q1-② | 毕(②) |
| C1 | 单项介绍 | test-intro→/test/$code intro | 能力名/介绍文/开始钮;开始前守卫 | 待 |
| C2 | 测试中 | test-running→running 态 | 题头(✅PM-9);中断/退出(**exit-confirm 形态待面检**);计时 | 待 |
| C3 | 退出确认 | exit-confirm→BottomConfirmSheet | G-208 首刀闭(逐字+几何实测) | 毕 |
| C4 | 通用态 | loading/empty/error→全局 | idiom 族分歧=Q1-③(P 票) | 裁决中 |
| C5 | 记录页 | test-records | 矩阵行 17+block4 三态在册 | 核登记毕 |
| D1 | 成绩面 | test-result→end/review | 活跑+回读双态(E5-f2+6b2db084+ca729f5f 刚毕=低优先复扫) | 待(低) |
| D2 | 完成页 | test-done→/report/session/$id | 逐字;四款清单态;去向(报告/首页) | 待 |
| D3 | 续测提醒 | resume-remind→/ 续测卡 | ✅R2 毕(去向差=G-208 页面片);Q1 只核在册 | 核在册 |

### 发现与处置(面级即时修;契约级入 lzx 队列)

**第一腿 11-03(B 段全+C 段毕)**:B2/B3 逐字全过(增强项有据:/match 步行「剩 N 次」收据=E7 追踪批、pick「推荐」角标=代码注登记);B4 过(rule2 分歧=G-208 走查 N1 已注册,零欠账);C3 过(G-208 首刀闭,逐字+几何实测);C5 核登记毕(test-records ❌=矩阵行 17+block4 三态,在册)。

| # | 面 | 断言与证据 | 级 | 处置 |
|---|---|---|---|---|
| Q1-① | B1 home greet 会员态 | 原型 home.html:52-54 会员活性时 greet=「专属会员 · 有效期内」;重建仅 欢迎/未登录 两态。C 端无会员活性读:ProfileService 仅 Get/Update/PreviewMerge(profile.proto 三 RPC),唯一 member_until 在合并预览事实(profile.proto:114);MemberService=admin 面(member.proto:20「this is the admin plane」);GetAbilityDigest 无会员字段(report.proto:231) | **契约级** | 入 lzx 队列。建议形状:GetProfileResponse 增会员活性(如 member_until)——home 会话探针本就调 GetProfile,零加跳;客户端三态 greet 随契约落 |
| Q1-② | B5 须知 rule5 过期语义 | 原型/LOGIC_AUDIT(D-AL 吸收锚「续测 7 天反推」):「超 24h 缓存失效时会按已测结果反推剩余项,不重排」;重建文案「过期后下次进入会重新编排」=**服务端真行为**(match/biz.go:195-203 expireIfStale 过窗懒闭+下场新抽)。文案从行为、分歧未在册 | 口径 | 一句裁:裁 A=重抽法定案(在册登记分歧,文案不动) / 裁 B=服务端改反推续测(lzx)。同族参照:switch rule2 同型分歧已由 G-208 N1 注册裁从行为 |
| Q1-③ | C4 通用态族 idiom | error:原型暖卡(📡+「网络连接失败」+「本次进度已保留」安抚行+重试/返回上一页/返回首页三钮,error.html:9-18)vs 重建 ERR 终端美学(ERR chip+巨型 HTTP 码+重试/返回首页,error-page.tsx);loading:原型骨架屏 vs 重建 mono 扫描条(StateScreen 注=自觉系统选择);empty:原型三域变体 vs 重建通用+域内空态 | 作者级 | 整族裁决(P 票):裁 A=网络错面还原原型暖卡(仅 network kind 有 oracle,余 kind 无原型面维持现状=两 idiom 并存) / 裁 B=骨架美学仓法定案(登记分歧,零代码) / 裁 C=全族还原。**进度保留安抚行**不论裁向都值得单独落(测评场景关键 UX,行为真:attempts 逐款提交) |

Q1 余量:A 段(login 11-01 E9 复走查毕/向导三步/age-required/agreement 壳)+C1/C2/D1/D2(近走查面,低优先独立复扫)——第二腿。

