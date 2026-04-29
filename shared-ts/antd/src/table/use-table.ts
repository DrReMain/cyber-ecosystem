import { useState } from "react"

export type TableSize = "small" | "middle" | "large"

export function useTable() {
  const [tableSize, setTableSize] = useState<TableSize>("middle")
  return { tableSize, setTableSize }
}
