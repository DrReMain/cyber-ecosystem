import { InputNumber, Modal } from "antd";
import { useEffect, useState } from "react";
import { m } from "#/paraglide/messages";

interface RemindModalProps {
  initial: number;
  open: boolean;
  onClose: () => void;
  onConfirm: (days: number) => void;
  submitting: boolean;
}

// The renewal reminder's single confirm — a plain scheduling knob, unlike
// the age line's red double confirm two cards up.
export function RemindModal({
  initial,
  open,
  onClose,
  onConfirm,
  submitting,
}: Readonly<RemindModalProps>) {
  const [days, setDays] = useState<number>(initial);

  // Each opening starts from the displayed value — a stale draft never
  // survives a reopen.
  useEffect(() => {
    if (open) {
      setDays(initial);
    }
  }, [initial, open]);

  const valid = Number.isInteger(days) && days >= 1 && days <= 90;

  return (
    <Modal
      cancelText={m.common_cancel()}
      okButtonProps={{ disabled: !valid, loading: submitting }}
      okText={m.esports_compliance_config_remind_modal_ok()}
      onCancel={onClose}
      onOk={() => onConfirm(days)}
      open={open}
      title={m.esports_compliance_config_remind_modal_title()}
    >
      <div className="flex items-center gap-3">
        <span className="text-text-secondary text-sm">
          {m.esports_compliance_config_remind_modal_label()}
        </span>
        <InputNumber
          max={90}
          min={1}
          onChange={(v) => setDays(v ?? 0)}
          precision={0}
          value={days}
        />
      </div>
    </Modal>
  );
}
