import { Think } from "@ant-design/x";
import { useState } from "react";
import { m } from "#/paraglide/messages";

export function ReasoningStrip({
  reasoning,
  streaming,
}: Readonly<{ reasoning: string; streaming: boolean }>) {
  const [pinned, setPinned] = useState<boolean | null>(null);
  return (
    <Think
      blink={streaming}
      expanded={pinned ?? streaming}
      loading={streaming}
      onExpand={setPinned}
      title={streaming ? m.agents_chat_thinking() : m.agents_chat_thought()}
    >
      <div className="whitespace-pre-wrap text-[12px]">{reasoning}</div>
    </Think>
  );
}
