// Fen amounts ride the wire as bigint (Int64Value flattened); yuan is a
// display concern, never a stored one. Shared by every purchase face —
// the report wall and the pay wall must quote the same numbers.
export function fenToYuan(fen: bigint): string {
  const yuan = Number(fen) / 100;
  return Number.isInteger(yuan) ? String(yuan) : yuan.toFixed((yuan * 10) % 1 === 0 ? 1 : 2);
}
