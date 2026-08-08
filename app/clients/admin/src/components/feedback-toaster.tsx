import { useTheme } from "@cyber-ecosystem/shared-theme";
import { useEffect } from "react";
import { Toaster, toast } from "sonner";
import { errorHandler } from "#/domains/errors/error-handler";
import { errorMessage } from "#/domains/i18n/error-message";

export function FeedbackToaster() {
  const { preference } = useTheme();

  useEffect(() => {
    errorHandler.register({
      feedback: (error) => {
        toast.error(errorMessage(error));
      },
    });
  }, []);

  return (
    <Toaster
      closeButton
      expand={false}
      position="bottom-right"
      richColors
      theme={preference === "dark" ? "dark" : "light"}
    />
  );
}
