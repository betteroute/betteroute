import { z } from "zod";

export const createSchema = z.object({
  dest_url: z.url("Must be a valid URL"),
  short_code: z
    .string()
    .max(50, "Short code is too long")
    .regex(
      /^[a-zA-Z0-9_-]*$/,
      "Only letters, numbers, hyphens, and underscores",
    ),
  title: z.string().max(200, "Title is too long"),
  description: z.string().max(500, "Description is too long"),
});

export type CreateInput = z.infer<typeof createSchema>;
