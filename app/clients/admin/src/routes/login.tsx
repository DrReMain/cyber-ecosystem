import { createFileRoute } from "@tanstack/react-router";
import { useState } from "react";

export const Route = createFileRoute("/login")({
  component: LoginPage,
});

function LoginPage() {
  const [email, setEmail] = useState("admin@cyber.local");
  const [password, setPassword] = useState("admin");

  return (
    <div className="flex min-h-screen items-center justify-center">
      <div className="flex w-64 flex-col gap-2">
        <input
          className="border px-2 py-1"
          onChange={(e) => setEmail(e.target.value)}
          placeholder="email"
          type="email"
          value={email}
        />
        <input
          className="border px-2 py-1"
          onChange={(e) => setPassword(e.target.value)}
          placeholder="password"
          type="password"
          value={password}
        />
        <button className="border px-2 py-1" type="submit">
          login
        </button>
      </div>
    </div>
  );
}
