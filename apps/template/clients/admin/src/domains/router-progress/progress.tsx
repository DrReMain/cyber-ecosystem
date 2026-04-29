import { useRouter } from "@tanstack/react-router"
import { useEffect, useRef } from "react"
import { defaultConfig } from "./config"

type Timer = ReturnType<typeof setTimeout>

export function RouterProgress() {
  const router = useRouter()
  const config = defaultConfig
  const containerRef = useRef<HTMLDivElement>(null)
  const barRef = useRef<HTMLDivElement>(null)
  const widthRef = useRef(0)
  const rafRef = useRef(0)
  const finishTimerRef = useRef<Timer | undefined>(undefined)
  const loadingRef = useRef(false)

  useEffect(() => {
    const bar = barRef.current
    const container = containerRef.current
    if (!bar || !container) return

    function startAnimation() {
      const initialWidth =
        config.initialMin +
        Math.random() * (config.initialMax - config.initialMin)
      widthRef.current = initialWidth

      // biome-ignore lint/style/noNonNullAssertion: guarded by parent null check
      const b = bar!
      // biome-ignore lint/style/noNonNullAssertion: guarded by parent null check
      const c = container!

      c.style.display = "block"
      b.style.transition = "none"
      b.style.width = `${initialWidth}%`

      const animate = () => {
        if (widthRef.current >= config.trickleCeiling) return
        const remaining = config.trickleCeiling - widthRef.current
        widthRef.current += remaining * 0.015
        b.style.width = `${widthRef.current}%`
        rafRef.current = requestAnimationFrame(animate)
      }
      rafRef.current = requestAnimationFrame(animate)
    }

    function finishAnimation() {
      // biome-ignore lint/style/noNonNullAssertion: guarded by parent null check
      const b = bar!
      // biome-ignore lint/style/noNonNullAssertion: guarded by parent null check
      const c = container!

      cancelAnimationFrame(rafRef.current)
      b.style.transition = "width 300ms ease-out"
      b.style.width = "100%"

      finishTimerRef.current = setTimeout(() => {
        c.style.display = "none"
        b.style.transition = "none"
        b.style.width = "0%"
        widthRef.current = 0
      }, config.finishDelay)
    }

    const unsubBeforeLoad = router.subscribe("onBeforeLoad", (evt) => {
      if (!evt.pathChanged) return
      if (loadingRef.current) return
      loadingRef.current = true
      cancelAnimationFrame(rafRef.current)
      if (finishTimerRef.current !== undefined) {
        clearTimeout(finishTimerRef.current)
        finishTimerRef.current = undefined
      }
      startAnimation()
    })

    const unsubResolved = router.subscribe("onResolved", (evt) => {
      if (!evt.pathChanged) return
      if (!loadingRef.current) return
      loadingRef.current = false
      finishAnimation()
    })

    return () => {
      unsubBeforeLoad()
      unsubResolved()
      cancelAnimationFrame(rafRef.current)
      if (finishTimerRef.current !== undefined) {
        clearTimeout(finishTimerRef.current)
      }
    }
  }, [router])

  return (
    <div
      aria-hidden="true"
      className="fixed inset-x-0 top-0 z-9999"
      ref={containerRef}
      style={{ display: "none", height: config.height }}
    >
      <style>{`@keyframes router-progress-cycle{to{background-position:300% 0}}`}</style>
      <div
        className="h-full"
        ref={barRef}
        style={{
          width: 0,
          background:
            "linear-gradient(90deg, #f87171, #fb923c, #fbbf24, #a3e635, #34d399, #38bdf8, #818cf8, #c084fc, #f87171)",
          backgroundSize: "300% 100%",
          animation: "router-progress-cycle 6s linear infinite",
          boxShadow:
            "0 0 10px rgba(99,102,241,0.5), 0 0 20px rgba(99,102,241,0.2)",
        }}
      />
    </div>
  )
}
