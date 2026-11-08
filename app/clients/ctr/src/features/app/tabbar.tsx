import { useNavigate } from "@tanstack/react-router";
import { BarChart3, House, MessageCircle, User } from "lucide-react";
import { toast } from "sonner";
import { useUnreadCount } from "#/features/app/unread";
import { m } from "#/paraglide/messages";

export type TabKey = "home" | "me" | "messages";

interface TabSpec {
  key: TabKey | "records";
  label: () => string;
  icon: typeof House;
  to: string;
  live: boolean;
}

const TABS: readonly TabSpec[] = [
  { key: "home", label: () => m.common_tab_home(), icon: House, to: "/", live: true },
  // Records stays a skeleton tab — the shape is the prototype's; the
  // test-records face is the M2 remainder's, not this window's.
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
    to: "/messages",
    live: true,
  },
  { key: "me", label: () => m.common_tab_me(), icon: User, to: "/profile", live: true },
];

/**
 * Bottom navigation chrome (prototype _shell TabBar): four slots, three
 * live. Mounted by the tab-owning pages (home, messages, profile
 * overview) through the layout's chrome slot. Fixed to the viewport with
 * the safe-area inset on phones; on ≥md the app is a centered frame and
 * the bar is a plain flow child of the frame — docked to the frame
 * bottom, never the scroll extent (an in-scroller absolute rides below
 * the fold mid-scroll). The messages slot carries the letterbox's unread
 * dot — the server owns read state, so the badge asks the server and
 * mark-read invalidations walk it back to zero; the accessible name
 * carries the count so the dot is not a sight-only signal.
 */
export function TabBar({ active }: Readonly<{ active: TabKey }>) {
  const navigate = useNavigate();
  const unread = useUnreadCount();
  const unreadUp = unread !== undefined && unread > 0;
  return (
    <nav className="border-line-soft bg-surface-card/95 fixed inset-x-0 bottom-0 z-40 border-t backdrop-blur-md md:static">
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
              aria-label={
                tab.key === "messages" && unreadUp
                  ? `${m.common_tab_messages()} ${unread} ${m.messages_unread_unit()}`
                  : undefined
              }
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
              <span className="relative">
                <Icon aria-hidden size={19} />
                {tab.key === "messages" && unreadUp ? (
                  <span
                    aria-hidden
                    className="absolute -top-0.5 -right-0.75 size-1.5 rounded-full bg-[#EF4444]"
                  />
                ) : null}
              </span>
              {tab.label()}
            </button>
          );
        })}
      </div>
      <div className="safe-hb" />
    </nav>
  );
}
