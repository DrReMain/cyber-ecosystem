// Fen amounts ride the wire as bigint (Int64Value flattened); yuan is a
// display concern, never a stored one. Shared by every purchase face —
// the report wall and the pay wall must quote the same numbers.
export function fenToYuan(fen: bigint): string {
  const yuan = Number(fen) / 100;
  return Number.isInteger(yuan) ? String(yuan) : yuan.toFixed((yuan * 10) % 1 === 0 ? 1 : 2);
}

// The receipt's own form: two decimals always. The wall's prices strip
// trailing zeros (the P7 draft's ¥19.9), but a settled receipt is a
// money record — the prototype's pay-result quotes ¥19.90, and ledger
// faces keep ledger formatting.
export function fenToYuanExact(fen: bigint): string {
  return `${(Number(fen) / 100).toFixed(2)}`;
}
