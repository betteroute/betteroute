import { z } from "zod";

export const magicLinkSchema = z.object({
  email: z.email("Enter a valid email address").max(254),
  name: z.string().max(100, "Name is too long").optional(),
});

export const verifyMagicLinkSchema = z.object({
  token: z.string().min(1),
});

export type MagicLinkInput = z.infer<typeof magicLinkSchema>;
export type VerifyMagicLinkInput = z.infer<typeof verifyMagicLinkSchema>;
