import { Provider } from "jotai"
import type { PropsWithChildren } from "react"

export function AppProvider({ children }: Readonly<PropsWithChildren>) {
  return <Provider>{children}</Provider>
}
