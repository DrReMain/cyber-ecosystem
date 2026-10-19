import type { FeedbackSink } from "@cyber-ecosystem/shared-error";
import { toast } from "sonner";

export const sonnerFeedback: FeedbackSink = (view) => {
  toast.error(view.title, { description: view.message !== "" ? view.message : undefined });
};
