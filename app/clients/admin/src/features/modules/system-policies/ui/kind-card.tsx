import { theme } from "antd";
import { Check } from "lucide-react";
import type { CSSProperties, ReactNode } from "react";

interface KindCardProps {
  selected: boolean;
  icon: ReactNode;
  title: string;
  desc: string;
  onSelect: () => void;
}

export function KindCard({ selected, icon, title, desc, onSelect }: Readonly<KindCardProps>) {
  const { token } = theme.useToken();
  const style: CSSProperties = selected
    ? { borderColor: token.colorPrimary, background: token.colorPrimaryBg }
    : { borderColor: token.colorBorderSecondary };
  return (
    <button
      aria-pressed={selected}
      className="flex flex-1 flex-col gap-1 rounded-lg border p-3 text-start transition-colors"
      onClick={onSelect}
      style={style}
      type="button"
    >
      <span className="flex items-center gap-2 font-medium text-[13px]">
        {icon}
        {title}
        {selected ? <Check className="ms-auto" color={token.colorPrimary} size={14} /> : null}
      </span>
      <span className="text-[12px] text-ink-tertiary">{desc}</span>
    </button>
  );
}
