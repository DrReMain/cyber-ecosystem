import { Input, Modal } from "antd";
import { useEffect, useState } from "react";
import { m } from "#/paraglide/messages";

interface PromptModalProps {
  initial: string | null;
  open: boolean;
  onClose: () => void;
  onSubmit: (text: string) => void;
  submitting: boolean;
}

// One prompt at a time: the opener names which row, this modal carries
// only the new literal copy — 200 runes is the wire's ceiling. Each
// opening starts from the standing override (null = no override yet) —
// an edit must show the operator what they are about to replace.
export function PromptModal({
  initial,
  open,
  onClose,
  onSubmit,
  submitting,
}: Readonly<PromptModalProps>) {
  const [text, setText] = useState("");

  useEffect(() => {
    if (open) {
      setText(initial ?? "");
    }
  }, [initial, open]);

  return (
    <Modal
      cancelText={m.common_cancel()}
      okButtonProps={{ disabled: text.trim().length === 0, loading: submitting }}
      okText={m.esports_compliance_config_prompts_save()}
      onCancel={onClose}
      onOk={() => onSubmit(text.trim())}
      open={open}
      title={m.esports_compliance_config_prompts_modal_title()}
    >
      <Input
        maxLength={200}
        onChange={(e) => setText(e.target.value)}
        placeholder={m.esports_compliance_config_prompts_placeholder()}
        showCount
        value={text}
      />
    </Modal>
  );
}
