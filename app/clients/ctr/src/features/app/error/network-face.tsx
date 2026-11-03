import { Link, useRouter } from "@tanstack/react-router";
import { RadioTower } from "lucide-react";
import { m } from "#/paraglide/messages";

interface NetworkFaceProps {
  readonly homePath: string;
  readonly http: number;
  readonly kind: string;
  /** The host boundary's own retry — defaults to a router invalidate. */
  readonly onRetry?: () => void;
}

/**
 * The network-error face (prototype error.html, Q1-③ ruling): the one
 * error kind a prototype oracle covers renders the warm card — the dish
 * icon, the plain failure word, the progress-preserved reassurance (true
 * behavior: attempts submit per game, the session persists server-side),
 * and three exits. Both error hosts render it: the root ErrorPage
 * (loader-thrown errors) and the inline ErrorFallback (component-thrown
 * ones) — whichever boundary catches the failure, the member sees this
 * card. The other kinds keep their idiom; no prototype speaks for them.
 */
export function NetworkErrorFace({ homePath, http, kind, onRetry }: Readonly<NetworkFaceProps>) {
  const router = useRouter();
  const retry = onRetry ?? (() => router.invalidate());
  return (
    <div className="w-full max-w-md">
      <div className="border-line bg-surface-card shadow-card rounded-card border px-5 py-[30px] text-center backdrop-blur-md">
        <RadioTower aria-hidden className="text-ink-tertiary mx-auto" size={34} strokeWidth={1.5} />
        <div className="text-ink mt-3 text-[14.5px] font-extrabold">
          {m.error_network_card_title()}
        </div>
        <p className="text-ink-muted mt-2.25 text-[12.5px] leading-[1.85]">
          {m.error_network_sub_lead()}
          <span className="font-bold">{m.error_network_sub_keep()}</span>
          {m.error_network_sub_tail()}
        </p>
        <div className="border-line text-ink-tertiary rounded-control mt-3.5 border px-2.75 py-2.25 text-start font-mono text-[10.5px] leading-[1.7]">
          {m.error_code()} <span dir="ltr">{`${http} · ${kind.toUpperCase()}`}</span>
        </div>
      </div>
      <button
        className="bg-cta-primary shadow-cta-primary rounded-control mt-3 w-full px-4 py-[14px] text-[15px] font-bold text-white active:opacity-80"
        onClick={retry}
        type="button"
      >
        {m.common_retry()}
      </button>
      <button
        className="bg-surface-input text-ink-secondary rounded-control mt-2.25 w-full border-[1.5px] border-[#4b5563] px-4 py-[11px] text-center text-[12.5px] font-medium active:opacity-60"
        onClick={() => router.history.back()}
        type="button"
      >
        {m.error_back()}
      </button>
      <Link
        className="bg-surface-input text-ink-secondary rounded-control mt-2.25 block w-full border-[1.5px] border-[#4b5563] px-4 py-[11px] text-center text-[12.5px] font-medium active:opacity-60"
        to={homePath as never}
      >
        {m.error_home()}
      </Link>
    </div>
  );
}
