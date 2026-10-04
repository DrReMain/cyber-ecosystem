import { Prompts, Welcome } from "@ant-design/x";
import { Button } from "antd";
import { FileText, GitCompareArrows, PenLine } from "lucide-react";
import { m } from "#/paraglide/messages";

interface EmptyStateProps {
  model?: string;
  onSuggest?: (text: string) => void;
  onConfigure?: () => void;
}

export function EmptyState({ model, onSuggest, onConfigure }: Readonly<EmptyStateProps>) {
  const suggestions = [
    {
      icon: FileText,
      key: m.agents_chat_suggest_1_title(),
      description: m.agents_chat_suggest_1_desc(),
    },
    {
      icon: GitCompareArrows,
      key: m.agents_chat_suggest_2_title(),
      description: m.agents_chat_suggest_2_desc(),
    },
    {
      icon: PenLine,
      key: m.agents_chat_suggest_3_title(),
      description: m.agents_chat_suggest_3_desc(),
    },
  ];

  return (
    <div className="relative flex grow items-center justify-center overflow-y-auto overflow-x-hidden">
      <div
        aria-hidden
        className="absolute top-0 left-1/2 h-140 w-140 -translate-x-1/2 -translate-y-35 rounded-full bg-primary/10 blur-3xl"
      />
      <div className="relative flex max-w-2xl flex-col items-center px-6 py-8 text-center">
        <Welcome
          classNames={{
            description: "max-w-md text-[13.5px] leading-relaxed",
            title: "font-semibold text-[22px]",
          }}
          description={m.agents_chat_welcome_desc()}
          title={m.agents_chat_welcome_title()}
          variant="borderless"
        />
        {onConfigure ? (
          <Button className="mt-5" color="primary" onClick={onConfigure} variant="filled">
            {m.agents_chat_no_config_action()}
          </Button>
        ) : (
          <>
            <span className="mt-4 inline-flex h-6 items-center gap-1.5 rounded-control-sm border border-line-soft px-2 text-[11.5px] text-primary">
              <i className="size-1.5 rounded-full bg-success" />
              {m.agents_chat_welcome_ready({ model: model ?? "" })}
            </span>
            <Prompts
              className="mt-6 w-full [&_.ant-prompts-item]:w-full"
              items={suggestions.map(({ icon: Icon, key, description }) => ({
                key,
                description,
                icon: (
                  <span className="flex size-7 items-center justify-center rounded-control-sm bg-fill-quaternary text-primary">
                    <Icon className="size-4" />
                  </span>
                ),
                label: key,
              }))}
              onItemClick={({ data }) => onSuggest?.(data.key)}
              title={
                <span className="text-[12px] text-ink-quaternary">
                  {m.agents_chat_suggest_label()}
                </span>
              }
              vertical
            />
          </>
        )}
        <div className="mt-6 text-[11.5px] text-ink-quaternary">
          {m.agents_chat_welcome_privacy()}
        </div>
      </div>
    </div>
  );
}
