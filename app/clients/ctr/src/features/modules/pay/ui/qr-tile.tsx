import { toDataURL } from "qrcode";
import { useEffect, useState } from "react";
import { m } from "#/paraglide/messages";

// The native channel's continuation (the gateway's CredentialQRCode): the
// code_url payload becomes a scannable tile. The pad is white with its
// own quiet zone — the app's dark card would swallow the modules — and
// an encoder refusal falls back to the payload text rather than
// pretending a code that is not there.
export function QrTile({ value }: Readonly<{ value: string }>) {
  const [src, setSrc] = useState<string | null>(null);
  const [failed, setFailed] = useState(false);

  useEffect(() => {
    let alive = true;
    setSrc(null);
    setFailed(false);
    toDataURL(value, { margin: 1, width: 360 })
      .then((url) => {
        if (alive) {
          setSrc(url);
        }
      })
      .catch(() => {
        if (alive) {
          setFailed(true);
        }
      });
    return () => {
      alive = false;
    };
  }, [value]);

  if (failed) {
    return (
      <div className="text-ink-muted max-w-65 text-center text-[10.5px] leading-[1.6] break-all">
        {value}
      </div>
    );
  }
  return (
    <div className="flex h-46 w-46 items-center justify-center rounded-[14px] bg-white p-2.5">
      {src !== null ? <img alt={m.pay_qr_alt()} className="size-full" src={src} /> : null}
    </div>
  );
}
