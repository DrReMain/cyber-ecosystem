import { Toaster } from "sonner";
import { getTextDirection } from "#/paraglide/runtime";

// Dark-only: the toaster theme is fixed, no theme context to read.
export function FeedbackSonner() {
  return (
    <Toaster
      closeButton
      expand={false}
      position={getTextDirection() === "rtl" ? "top-left" : "top-right"}
      richColors
      theme="dark"
    />
  );
}
