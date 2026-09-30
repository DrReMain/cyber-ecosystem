import { useNavigate, useSearch } from "@tanstack/react-router";
import type { SubmitEvent } from "react";
import { useEffect } from "react";
import { toast } from "sonner";
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

export function LoginPage() {
  const { expired } = useSearch({ from: LOGIN_ROUTE_ID });
  const navigate = useNavigate();
  const login = usePasswordLogin();

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
    await login
      .mutateAsync({
        email: String(form.get("email")),
        password: String(form.get("password")),
      })
      .catch(() => {});
  };

  return (
    <LoginShell>
      <LoginForm onSubmit={handleSubmit} pending={login.isPending} />
      <ForgotPassword />
      <AuthDivider />
      <SsoLogin />
    </LoginShell>
  );
}
