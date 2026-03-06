import { z } from "zod";

// Matches backend constraints from 001_workspaces.sql

const slugRegex = /^[a-z0-9]([a-z0-9-]*[a-z0-9])?$/;

export const createSchema = z.object({
  name: z.string().min(1, "Name is required").max(100, "Name is too long"),
  slug: z
    .string()
    .max(50, "Slug is too long")
    .regex(slugRegex, "Only lowercase letters, numbers, and hyphens")
    .or(z.literal("")),
});

export const updateSchema = z.object({
  name: z
    .string()
    .min(1, "Name is required")
    .max(100, "Name is too long")
    .optional(),
  slug: z
    .string()
    .min(1, "Slug is required")
    .max(50, "Slug is too long")
    .regex(slugRegex, "Only lowercase letters, numbers, and hyphens")
    .or(z.literal(""))
    .optional(),
});

export const inviteSchema = z.object({
  email: z.email("Invalid email").max(254, "Email is too long"),
  role: z.enum(["admin", "member", "viewer"], {
    message: "Role is required",
  }),
});

export const updateMemberRoleSchema = z.object({
  role: z.enum(["admin", "member", "viewer"], {
    message: "Role is required",
  }),
});

// Inferred types
export type CreateInput = z.infer<typeof createSchema>;
export type UpdateInput = z.infer<typeof updateSchema>;
export type InviteInput = z.infer<typeof inviteSchema>;
export type UpdateRoleInput = z.infer<typeof updateMemberRoleSchema>;
