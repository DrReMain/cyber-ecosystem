# Admin 渲染覆盖台账(U1-admin 活证走查)

> 立档:2026-11-06(admin 起栈日)。与 render-coverage.md(ctr)同法:逐面逐门记 活证/待造态/观测受限;判定词同族(已活证/部分活证/待造态/观测受限/预留门)。走查法=chrome-devtools-mcp 三件套(docs/conventions/browser-walkthrough)。
> 范围:`app/clients/admin/src/routes/` 48 文件 → 面清单;业务区(esports-*)为主,U3 基座范式审计(calibration-log 2026-11-06 三裁定)与走查同轨。
> 登录法:admin/admin123456 种子,密码表单;隔离上下文(与 ctr 共享 session_token cookie 名)。

## §0 已活证门(2026-11-06 X2-1 起栈日)

| 门 | 判定 | 证据 |
|---|---|---|
| 匿名守卫·redirect 播种 | 已活证 | 直击 /dashboard/feedback/tickets → /login?redirect=%2Fdashboard%2F…%3FpageNo%3D1%26pageSize%3D20(全路径+search) |
| 登录·redirect 兑现 | 已活证 | admin/admin123456 提交 → 直达播种目的地反馈队列页 |
| 反馈队列四态 Tab | 已活证 | 全部/待处理/已回复/已关闭;已关闭 Tab 写 status=closed 入 URL,共 2 条 |
| 队列行·状态门控动作 | 已活证 | pending 行=详情+回复+关闭(aria-label 图标钮);已关闭行=仅 详情 |
| 快捷回复 modal | 已活证 | 行内回复 → textarea(0/2000)+取消/回复 → toast「已回复 …(同步 C 端消息中心)」 |
| 详情页·meta 四格 | 已活证 | 用户(链 user-detail)/分类/提交时间/联系方式(contact 自留行在场) |
| 详情页·线程四节 | 已活证 | user→admin→user→admin,双端逐字一致 |
| 关单·O-3 警示确认 | 已活证 | Popconfirm「关闭这条反馈?关闭后对话终止、不可恢复;用户如有新问题需提交新反馈」 |
| 关单·幂等形态 | 已活证 | 结构性缺席:已关闭行动作钮消失;服务端 status=closed |
| 导航菜单九组+面包屑+页签栏 | 已活证 | 工作台/用户与会员/…/平台管理;A→B 页签可关闭 |
| 公告台·两步发布流(X2-2) | 已活证 | 新建 modal(标题64/类型/范围/正文5000+helper)→存草稿 toast→行内发布确认+toast「并投递 全部用户」;存量三行(信全员/横幅/信会员已下线) |
| 公告扇出·服务端对账 | 已活证 | user_message 13 行=wholeRoster 快照(中秋 11+新号;wt 首投即达);banner 零投递代码+存量横幅零行双证 |
| 用户详情·总览 meta | 已活证 | 昵称/年龄段/出生年月/注册时间/游戏偏好 cs2/用户 ID——向导写入全量回读 |
| 用户详情·测评与报告三表 | 已活证 | 场次表(七列)+能力汇总(六列)+明细(七列)空板态 |
| 用户详情·账号面 | 已活证 | 状态 正常/人群 member |
| 用户详情·消息与反馈 Tab | **parked** | 「暂未开放」stub → G-413(X2-3 admin 腿) |
| 用户详情·溯源 Tab | parked | 「暂未开放」(疑有意,未见登记) |
| ctr 侧 X2-2/3 连带 | 已活证 | tabbar aria「消息 3 条未读」(信1+回复2 对账)/信箱计数条+信卡/开信即读(is_read 服务端翻)/CTA 查看对话→反馈会话精准落 |

## §0.1 走查现场发现(待裁/待修)

| 发现 | 去向 |
|---|---|
| 详情链 href 缺 locale 前缀(用户链有 /zh-CN,详情链无;路由功能正常,复制链接丢 locale) | 轻伤入 gap 候选,随 U3 批 |
| 详情页无关闭动作(关单只在队列行) | U3 交互审计项 |
| reopened 两词汇:ctr「处理中」/admin「待处理」(同态两词,疑有意:用户面=处理中,运营面=待你处理) | 观察在案,产品确认 |
| 动作列=aria-label 图标钮(§1.2 陷阱主场) | 走查法已入 conventions |

## §1 面清单(48 路由文件;2026-11-06 U1-admin 首轮逐面盘点毕:`__TSR_ROUTER__` 串走 25 面 = 22 live + 1 观测受限 + 2 parked;菜单全项直达**零死链**(A4 死线面首证))

| 区 | 面 | 态 |
|---|---|---|
| login | 登录(账密) | 已活证(§0) |
| dashboard/index | 工作台 | **观测受限**:agent 服务(11002)未起,vite 代理 GetMyAgentConfig ECONNREFUSED,main 悬空;起 agent 后复走 |
| people/users | 用户列表+详情($id 六 Tab:总览✓/溯源 parked/测评与报告✓/会员待细走/消息与反馈 parked=G-413/账号✓) | 部分活证;列表=基座 Filter(昵称/状态)+5 列 live |
| people/members | 会员列表+详情 | 已活证:列表四 Tab 带活计数(全部2/有效1/≤7天1/过期0)+来源筛+8 列;详情=会员锚面全 live(即将到期徽/续费/终止/跳转用户档案/掩码手机/剩余天数/累计会员期/期内解锁永久保留) |
| people/announcements | 公告控制台 | 已活证(§0) |
| assessment/game-config | 游戏与因子配置 | 已活证:6 列(代码/名称/类别/能力/时长/操作),red_dot 行在场 |
| pricing/orders | 订单与解锁记录 | 已活证:三 Tab(全部/已支付/待支付)+Filter(订单号/解锁方式/商品)+9 列;desc 自述「数字内容默认不退」 |
| pricing/skus | SKU 管理 | 已活证:新建+三 Tab(全部/上架/未上架)+7 列;desc「金额以分存储,改价敏感入审计」 |
| reports/report-policies | 报告访问策略 | 已活证:6 列,认知=会员/职业=开放策略行在场 |
| channels/activation-codes | 批次列表+$batchNo 码表 | 已活证:列表三筛+生成批次+6 列(B-2 批次行);详情 B-2611-LFQR=启用/复制未核销/导出 CSV+进度 100.0%+**需补码**+码表三态 Tab+5 列 |
| feedback/tickets | 队列+详情 | 已活证(§0) |
| compliance | agreement-versions/consents/config | 已活证×3:versions 7 列+新建;consents **共 62 条**+导出 CSV+「一次登录写两行」自述;config 开关组(出生采集/年龄线/到期提醒/文案) |
| agents/connections | 连接管理 | 已活证:4 列+「暂无 agent 连接」诚实空板 |
| system | users/roles/depts/policies/audit/diag | 已活证×6(IAM 底座回归):users 三筛+7 列;roles/policies 同式;depts 空板;audit 四筛+9 列**共 97 条**+导出;diag 判定模拟器(主体/操作/判定过程) |
| files | list | 已活证:四筛+代理/直连上传+9 列 |
| platform | populations/tenants | 已活证×2:关键词筛+Member/Default 行在场 |
| profile | 密码/智能体 | 已活证:密码表单(旧/新/确认+确认修改)+智能体子页导航 |
