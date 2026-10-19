import { Sender, XProvider } from "@ant-design/x";
import xZhCN from "@ant-design/x/es/locale/zh_CN";
import { Grid } from "antd";
import { MessageSquareMore } from "lucide-react";
import { type ReactNode, useState } from "react";
import { useAntdLocale } from "#/domains/antd";
import { m } from "#/paraglide/messages";
import { ModelChip } from "./model-chip";

interface RailSlot {
  mode: "drawer" | "static";
  open: boolean;
  onClose: () => void;
}

interface ChatViewProps {
  rail: (slot: RailSlot) => ReactNode;
  content: ReactNode;
  draft: string;
  onDraftChange: (value: string) => void;
  onSend: (text: string) => void;
  streaming: boolean;
  onStop: () => void;
  model: string | undefined;
  models: string[];
  modelsLoading: boolean;
  onModelChange: (model: string) => void;
}

export function ChatView({
  rail,
  content,
  draft,
  onDraftChange,
  onSend,
  streaming,
  onStop,
  model,
  models,
  modelsLoading,
  onModelChange,
}: Readonly<ChatViewProps>) {
  const [railOpen, setRailOpen] = useState(false);
  const narrow = Grid.useBreakpoint().lg === false;
  const { direction, locale, tag } = useAntdLocale();
  const xPack = tag === "zh-CN" ? xZhCN : undefined;

  const modelChip = (
    <ModelChip
      loading={modelsLoading}
      model={model}
      models={models}
      onModelChange={onModelChange}
    />
  );
  const renderActions = (actionNode: ReactNode) => (
    <div className="flex items-center justify-end gap-2 pt-1.5 pb-0.5">
      {modelChip}
      {actionNode}
    </div>
  );

  return (
    <XProvider direction={direction} locale={xPack ? { ...locale, ...xPack } : locale}>
      <div className="flex min-h-0 grow flex-row">
        <div
          className={
            narrow
              ? "size-0 shrink-0 overflow-hidden"
              : "flex w-60 shrink-0 flex-col overflow-hidden border-line-soft border-e bg-container"
          }
        >
          {rail({
            mode: narrow ? "drawer" : "static",
            open: railOpen,
            onClose: () => setRailOpen(false),
          })}
        </div>
        <div className="flex min-w-0 grow flex-col bg-canvas">
          <div className="mx-auto flex min-h-0 w-full grow flex-col px-4">
            {content}
            <div className="pt-2 pb-4">
              <Sender
                autoSize={{ minRows: 1, maxRows: 8 }}
                disabled={model === undefined}
                footer={narrow ? renderActions : undefined}
                loading={streaming}
                onCancel={onStop}
                onChange={onDraftChange}
                onSubmit={onSend}
                placeholder={m.agents_chat_placeholder()}
                prefix={
                  narrow ? (
                    <button
                      aria-label={m.agents_chat_rail_history()}
                      className="me-1 inline-flex size-8 shrink-0 items-center justify-center rounded-control border border-line-soft text-ink-secondary transition-colors hover:border-line hover:text-ink"
                      onClick={() => setRailOpen(true)}
                      type="button"
                    >
                      <MessageSquareMore className="size-4" />
                    </button>
                  ) : undefined
                }
                suffix={narrow ? false : renderActions}
                value={draft}
              />
            </div>
          </div>
        </div>
      </div>
    </XProvider>
  );
}
