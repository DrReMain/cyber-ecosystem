import { BottomConfirmSheet } from "#/features/app/dialog/bottom-confirm-sheet";
import { m } from "#/paraglide/messages";

interface LogoutSheetProps {
  open: boolean;
  onCancel: () => void;
  onConfirm: () => void;
}

/**
 * The prototype's bottom-sheet confirm for sign-out: the primary button is
 * the plain brand CTA (sign-out is an ordinary move here, not a danger
 * one), cancel leads the stack as the focused answer, and confirm never
 * auto-closes — the caller owns closing after logout settles.
 */
export function LogoutSheet({ open, onCancel, onConfirm }: Readonly<LogoutSheetProps>) {
  if (!open) {
    return null;
  }
  return (
    <BottomConfirmSheet
      actions={[
        { label: m.profile_logout_cancel(), kind: "ghost", onPress: onCancel, autoFocus: true },
        { label: m.profile_logout_confirm(), kind: "primary", onPress: onConfirm },
      ]}
      body={m.profile_logout_body()}
      onDismiss={onCancel}
      title={m.profile_logout_title()}
    />
  );
}
