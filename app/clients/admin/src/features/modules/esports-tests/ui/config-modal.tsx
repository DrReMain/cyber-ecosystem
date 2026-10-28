import { useMutation, useQuery } from "@connectrpc/connect-query";
import {
  getTestConfig,
  updateTestConfig,
} from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/test_admin-TestAdminService_connectquery";
import type { Test } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/test_pb";
import { Alert, Button, Form, Input, InputNumber, Modal, Skeleton } from "antd";
import { useEffect } from "react";
import { m } from "#/paraglide/messages";

interface ConfigModalProps {
  test: Test;
  onClose: () => void;
}

interface ParamForm {
  key: string;
  value: string;
}

// The engine bag replaces whole on submit — the form always loads the
// stored params first, so what the operator sees is what a save replaces.
export function ConfigModal({ test, onClose }: Readonly<ConfigModalProps>) {
  const configQuery = useQuery(getTestConfig, { id: test.id ?? "" });
  // The modal unmounts on close and remounts fresh on the next open, so a
  // post-save refetch is the reopen itself — no cache surgery needed.
  const updateMut = useMutation(updateTestConfig, { onSuccess: onClose });
  const [form] = Form.useForm<{ backgroundInvalidSeconds: number; params: ParamForm[] }>();

  const config = configQuery.data?.config;
  useEffect(() => {
    if (!config) return;
    form.setFieldsValue({
      backgroundInvalidSeconds: config.backgroundInvalidSeconds,
      params: Object.entries(config.engineParams ?? {}).map(([key, value]) => ({ key, value })),
    });
  }, [config, form]);

  return (
    <Modal
      cancelText={m.common_cancel()}
      confirmLoading={updateMut.isPending}
      okText={m.common_confirm()}
      onCancel={onClose}
      onOk={() =>
        void form.validateFields().then((v) => {
          const engineParams: Record<string, string> = {};
          for (const p of v.params ?? []) {
            if (p?.key) engineParams[p.key] = p.value ?? "";
          }
          updateMut.mutate({
            testId: test.id ?? "",
            backgroundInvalidSeconds: v.backgroundInvalidSeconds,
            engineParams,
          });
        })
      }
      open
      title={`${m.esports_tests_config_title()} · ${test.name ?? test.code ?? ""}`}
    >
      {updateMut.isError ? (
        <Alert className="mb-3" message={updateMut.error.message} type="error" />
      ) : null}
      {configQuery.isFetching || !config ? (
        <Skeleton active paragraph={{ rows: 5 }} />
      ) : (
        <Form form={form} layout="vertical">
          <Form.Item
            label={m.esports_tests_config_background()}
            name="backgroundInvalidSeconds"
            rules={[{ required: true }]}
          >
            <InputNumber className="w-full" max={3600} min={1} precision={0} />
          </Form.Item>
          <Form.Item label={m.esports_tests_config_params()}>
            <Form.List name="params">
              {(fields, { add, remove }) => (
                <div className="flex flex-col gap-2">
                  {fields.map((f) => (
                    <div className="flex gap-2" key={f.key}>
                      <Form.Item
                        className="m-0 flex-1"
                        name={[f.name, "key"]}
                        rules={[{ required: true }]}
                      >
                        <Input placeholder={m.esports_tests_config_param_key()} />
                      </Form.Item>
                      <Form.Item
                        className="m-0 flex-1"
                        name={[f.name, "value"]}
                        rules={[{ required: true }]}
                      >
                        <Input placeholder={m.esports_tests_config_param_value()} />
                      </Form.Item>
                      <Form.Item className="m-0">
                        <Button onClick={() => remove(f.name)} size="small" type="link">
                          {m.esports_tests_config_remove()}
                        </Button>
                      </Form.Item>
                    </div>
                  ))}
                  <Button onClick={() => add({ key: "", value: "" })} size="small" type="link">
                    {m.esports_tests_config_add()}
                  </Button>
                </div>
              )}
            </Form.List>
          </Form.Item>
          <Alert message={m.esports_tests_config_hint()} showIcon type="info" />
        </Form>
      )}
    </Modal>
  );
}
