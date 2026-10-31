import type { AdminSku } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/sku_admin_pb";
import { Input, Modal, Typography } from "antd";
import { useEffect, useState } from "react";
import { m } from "#/paraglide/messages";
import { fenOfYuan, fenToYuan } from "./fen";

interface RepriceModalProps {
  sku: AdminSku | null;
  pending: boolean;
  onCancel: () => void;
  onSubmit: (priceFen: bigint) => void;
}

export function RepriceModal({ sku, pending, onCancel, onSubmit }: Readonly<RepriceModalProps>) {
  const [price, setPrice] = useState("");
  useEffect(() => {
    if (sku) setPrice(fenToYuan(sku.sku?.priceFen).replace(/^¥/, ""));
  }, [sku]);

  const priceFen = fenOfYuan(price);

  return (
    <Modal
      cancelText={m.common_cancel()}
      okButtonProps={{ loading: pending, disabled: priceFen == null || priceFen <= 0n }}
      okText={m.common_save()}
      onCancel={onCancel}
      onOk={() => priceFen && onSubmit(priceFen)}
      open={sku != null}
      title={m.esports_skus_reprice_title({ name: sku?.sku?.name ?? "" })}
    >
      <div className="mb-3 flex items-center gap-2">
        <span className="text-secondary w-24 flex-none text-xs">
          {m.esports_skus_reprice_current()}
        </span>
        <Input onChange={(e) => setPrice(e.target.value)} value={price} />
      </div>
      <Typography.Text className="text-xs" type="warning">
        {m.esports_skus_reprice_note()}
      </Typography.Text>
    </Modal>
  );
}
