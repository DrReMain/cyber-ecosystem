import { useNavigate } from "@tanstack/react-router";
import { BarChart3, House, MessageCircle, User } from "lucide-react";
import { toast } from "sonner";
import { m } from "#/paraglide/messages";

export type TabKey = "home" | "me";

interface TabSpec {
  key: TabKey | "records" | "messages";
  label: () => string;
  icon: typeof House;
  to: string;
  live: boolean;
}

const TABS: readonly TabSpec[] = [
  { key: "home", label: () => m.common_tab_home(), icon: House, to: "/", live: true },
  // Records and messages stay skeleton tabs — the shape is the prototype's;
  // the faces are not ours to build yet.
  {
    key: "records",
    label: () => m.common_tab_records(),
    icon: BarChart3,
    to: "",
    live: false,
  },
  {
    key: "messages",
    label: () => m.common_tab_messages(),
    icon: MessageCircle,
    to: "",
    live: false,
  },
  { key: "me", label: () => m.common_tab_me(), icon: User, to: "/profile", live: true },
];

/**
 * Bottom navigation chrome (prototype _shell TabBar): four slots, two live.
 * Mounted by the tab-owning pages (home, profile overview); fixed to the
 * viewport with the safe-area inset, blurred over ambient content. On ≥md
 * the app is a centered frame — the bar anchors to the frame's bottom
 * instead of the viewport's.
 */
export function TabBar({ active }: Readonly<{ active: TabKey }>) {
  const navigate = useNavigate();
  return (
    <nav className="border-line-soft bg-surface-card/95 fixed inset-x-0 bottom-0 z-40 border-t backdrop-blur-md md:absolute">
      <div className="mx-auto flex h-14 max-w-md items-stretch">
        {TABS.map((tab) => {
          const Icon = tab.icon;
          const on = tab.key === active;
          const cls = on
            ? "text-focus font-semibold"
            : tab.live
              ? "text-ink-tertiary"
              : "text-ink-tertiary/50";
          return (
            <button
              aria-current={on ? "page" : undefined}
              className={`flex flex-1 flex-col items-center justify-center gap-0.5 text-[10px] active:opacity-60 ${cls}`}
              key={tab.key}
              onClick={() => {
                if (!tab.live) {
                  toast.info(m.common_soon());
                  return;
                }
                if (!on) {
                  void navigate({ to: tab.to });
                }
              }}
              type="button"
            >
              <Icon aria-hidden size={19} />
              {tab.label()}
            </button>
          );
        })}
      </div>
      <div className="safe-hb" />
    </nav>
  );
}
