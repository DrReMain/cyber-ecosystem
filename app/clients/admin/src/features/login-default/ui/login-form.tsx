import { type SubmitEvent, useState } from "react";
import { Button, Form, Input, Label, TextField } from "react-aria-components";
import { m } from "#/paraglide/messages";

const STYLES = {
  field: "mb-4.5",
  label: "mb-2 block font-medium text-[13px] text-black/55 dark:text-white/55",
  labelHint: "ms-2 font-mono text-[9.5px] text-black/35 tracking-[0.2em] dark:text-white/35",
  inputWrap: "relative",
  input:
    "h-11 w-full rounded-[10px] border border-black/10 bg-black/3 ps-3.5 pe-3.5 text-sm outline-none transition-[border-color,box-shadow,background-color] duration-200 placeholder:text-black/35 hover:border-black/22 focus:border-primary focus:bg-primary/4 focus:shadow-[0_0_0_3px] focus:shadow-primary/14 dark:border-white/10 dark:bg-white/3 dark:placeholder:text-white/35 dark:hover:border-white/22",
  reveal:
    "absolute end-2 top-1/2 -translate-y-1/2 cursor-pointer rounded-md border-0 bg-transparent px-2 py-1.5 font-mono text-[9.5px] text-black/35 tracking-[0.18em] transition-colors hover:text-primary dark:text-white/35",
  submit:
    "h-11.5 w-full cursor-pointer rounded-xl bg-linear-to-r from-primary to-primary-bright font-semibold text-[14.5px] text-white tracking-[0.06em] shadow-[0_4px_28px] shadow-primary/35 transition-all duration-250 hover:-translate-y-0.5 hover:shadow-[0_8px_36px] hover:shadow-primary/50 rtl:bg-linear-to-l rtl:tracking-normal",
};

interface LoginFormProps {
  onSubmit?: (event: SubmitEvent<HTMLFormElement>) => void;
  pending?: boolean;
}

export function LoginForm({ onSubmit, pending }: Readonly<LoginFormProps>) {
  const [passwordVisible, setPasswordVisible] = useState(false);

  return (
    <Form onSubmit={(event) => onSubmit?.(event as SubmitEvent<HTMLFormElement>)}>
      <TextField className={STYLES.field} name="email" type="text">
        <Label className={STYLES.label}>
          {m.login_default_label_email()}
          <span className={STYLES.labelHint}>EMAIL</span>
        </Label>
        <Input className={STYLES.input} placeholder="cyber@example.com" />
      </TextField>

      <TextField
        className={STYLES.field}
        name="password"
        type={passwordVisible ? "text" : "password"}
      >
        <Label className={STYLES.label}>
          {m.login_default_label_password()}
          <span className={STYLES.labelHint}>PASSWORD</span>
        </Label>
        <div className={STYLES.inputWrap}>
          <Input className={STYLES.input} placeholder="••••••••" />
          <Button
            className={STYLES.reveal}
            onPress={() => setPasswordVisible((visible) => !visible)}
            type="button"
          >
            {passwordVisible ? "HIDE" : "SHOW"}
          </Button>
        </div>
      </TextField>

      <Button className={STYLES.submit} isDisabled={pending} type="submit">
        {m.login_default_submit()}
      </Button>
    </Form>
  );
}
