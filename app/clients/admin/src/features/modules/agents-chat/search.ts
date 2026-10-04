import { z } from "zod";

export const agentsChatSearchSchema = z.looseObject({
  session: z.string().min(1).optional(),
});
