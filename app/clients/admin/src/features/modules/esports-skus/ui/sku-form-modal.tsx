import { SkuKind } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/sku_pb";
import { Input, Modal, Select, Typography } from "antd";
import { useEffect, useState } from "react";
import { m } from "#/paraglide/messages";
import { fenOfYuan } from "./fen";

export interface SkuDraft {
  name: string;
  kind: SkuKind;
  priceFen: bigint;
  periodDays?: number;
  description?: string;
}

interface SkuFormModalProps {
  open: boolean;
  pending: boolean;
  onCancel: () => void;
  onSubmit: (draft: SkuDraft) => void;
}

export function SkuFormModal({ open, pending, onCancel, onSubmit }: Readonly<SkuFormModalProps>) {
  const [name, setName] = useState("");
  const [kind, setKind] = useState<SkuKind>(SkuKind.TRIAL);
  const [price, setPrice] = useState("");
  const [period, setPeriod] = useState("");
  const [desc, setDesc] = useState("");
  useEffect(() => {
    if (!open) return;
    setName("");
    setKind(SkuKind.TRIAL);
    setPrice("");
    setPeriod("");
    setDesc("");
  }, [open]);

  const priceFen = fenOfYuan(price);
  const periodDays = /^\d+$/.test(period.trim()) ? Number.parseInt(period.trim(), 10) : undefined;
  const valid =
    name.trim().length > 0 &&
    priceFen != null &&
    priceFen > 0n &&
    (kind !== SkuKind.MEMBERSHIP || (periodDays != null && periodDays > 0));

  return (
    <Modal
      cancelText={m.common_cancel()}
      okButtonProps={{ loading: pending, disabled: !valid }}
      okText={m.common_save()}
      onCancel={onCancel}
      onOk={() =>
        onSubmit({
          name: name.trim(),
          kind,
          priceFen: priceFen ?? 0n,
          ...(kind === SkuKind.MEMBERSHIP && periodDays ? { periodDays } : {}),
          ...(desc.trim() ? { description: desc.trim() } : {}),
        })
      }
      open={open}
      title={m.esports_skus_new_title()}
    >
      <div className="mb-3 flex items-center gap-2">
        <span className="text-secondary w-24 flex-none text-xs">{m.esports_skus_new_name()}</span>
        <Input
          onChange={(e) => setName(e.target.value)}
          placeholder={m.esports_skus_new_name_ph()}
          value={name}
        />
      </div>
      <div className="mb-3 flex items-center gap-2">
        <span className="text-secondary w-24 flex-none text-xs">{m.esports_skus_new_kind()}</span>
        <Select
          className="w-32"
          onChange={(v) => setKind(v)}
          options={[
            { value: SkuKind.TRIAL, label: m.esports_skus_kind_trial() },
            { value: SkuKind.MEMBERSHIP, label: m.esports_skus_kind_membership() },
          ]}
          value={kind}
        />
      </div>
      <div className="mb-3 flex items-center gap-2">
        <span className="text-secondary w-24 flex-none text-xs">{m.esports_skus_new_price()}</span>
        <Input
          onChange={(e) => setPrice(e.target.value)}
          placeholder={m.esports_skus_new_price_ph()}
          value={price}
        />
      </div>
      {kind === SkuKind.MEMBERSHIP ? (
        <div className="mb-3 flex items-center gap-2">
          <span className="text-secondary w-24 flex-none text-xs">
            {m.esports_skus_new_period()}
          </span>
          <Input
            onChange={(e) => setPeriod(e.target.value)}
            placeholder={m.esports_skus_new_period_ph()}
            suffix={m.esports_skus_days()}
            value={period}
          />
        </div>
      ) : null}
      <div className="mb-3 flex items-center gap-2">
        <span className="text-secondary w-24 flex-none text-xs">{m.esports_skus_new_desc()}</span>
        <Input onChange={(e) => setDesc(e.target.value)} value={desc} />
      </div>
      <Typography.Text className="text-xs" type="secondary">
        {m.esports_skus_new_note()}
      </Typography.Text>
    </Modal>
  );
}
