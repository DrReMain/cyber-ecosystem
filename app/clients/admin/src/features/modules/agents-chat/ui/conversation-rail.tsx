import { Conversations } from "@ant-design/x";
import { App, Button, Drawer, Input, Modal } from "antd";
import clsx from "clsx";
import { Ellipsis, Loader2, MessageSquare, Pencil, Plus, Trash2 } from "lucide-react";
import { useState } from "react";
import { m } from "#/paraglide/messages";
import type { Conversation } from "../conversations";
import { conversationTitle } from "../conversations";

interface ConversationRailProps {
  conversations: Conversation[];
  activeId: string;
  streamingIds: string[];
  mode: "drawer" | "static";
  open: boolean;
  hasMore: boolean;
  loadingMore: boolean;
  onClose: () => void;
  onSelect: (id: string) => void;
  onNew: () => void;
  onDelete: (id: string) => void;
  onRename: (id: string, title: string) => void;
  onLoadMore: () => void;
}

export function ConversationRail({
  conversations,
  activeId,
  streamingIds,
  mode,
  open,
  hasMore,
  loadingMore,
  onClose,
  onSelect,
  onNew,
  onDelete,
  onRename,
  onLoadMore,
}: Readonly<ConversationRailProps>) {
  const { modal } = App.useApp();
  const [renaming, setRenaming] = useState<{ id: string; value: string } | null>(null);

  const confirmDelete = (id: string) => {
    modal.confirm({
      cancelText: m.common_cancel(),
      cancelButtonProps: { color: "default", variant: "filled" },
      okButtonProps: { color: "danger", variant: "filled" },
      okText: m.common_confirm(),
      onOk: () => onDelete(id),
      title: m.agents_chat_rail_delete_confirm(),
    });
  };

  const items = conversations.map((conversation) => ({
    key: conversation.id,
    label: conversation.title ?? conversationTitle(conversation),
    // Streaming swaps the glyph in place so labels never shift.
    icon: streamingIds.includes(conversation.id) ? (
      <Loader2 className="size-3.5 animate-spin text-primary" />
    ) : (
      <MessageSquare className="size-3.5" />
    ),
  }));

  const body = (
    <div className="flex h-full min-h-0 flex-col">
      <div className="shrink-0 p-2">
        <Button
          block
          color="primary"
          icon={<Plus className="size-3.5" />}
          onClick={() => {
            onNew();
            // Drawer form dismisses on any action that picks a conversation.
            onClose();
          }}
          variant="filled"
        >
          {m.agents_chat_rail_new()}
        </Button>
      </div>
      <div className={clsx("min-h-0 pt-1", conversations.length > 0 && "grow overflow-y-auto")}>
        <Conversations
          activeKey={activeId}
          className="px-2 pb-2"
          items={items}
          menu={(item) => {
            const source = conversations.find((c) => c.id === item.key);
            return {
              trigger: (
                <Ellipsis
                  aria-label={m.agents_chat_rail_more()}
                  className="ant-conversations-menu-icon size-4 cursor-pointer [@media(hover:none)]:opacity-60"
                  onClick={(e) => e.stopPropagation()}
                />
              ),
              items: [
                ...(source?.title !== undefined
                  ? [
                      {
                        icon: <Pencil className="size-3.5" />,
                        key: "rename",
                        label: m.agents_chat_rail_rename(),
                      },
                    ]
                  : []),
                {
                  icon: <Trash2 className="size-3.5" />,
                  key: "delete",
                  label: m.agents_chat_rail_delete(),
                },
              ],
              onClick: (info) => {
                if (info.key === "delete") confirmDelete(item.key);
                if (info.key === "rename" && source) {
                  setRenaming({ id: item.key, value: source.title ?? conversationTitle(source) });
                }
              },
            };
          }}
          onActiveChange={(id) => {
            onSelect(id);
            onClose();
          }}
        />
        {hasMore && conversations.length > 0 && (
          <div className="flex justify-center py-2">
            <Button
              color="default"
              loading={loadingMore}
              onClick={onLoadMore}
              size="small"
              variant="text"
            >
              {m.agents_chat_rail_load_more()}
            </Button>
          </div>
        )}
      </div>
      {conversations.length === 0 && !loadingMore && (
        <p className="px-4 py-3 text-[12px] text-ink-quaternary leading-relaxed">
          {m.agents_chat_rail_empty()}
        </p>
      )}
      <Modal
        cancelButtonProps={{ color: "default", variant: "filled" }}
        cancelText={m.common_cancel()}
        okButtonProps={{ color: "primary", variant: "filled" }}
        okText={m.common_confirm()}
        onCancel={() => setRenaming(null)}
        onOk={() => {
          const title = renaming?.value.trim();
          if (renaming && title) onRename(renaming.id, title);
          setRenaming(null);
        }}
        open={renaming !== null}
        title={m.agents_chat_rail_rename()}
      >
        <Input
          maxLength={64}
          onChange={(e) =>
            setRenaming((prev) => (prev ? { ...prev, value: e.target.value } : prev))
          }
          value={renaming?.value ?? ""}
        />
      </Modal>
    </div>
  );

  if (mode === "drawer") {
    return (
      <Drawer
        closable={false}
        onClose={onClose}
        open={open}
        placement="left"
        size={288}
        styles={{ body: { padding: 0 } }}
      >
        {body}
      </Drawer>
    );
  }
  // Static geometry (column width, border) belongs to the layout's rail slot.
  return body;
}
