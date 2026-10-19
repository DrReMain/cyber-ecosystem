import { Loader2 } from "lucide-react";
import { EmptyState } from "./empty-state";

export function EmptyConversation({
  messagesPending,
  model,
  onSuggest,
}: Readonly<{
  messagesPending: boolean;
  model: string | undefined;
  onSuggest: (text: string) => void;
}>) {
  if (messagesPending) {
    return (
      <div className="flex grow items-center justify-center">
        <Loader2 className="size-5 animate-spin text-ink-quaternary" />
      </div>
    );
  }
  return <EmptyState model={model} onSuggest={onSuggest} />;
}
