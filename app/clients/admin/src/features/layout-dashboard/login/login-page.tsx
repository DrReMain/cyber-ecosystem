import { System } from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/error_reason_pb";
import { useNavigate, useSearch } from "@tanstack/react-router";
import type { SubmitEvent } from "react";
import { useEffect, useState } from "react";
import { toast } from "sonner";
import { systemReason } from "#/domains/error";
import {
  AuthDivider,
  ForgotPassword,
  LoginForm,
  LoginShell,
  SsoLogin,
} from "#/features/login-default";
import { m } from "#/paraglide/messages";
import { LOGIN_PATH, LOGIN_ROUTE_ID } from "../area";
import { usePasswordLogin } from "./use-password-login";

// The login verdict branches per reason — wrong password, no seat at this
// client's door, a disabled account, a disabled tenant are different stories
// with different next actions; none of them collapses into "bad credentials".
function failureCopy(reason: string | undefined): string | null {
  switch (systemReason(reason)) {
    case System.LOGIN_FAILED:
      return m.error_SYSTEM_LOGIN_FAILED();
    case System.LOGIN_SEAT_REQUIRED:
      return m.error_SYSTEM_LOGIN_SEAT_REQUIRED();
    case System.USER_DISABLED:
      return m.error_SYSTEM_USER_DISABLED();
    case System.TENANT_DISABLED:
      return m.error_SYSTEM_TENANT_DISABLED();
    case undefined:
      return null;
    default:
      return m.error_GENERAL_ERROR_UNSPECIFIED();
  }
}

export function LoginPage() {
  const { expired } = useSearch({ from: LOGIN_ROUTE_ID });
  const navigate = useNavigate();
  const login = usePasswordLogin();
  const [serverReason, setServerReason] = useState<string | undefined>(undefined);

  useEffect(() => {
    if (!expired) return;
    toast(m.login_default_session_expired(), { id: "session-expired" });
    void navigate({
      to: LOGIN_PATH,
      search: (prev) => ({ ...prev, expired: undefined }),
      replace: true,
      viewTransition: false,
    });
  }, [expired, navigate]);

  const handleSubmit = async (event: SubmitEvent<HTMLFormElement>) => {
    event.preventDefault();
    const form = new FormData(event.currentTarget);
    setServerReason(undefined);
    const result = await login
      .mutateAsync({
        account: String(form.get("account")),
        password: String(form.get("password")),
      })
      .catch(() => undefined);
    if (result !== undefined && result.status === "rejected") {
      setServerReason(result.reason);
    }
  };

  const copy = failureCopy(serverReason);

  return (
    <LoginShell>
      <LoginForm
        feedback={
          copy === null
            ? null
            : systemReason(serverReason) === System.LOGIN_SEAT_REQUIRED
              ? {
                  copy,
                  action: {
                    label: m.login_default_error_switch_account(),
                    onAction: () => setServerReason(undefined),
                  },
                }
              : { copy }
        }
        onSubmit={handleSubmit}
        pending={login.isPending}
      />
      <ForgotPassword />
      <AuthDivider />
      <SsoLogin />
    </LoginShell>
  );
}
