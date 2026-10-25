// The scenario bank's client mirror. The judging truth lives server-side
// (the seeded answer key, loaded through the submit seam — an unknown
// scenario id is an invalid attempt, never a degraded one); this mirror
// exists because M1 carries no bank-read RPC, so the run face renders
// from a verbatim copy of the same twenty rows. The optimal field feeds
// only the per-round feedback card — the legacy client-side isCorrect
// assertion stays dead. Content changes must land on both sides.

export type OdChoice = "attack" | "defense" | "retreat";
export type OdComplexity = "simple" | "medium" | "complex";

export interface OdScenario {
  readonly code: string;
  readonly category: string;
  readonly description: string;
  readonly choices: readonly { readonly choice: OdChoice; readonly text: string }[];
  readonly optimal: OdChoice;
  readonly complexity: OdComplexity;
}

export const OFFENSE_SCENARIOS: readonly OdScenario[] = [
  {
    code: "OD_001",
    category: "中期",
    description: "敌方中单消失，我方下路正在推进",
    choices: [
      { choice: "attack", text: "敌方中单可能来gank，不安全" },
      { choice: "defense", text: "被动防守无法获得优势" },
      { choice: "retreat", text: "安全撤退，避免被gank" },
    ],
    optimal: "retreat",
    complexity: "simple",
  },
  {
    code: "OD_002",
    category: "早期",
    description: "我方打野在附近，敌人塔下残血",
    choices: [
      { choice: "attack", text: "配合打野越塔击杀" },
      { choice: "defense", text: "错失击杀机会" },
      { choice: "retreat", text: "浪费打野gank机会" },
    ],
    optimal: "attack",
    complexity: "medium",
  },
  {
    code: "OD_003",
    category: "中期",
    description: "敌方团队正在打龙，我方队伍分散",
    choices: [
      { choice: "attack", text: "团队分散无法参战" },
      { choice: "defense", text: "防守等待时机，避免团灭" },
      { choice: "retreat", text: "放弃龙损失太大" },
    ],
    optimal: "defense",
    complexity: "complex",
  },
  {
    code: "OD_004",
    category: "早期",
    description: "敌方打野上路gank，我方有视野且闪现未用",
    choices: [
      { choice: "attack", text: "打野在gank无法进攻" },
      { choice: "defense", text: "被gank风险高" },
      { choice: "retreat", text: "使用视野优势及时撤退" },
    ],
    optimal: "retreat",
    complexity: "medium",
  },
  {
    code: "OD_005",
    category: "中期",
    description: "我方在中路防御塔处5打3人数优势",
    choices: [
      { choice: "attack", text: "人数优势推进获胜" },
      { choice: "defense", text: "浪费人数优势" },
      { choice: "retreat", text: "错失优势期" },
    ],
    optimal: "attack",
    complexity: "simple",
  },
  {
    code: "OD_006",
    category: "早期",
    description: "敌方刺客正在游走支援你的线路，无视野",
    choices: [
      { choice: "attack", text: "无视野容易被击杀" },
      { choice: "defense", text: "被动等待不利" },
      { choice: "retreat", text: "无视野情况下保守撤退" },
    ],
    optimal: "retreat",
    complexity: "medium",
  },
  {
    code: "OD_007",
    category: "后期",
    description: "我方刚赢下团战，大龙刷新，敌方存活2人",
    choices: [
      { choice: "attack", text: "人数劣势打龙风险大" },
      { choice: "defense", text: "守住优势防止翻盘" },
      { choice: "retreat", text: "浪费团战胜利" },
    ],
    optimal: "defense",
    complexity: "complex",
  },
  {
    code: "OD_008",
    category: "中期",
    description: "敌方射手无闪现，我方辅助控制技能已好",
    choices: [
      { choice: "attack", text: "抓住技能差击杀ADC" },
      { choice: "defense", text: "错失击杀机会" },
      { choice: "retreat", text: "浪费优势时机" },
    ],
    optimal: "attack",
    complexity: "simple",
  },
  {
    code: "OD_009",
    category: "后期",
    description: "敌方团队正在推我方基地，我方可以偷家",
    choices: [
      { choice: "attack", text: "偷家速度不够快" },
      { choice: "defense", text: "回防保护基地" },
      { choice: "retreat", text: "放弃基地必败" },
    ],
    optimal: "defense",
    complexity: "complex",
  },
  {
    code: "OD_010",
    category: "后期",
    description: "后期，双方在大龙处对峙，我方团战能力更强",
    choices: [
      { choice: "attack", text: "团战能力强主动开团" },
      { choice: "defense", text: "后期机会难得" },
      { choice: "retreat", text: "错失取胜机会" },
    ],
    optimal: "attack",
    complexity: "complex",
  },
  {
    code: "OD_011",
    category: "中期",
    description: "我方上单正在单带，敌方正在包围他",
    choices: [
      { choice: "attack", text: "队友被包围无法进攻" },
      { choice: "defense", text: "防守或支援上单" },
      { choice: "retreat", text: "放弃队友不利" },
    ],
    optimal: "defense",
    complexity: "medium",
  },
  {
    code: "OD_012",
    category: "早期",
    description: "敌方防御塔血量很低，但敌方即将复活",
    choices: [
      { choice: "attack", text: "敌方复活后危险" },
      { choice: "defense", text: "塔已残血不需防守" },
      { choice: "retreat", text: "撤退避免被反杀" },
    ],
    optimal: "retreat",
    complexity: "medium",
  },
  {
    code: "OD_013",
    category: "后期",
    description: "我方拥有远古巨龙buff，敌方防守",
    choices: [
      { choice: "attack", text: "远古龙buff碾压性推进" },
      { choice: "defense", text: "浪费龙buff时间" },
      { choice: "retreat", text: "放弃绝佳机会" },
    ],
    optimal: "attack",
    complexity: "simple",
  },
  {
    code: "OD_014",
    category: "中期",
    description: "敌方优势核心位置未知，我方想要强开团战",
    choices: [
      { choice: "attack", text: "核心位置未知风险大" },
      { choice: "defense", text: "谨慎等待敌方暴露位置" },
      { choice: "retreat", text: "过于保守" },
    ],
    optimal: "defense",
    complexity: "complex",
  },
  {
    code: "OD_015",
    category: "早期",
    description: "在我方野区抓到敌方打野，队友在附近",
    choices: [
      { choice: "attack", text: "人多击杀落单打野" },
      { choice: "defense", text: "错失击杀机会" },
      { choice: "retreat", text: "浪费优势" },
    ],
    optimal: "attack",
    complexity: "simple",
  },
  {
    code: "OD_016",
    category: "中期",
    description: "敌方有你的视野，他们正在设置埋伏",
    choices: [
      { choice: "attack", text: "已被发现会中埋伏" },
      { choice: "defense", text: "被动等待不利" },
      { choice: "retreat", text: "撤离危险区域" },
    ],
    optimal: "retreat",
    complexity: "medium",
  },
  {
    code: "OD_017",
    category: "后期",
    description: "我方高地水晶已破，超级兵正在推进",
    choices: [
      { choice: "attack", text: "兵线压力无法进攻" },
      { choice: "defense", text: "清理超级兵线" },
      { choice: "retreat", text: "放弃防守必败" },
    ],
    optimal: "defense",
    complexity: "simple",
  },
  {
    code: "OD_018",
    category: "中期",
    description: "敌方交出关键大招，我方技能冷却完毕",
    choices: [
      { choice: "attack", text: "抓住技能差发起团战" },
      { choice: "defense", text: "错失优势窗口" },
      { choice: "retreat", text: "浪费机会" },
    ],
    optimal: "attack",
    complexity: "medium",
  },
  {
    code: "OD_019",
    category: "早期",
    description: "前期阶段，敌方打野等级领先",
    choices: [
      { choice: "attack", text: "等级劣势会被反杀" },
      { choice: "defense", text: "前期不需过度防守" },
      { choice: "retreat", text: "前期发育避免战斗" },
    ],
    optimal: "retreat",
    complexity: "simple",
  },
  {
    code: "OD_020",
    category: "中期",
    description: "我方实力更强但打完目标后血量较低",
    choices: [
      { choice: "attack", text: "血量低容易被反杀" },
      { choice: "defense", text: "已经拿到目标" },
      { choice: "retreat", text: "安全撤退巩固优势" },
    ],
    optimal: "retreat",
    complexity: "complex",
  },
];
