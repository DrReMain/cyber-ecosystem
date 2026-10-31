// Fen is the wire unit everywhere; the operator types and reads yuan.
// Trailing zeros drop — ¥19.9, not ¥19.90.
export function fenToYuan(fen: bigint | number | null | undefined): string {
  if (fen == null) return "-";
  const v = Number(fen);
  const s = (v / 100).toFixed(2).replace(/\.?0+$/, "");
  return `¥${s}`;
}

// Parses operator input ("19.9", "¥19.90") into fen; null = not a price.
export function fenOfYuan(input: string): bigint | null {
  const t = input.trim().replace(/^¥/, "");
  if (!/^\d+(\.\d{1,2})?$/.test(t)) return null;
  const parts = t.split(".");
  const yuan = parts[0] ?? "";
  const cents = parts[1] ?? "";
  return BigInt(yuan) * 100n + BigInt(cents.padEnd(2, "0"));
}
