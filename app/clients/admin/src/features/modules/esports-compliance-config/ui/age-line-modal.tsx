import { Alert, InputNumber, Modal } from "antd";
import { useEffect, useState } from "react";
import { m } from "#/paraglide/messages";

interface AgeLineModalProps {
  initial: number;
  open: boolean;
  onClose: () => void;
  onConfirm: (age: number) => void;
  submitting: boolean;
}

// The age line's double confirm: the modal is the second yes, drawn red
// because the flip is the panel the prototype marks 敏感 — and the write
// lands in the audit log like every admin action.
export function AgeLineModal({
  initial,
  open,
  onClose,
  onConfirm,
  submitting,
}: Readonly<AgeLineModalProps>) {
  const [age, setAge] = useState<number>(initial);

  // Each opening starts from the displayed line — a stale draft never
  // survives a reopen.
  useEffect(() => {
    if (open) {
      setAge(initial);
    }
  }, [initial, open]);

  const valid = Number.isInteger(age) && age >= 1 && age <= 120;

  return (
    <Modal
      cancelText={m.common_cancel()}
      okButtonProps={{ danger: true, disabled: !valid, loading: submitting }}
      okText={m.esports_compliance_config_age_modal_ok()}
      onCancel={onClose}
      onOk={() => onConfirm(age)}
      open={open}
      title={m.esports_compliance_config_age_modal_title()}
    >
      <div className="mb-3 flex items-center gap-3">
        <span className="text-text-secondary text-sm">
          {m.esports_compliance_config_age_modal_label()}
        </span>
        <InputNumber max={120} min={1} onChange={(v) => setAge(v ?? 0)} precision={0} value={age} />
      </div>
      <Alert showIcon title={m.esports_compliance_config_age_modal_warn()} type="error" />
    </Modal>
  );
}
