const AVATAR_COLORS = ["#1677ff", "#722ed1", "#13c2c2", "#fa8c16", "#eb2f96", "#52c41a"];

export function avatarColor(seed: string): string {
  let sum = 0;
  for (const ch of seed) sum += ch.charCodeAt(0);
  return AVATAR_COLORS[sum % AVATAR_COLORS.length] ?? "#1677ff";
}
