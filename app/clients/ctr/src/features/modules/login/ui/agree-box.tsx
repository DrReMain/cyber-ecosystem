import { Check, Minus } from "lucide-react";

interface AgreeBoxProps {
  checked: boolean;
  indeterminate?: boolean;
}

/**
 * The visual checkbox square of the login agreement rows — aria-hidden
 * decoration; the RAC CheckboxButton wrapping it owns the role and state.
 */
export function AgreeBox({ checked, indeterminate = false }: Readonly<AgreeBoxProps>) {
  return (
    <span
      aria-hidden
      className={`mt-0.5 flex h-4 w-4 shrink-0 items-center justify-center rounded-sm border text-white ${
        checked || indeterminate ? "border-ambient-blue bg-ambient-blue" : "border-line"
      }`}
    >
      {checked ? (
        <Check size={11} strokeWidth={3} />
      ) : indeterminate ? (
        <Minus size={11} strokeWidth={3} />
      ) : null}
    </span>
  );
}
