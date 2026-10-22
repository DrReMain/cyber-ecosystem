import { Button } from "react-aria-components";
import { m } from "#/paraglide/messages";

export function ForgotPassword() {
  const handleLaunch = () => {};

  return (
    <div className="mt-4 flex justify-end">
      <Button
        className="hover:text-primary cursor-pointer border-0 bg-transparent p-0 text-[13px] text-black/55 transition-colors dark:text-white/55"
        onPress={handleLaunch}
        type="button"
      >
        {m.login_default_forgot()}
      </Button>
    </div>
  );
}
