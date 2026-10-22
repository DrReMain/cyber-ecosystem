import { Image } from "@unpic/react";
import { APP_NAME } from "#/config";
import { m } from "#/paraglide/messages";

const SYSLINE = [
  { key: "KEY1", num: "1", unit: "unit", label: () => "label" },
  { key: "KEY2", num: "2", unit: "unit", label: () => "label" },
  { key: "KEY3", num: "3", unit: "unit", label: () => "label" },
  { key: "KEY4", num: "4", unit: "unit", label: () => "label" },
] as const;

export function BrandCanvas() {
  return (
    <section className="relative flex min-w-0 flex-[1.35] overflow-hidden bg-[#060a14] max-[899px]:fixed max-[899px]:inset-0 max-[899px]:flex-none">
      <div aria-hidden className="absolute inset-0 rtl:-scale-x-100">
        <div className="absolute inset-0 bg-[url('/assets/hero-poster.png')] bg-cover bg-center motion-safe:animate-[nc-drift_60s_ease-in-out_infinite_alternate]" />
      </div>
      <div
        aria-hidden
        className="absolute inset-0 bg-linear-to-r from-[#05070d]/30 via-[#05070d]/55 via-55% to-[#05070d]/92 max-[899px]:from-[#05070d]/72 max-[899px]:via-[#05070d]/60 max-[899px]:to-[#05070d]/72 rtl:bg-linear-to-l"
      />
      <div
        aria-hidden
        className="absolute inset-0 bg-linear-to-t from-[#05070d]/82 via-[#05070d]/15 via-34% to-transparent to-58% max-[899px]:from-[#05070d]/88 max-[899px]:via-[#05070d]/35 max-[899px]:via-40% max-[899px]:to-[#05070d]/55"
      />
      <div
        aria-hidden
        className="via-primary/30 absolute inset-x-0 z-1 h-px bg-linear-to-r from-transparent to-transparent motion-safe:animate-[nc-vscan_9s_linear_infinite]"
      />

      <div className="relative z-2 flex flex-1 flex-col p-[clamp(20px,4.5vw,56px)] max-[899px]:justify-start max-[899px]:pt-[clamp(28px,6vh,56px)] max-[899px]:pb-30">
        <div className="flex items-center gap-2.75 font-mono text-[12.5px] font-semibold tracking-[0.2em]">
          <Image
            alt=""
            aria-hidden
            className="size-5.5 flex-none"
            height={22}
            src="/mark-accent.svg"
            width={22}
          />
          <span>{APP_NAME}</span>
        </div>

        <div className="my-auto max-w-140 max-[899px]:mt-6.5 max-[899px]:mb-0 max-[899px]:max-w-155">
          <span className="border-primary/35 bg-primary/6 text-primary inline-flex items-center gap-2.5 rounded-full border px-3.5 py-1.75 font-mono text-[11px] tracking-[0.2em] motion-safe:animate-[nc-pulse_3s_ease-in-out_infinite]">
            <span className="bg-primary shadow-primary size-1.5 flex-none rounded-full shadow-[0_0_8px]" />
            {`SYSTEM READY`}
          </span>

          <h1 className="mt-6 mb-3.5 text-[clamp(34px,4.6vw,58px)] leading-[1.22] font-extrabold max-[899px]:text-[clamp(30px,8vw,44px)]">
            {m.login_default_hero_title_a()}
            <span className="via-primary bg-linear-[110deg] from-white from-5% via-45% to-[#8b5cf6] to-92% bg-clip-text text-transparent drop-shadow-[0_6px_32px_rgba(139,92,246,0.2)] rtl:bg-linear-[70deg]">
              {m.login_default_hero_title_b()}
            </span>
          </h1>

          <p className="max-w-[40ch] text-[15px] leading-[1.95] text-white/55">
            {m.login_default_hero_sub()}
          </p>
        </div>

        <div
          aria-hidden
          className="grid grid-cols-2 gap-x-6 gap-y-4 border-t border-white/14 pt-5 max-[899px]:hidden"
        >
          {SYSLINE.map((item) => (
            <div className="border-s border-white/6 ps-3.5" key={item.key}>
              <div className="font-mono text-[15px] font-semibold">
                {item.num}
                {item.unit ? <span className="text-primary">{` ${item.unit}`}</span> : null}
              </div>
              <div className="mt-1 font-mono text-[9.5px] tracking-[0.18em] text-white/35 rtl:tracking-normal">
                {`${item.key} · ${item.label()}`}
              </div>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}
