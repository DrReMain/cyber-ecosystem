import { z } from "zod";

export const diagSearchSchema = z.looseObject({
  userId: z.string().optional(),
});
