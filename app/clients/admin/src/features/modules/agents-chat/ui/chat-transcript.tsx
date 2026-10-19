import type { BubbleItemType } from "@ant-design/x";
import { Bubble } from "@ant-design/x";
import { Avatar, Button } from "antd";
import { Sparkles } from "lucide-react";
import { m } from "#/paraglide/messages";
import type { ChatTurn } from "../conversations";
import { MarkdownContent } from "./markdown-content";
import { ReasoningStrip } from "./reasoning-strip";

interface ChatTranscriptProps {
  turns: ReadonlyArray<ChatTurn>;
  streaming: boolean;
  canLoadEarlier: boolean;
  onLoadEarlier: () => void;
}

export function ChatTranscript({
  turns,
  streaming,
  canLoadEarlier,
  onLoadEarlier,
}: Readonly<ChatTranscriptProps>) {
  const streamingKey = streaming ? turns[turns.length - 1]?.key : undefined;

  const chatItems: BubbleItemType[] = turns.map((turn) => {
    const turnStreaming = turn.key === streamingKey;
    const reasoningActive = turnStreaming && turn.content === "";
    return turn.role === "assistant"
      ? {
          key: turn.key,
          role: "assistant",
          content: <MarkdownContent content={turn.content} streaming={turnStreaming} />,
          header:
            turn.reasoning === "" ? undefined : (
              <ReasoningStrip reasoning={turn.reasoning} streaming={reasoningActive} />
            ),
          streaming: turnStreaming,
        }
      : { key: turn.key, role: "user", content: turn.content };
  });
  const items = canLoadEarlier
    ? [
        {
          key: "load-earlier",
          role: "system",
          content: "",
          contentRender: () => (
            <Button color="default" onClick={onLoadEarlier} size="small" variant="text">
              {m.agents_chat_load_earlier()}
            </Button>
          ),
        },
        ...chatItems,
      ]
    : chatItems;

  return (
    <Bubble.List
      className="min-h-0 grow pt-6 pb-4"
      items={items}
      role={{
        assistant: {
          avatar: <Avatar icon={<Sparkles className="size-3.5" />} shape="square" size={28} />,
          placement: "start",
          variant: "borderless",
        },
        user: {
          placement: "end",
          styles: { content: { whiteSpace: "pre-wrap" } },
          variant: "filled",
        },
      }}
    />
  );
}
