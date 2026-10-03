import { useMutation, useQuery } from "@connectrpc/connect-query";
import {
  deleteMyAgentConfig,
  getMyAgentConfig,
  updateMyAgentConfig,
} from "@cyber-ecosystem/gen-connect-ts/cyber/agent/v1/agent_config-AgentConfigService_connectquery";
import { Button, Card, Form, Input, Popconfirm } from "antd";
import { useEffect, useRef } from "react";
import { toast } from "sonner";
import { m } from "#/paraglide/messages";
import { getTextDirection } from "#/paraglide/runtime";

interface AgentConfigForm {
  baseUrl: string;
  apiKey?: string;
}

export function AgentsPage() {
  const [form] = Form.useForm<AgentConfigForm>();
  const cfgQuery = useQuery(getMyAgentConfig, {});
  const config = cfgQuery.data?.agentConfig ?? null;
  const hydrated = useRef(false);
  useEffect(() => {
    if (!hydrated.current && cfgQuery.data) {
      hydrated.current = true;
      form.setFieldsValue({ baseUrl: config?.baseUrl ?? "" });
    }
  }, [cfgQuery.data, config?.baseUrl, form]);

  const updateMut = useMutation(updateMyAgentConfig, {
    onSuccess: () => {
      toast.success(m.profile_agents_saved());
      form.setFieldValue("apiKey", "");
      void cfgQuery.refetch();
    },
  });
  const deleteMut = useMutation(deleteMyAgentConfig, {
    onSuccess: () => {
      toast.success(m.profile_agents_removed());
      form.resetFields();
      void cfgQuery.refetch();
    },
  });

  const submit = (values: AgentConfigForm) => {
    const apiKey = values.apiKey?.trim();
    updateMut.mutate({
      baseUrl: values.baseUrl.trim(),
      ...(apiKey ? { apiKey } : {}),
    });
  };

  return (
    <Card
      extra={
        config ? (
          <Popconfirm
            cancelButtonProps={{ variant: "filled", color: "default" }}
            okButtonProps={{ variant: "filled", color: "danger" }}
            onConfirm={() => deleteMut.mutate({})}
            placement={getTextDirection() === "rtl" ? "bottomLeft" : "bottomRight"}
            title={m.profile_agents_act_remove_confirm()}
          >
            <Button color="danger" loading={deleteMut.isPending} size="small" variant="text">
              {m.profile_agents_act_remove()}
            </Button>
          </Popconfirm>
        ) : undefined
      }
      loading={cfgQuery.isPending}
      title={m.profile_agents_title()}
    >
      <Form className="max-w-md" form={form} layout="vertical" onFinish={submit}>
        <Form.Item
          label={m.profile_agents_form_base_url()}
          name="baseUrl"
          rules={[
            { required: true, message: m.profile_agents_form_base_url_required() },
            { pattern: /^https?:\/\//, message: m.profile_agents_form_base_url_invalid() },
          ]}
        >
          <Input maxLength={512} />
        </Form.Item>
        <Form.Item label={m.profile_agents_form_api_key()} name="apiKey">
          <Input.Password
            maxLength={512}
            placeholder={
              config?.apiKeySet === true
                ? m.profile_agents_form_api_key_set()
                : m.profile_agents_form_api_key_unset()
            }
          />
        </Form.Item>
        <Button color="primary" htmlType="submit" loading={updateMut.isPending} variant="filled">
          {m.profile_agents_form_submit()}
        </Button>
      </Form>
    </Card>
  );
}
