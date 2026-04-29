export function TailwindIndicator() {
  if (!import.meta.env.DEV) return null

  return (
    <div className="fixed right-1 bottom-1 z-50 rounded bg-black/60 px-1.5 py-0.5 font-mono text-white text-xs">
      <div className="block sm:hidden">xs</div>
      <div className="hidden sm:block md:hidden">sm</div>
      <div className="hidden md:block lg:hidden">md</div>
      <div className="hidden lg:block xl:hidden">lg</div>
      <div className="hidden xl:block 2xl:hidden">xl</div>
      <div className="hidden 2xl:block">2xl</div>
    </div>
  )
}
