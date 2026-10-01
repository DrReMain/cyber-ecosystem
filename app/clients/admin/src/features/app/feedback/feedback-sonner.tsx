import { useTheme } from "@cyber-ecosystem/shared-theme";
import { Toaster } from "sonner";
import { getTextDirection } from "#/paraglide/runtime";

export function FeedbackSonner() {
  const { preference } = useTheme();

  return (
    <Toaster
      closeButton
      expand={false}
      position={getTextDirection() === "rtl" ? "top-left" : "top-right"}
      richColors
      theme={preference === "dark" ? "dark" : "light"}
    />
  );
}
