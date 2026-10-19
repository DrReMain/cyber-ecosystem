import { Button } from "react-aria-components";
import { m } from "#/paraglide/messages";

export function SsoLogin() {
  const handleInitiate = () => {};

  return (
    <Button
      className="h-11 w-full cursor-pointer rounded-[10px] border border-black/10 bg-black/3 font-medium text-[13.5px] transition-all duration-200 hover:border-black/40 hover:bg-black/8 dark:border-white/10 dark:bg-white/4 dark:hover:border-white/40 dark:hover:bg-white/8"
      onPress={handleInitiate}
      type="button"
    >
      {m.login_default_sso()}
    </Button>
  );
}
