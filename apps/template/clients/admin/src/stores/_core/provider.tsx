import { Provider } from "jotai"
import type { DevToolsProps } from "jotai-devtools"
import {
  type ComponentType,
  lazy,
  type PropsWithChildren,
  Suspense,
  useEffect,
  useRef,
  useState,
} from "react"
import { createMainStore } from "./createStore"
import { setApiStore } from "./store-ref"

declare global {
  interface Window {
    __TOGGLE_JOTAI_DEVTOOLS__: () => void
  }
}

if (import.meta.env.DEV) {
  // Side-effect: patches jotai createStore for devtools atom detection.
  // Dead-code eliminated in production builds.
  await import("jotai-devtools")
}

const JotaiDevTools: ComponentType<DevToolsProps> | null = import.meta.env.DEV
  ? lazy(async () => {
      import("jotai-devtools/styles.css")
      const mod = await import("jotai-devtools")
      return { default: mod.DevTools }
    })
  : null

interface IProps {
  initialData?: Record<string, unknown>
}

export function JotaiProvider({
  initialData,
  children,
}: Readonly<PropsWithChildren<IProps>>) {
  const [show, setShow] = useState(false)
  useEffect(() => {
    window.__TOGGLE_JOTAI_DEVTOOLS__ = () => setShow((v) => !v)
  }, [])

  const frozen = useRef(initialData).current
  const store = useRef(createMainStore(frozen)).current

  // Client-only: API interceptors read store state at request time
  setApiStore(store)

  return (
    <Provider store={store}>
      {children}
      {show && JotaiDevTools ? (
        <Suspense fallback={null}>
          <JotaiDevTools
            options={{
              shouldShowPrivateAtoms: false,
              shouldExpandJsonTreeViewInitially: true,
              timeTravelPlaybackInterval: 750,
              snapshotHistoryLimit: Infinity,
            }}
            position="bottom-right"
            store={store}
            theme="dark"
          />
        </Suspense>
      ) : null}
    </Provider>
  )
}
