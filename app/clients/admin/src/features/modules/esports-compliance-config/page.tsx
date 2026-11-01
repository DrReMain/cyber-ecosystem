import { useMutation, useQuery } from "@connectrpc/connect-query";
import {
  getComplianceConfig,
  setBirthCollection,
  setCompliancePrompts,
  setPayAgeLine,
} from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/compliance_admin-ComplianceAdminService_connectquery";
import { Alert, Button, Card, Switch, Tag, Typography } from "antd";
import { Pencil, RefreshCw } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";
import { m } from "#/paraglide/messages";
import { AgeLineModal } from "./ui/age-line-modal";
import { PromptModal } from "./ui/prompt-modal";

// Which of the two prompt rows an edit opened — the modal writes exactly
// that one, nil on the wire for the other.
type PromptKey = "minor" | "missing";

// The prototype's three panels: the birth-collection switch (an instant
// flip), the pay age line behind a sensitive double confirm, and the two
// pay-wall prompt overrides. Every knob answers from the one singleton
// row; the prompts fall back to this locale's shipped copy until an
// operator overrides them.
export function ComplianceConfigPage() {
  const [ageOpen, setAgeOpen] = useState(false);
  const [promptKey, setPromptKey] = useState<PromptKey | null>(null);
  const configQuery = useQuery(getComplianceConfig, {});

  const birthMut = useMutation(setBirthCollection, {
    onSuccess: (resp) => {
      toast.success(
        resp.config?.birthCollect
          ? m.esports_compliance_config_birth_on_ok()
          : m.esports_compliance_config_birth_off_ok(),
      );
      void configQuery.refetch();
    },
  });
  const ageMut = useMutation(setPayAgeLine, {
    onSuccess: (resp) => {
      toast.success(m.esports_compliance_config_age_ok({ age: resp.config?.payAgeLine ?? 0 }));
      setAgeOpen(false);
      void configQuery.refetch();
    },
  });
  const promptMut = useMutation(setCompliancePrompts, {
    onSuccess: () => {
      toast.success(m.esports_compliance_config_prompts_ok());
      setPromptKey(null);
      void configQuery.refetch();
    },
  });

  const config = configQuery.data?.config;
  const birthOn = config?.birthCollect ?? true;
  const ageLine = config?.payAgeLine ?? 18;
  const prompts: Record<PromptKey, string | null> = {
    minor: config?.minorPrompt ?? null,
    missing: config?.missingBirthPrompt ?? null,
  };
  const promptDefaults: Record<PromptKey, () => string> = {
    minor: m.esports_compliance_config_prompts_minor_default,
    missing: m.esports_compliance_config_prompts_missing_default,
  };

  const promptRow = (key: PromptKey) => {
    const overridden = prompts[key] !== null;
    return (
      <div className="border-fill-secondary flex items-center gap-3 border-b py-3 last:border-b-0">
        <span className="w-40 shrink-0 text-sm">
          {key === "minor"
            ? m.esports_compliance_config_prompts_minor_label()
            : m.esports_compliance_config_prompts_missing_label()}
        </span>
        <span className={`flex-1 text-sm ${overridden ? "" : "text-text-secondary"}`}>
          {prompts[key] ?? promptDefaults[key]()}
        </span>
        {!overridden && <Tag>{m.esports_compliance_config_prompts_default_tag()}</Tag>}
        <Button
          icon={<Pencil size={14} />}
          onClick={() => setPromptKey(key)}
          size="small"
          type="text"
        >
          {m.esports_compliance_config_prompts_edit()}
        </Button>
      </div>
    );
  };

  return (
    <div className="flex flex-col gap-4 p-4">
      <div className="flex items-start justify-between gap-2">
        <div>
          <Typography.Title className="!mb-0" level={4}>
            {m.esports_compliance_config_title()}
          </Typography.Title>
          <Typography.Text className="text-xs" type="secondary">
            {m.esports_compliance_config_desc()}
          </Typography.Text>
        </div>
        <Button icon={<RefreshCw size={14} />} onClick={() => void configQuery.refetch()}>
          {m.esports_compliance_config_refresh()}
        </Button>
      </div>

      <Card loading={configQuery.isLoading} title={m.esports_compliance_config_birth_title()}>
        <div className="flex items-center gap-3 py-1">
          <span className="text-sm">{m.esports_compliance_config_birth_label()}</span>
          <Switch
            checked={birthOn}
            loading={birthMut.isPending}
            onChange={(checked) => birthMut.mutate({ collect: checked })}
          />
          <Tag color={birthOn ? "success" : "default"}>
            {birthOn
              ? m.esports_compliance_config_birth_on()
              : m.esports_compliance_config_birth_off()}
          </Tag>
          <span className="text-text-secondary ml-auto text-xs">
            {m.esports_compliance_config_birth_hint()}
          </span>
        </div>
      </Card>

      <Card
        loading={configQuery.isLoading}
        title={
          <span className="flex items-center gap-2">
            {m.esports_compliance_config_age_title()}
            <Tag color="red">{m.esports_compliance_config_age_sensitive()}</Tag>
          </span>
        }
      >
        <div className="flex items-center gap-3 py-1">
          <span className="text-sm">{m.esports_compliance_config_age_label()}</span>
          <span className="text-warning-text text-3xl font-black">{ageLine}</span>
          <span className="text-text-secondary text-sm">
            {m.esports_compliance_config_age_unit()}
          </span>
          <Button className="ml-auto" onClick={() => setAgeOpen(true)} type="link">
            {m.esports_compliance_config_age_edit()}
          </Button>
        </div>
      </Card>

      <Card loading={configQuery.isLoading} title={m.esports_compliance_config_prompts_title()}>
        {promptRow("minor")}
        {promptRow("missing")}
      </Card>

      <Alert showIcon title={m.esports_compliance_config_callout()} type="warning" />

      <AgeLineModal
        initial={ageLine}
        onClose={() => setAgeOpen(false)}
        onConfirm={(age) => ageMut.mutate({ age })}
        open={ageOpen}
        submitting={ageMut.isPending}
      />
      <PromptModal
        initial={promptKey !== null ? prompts[promptKey] : null}
        onClose={() => setPromptKey(null)}
        onSubmit={(text) =>
          promptMut.mutate(
            promptKey === "minor" ? { minorPrompt: text } : { missingBirthPrompt: text },
          )
        }
        open={promptKey !== null}
        submitting={promptMut.isPending}
      />
    </div>
  );
}
