-- Create "api_keys" table
CREATE TABLE "api_keys" (
  "id" text NOT NULL,
  "workspace_id" text NOT NULL,
  "created_by" text NULL,
  "name" text NOT NULL,
  "key_hash" text NOT NULL,
  "key_prefix" text NOT NULL,
  "expires_at" timestamptz NULL,
  "last_used_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "api_keys_created_by_fkey" FOREIGN KEY ("created_by") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "api_keys_workspace_id_fkey" FOREIGN KEY ("workspace_id") REFERENCES "workspaces" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "api_keys_name_length" CHECK ((char_length(name) >= 1) AND (char_length(name) <= 100))
);
-- Create index "idx_api_keys_hash" to table: "api_keys"
CREATE UNIQUE INDEX "idx_api_keys_hash" ON "api_keys" ("key_hash") INCLUDE ("workspace_id", "expires_at") WHERE (deleted_at IS NULL);
-- Create index "idx_api_keys_workspace" to table: "api_keys"
CREATE INDEX "idx_api_keys_workspace" ON "api_keys" ("workspace_id") WHERE (deleted_at IS NULL);
-- Create "workspace_invitations" table
CREATE TABLE "workspace_invitations" (
  "id" text NOT NULL,
  "workspace_id" text NOT NULL,
  "email" text NOT NULL,
  "role" text NOT NULL DEFAULT 'member',
  "token_hash" text NOT NULL,
  "invited_by" text NULL,
  "expires_at" timestamptz NOT NULL,
  "accepted_at" timestamptz NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "workspace_invitations_invited_by_fkey" FOREIGN KEY ("invited_by") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "workspace_invitations_workspace_id_fkey" FOREIGN KEY ("workspace_id") REFERENCES "workspaces" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "workspace_invitations_role_check" CHECK (role = ANY (ARRAY['admin'::text, 'member'::text, 'viewer'::text]))
);
-- Create index "idx_workspace_invitations_email" to table: "workspace_invitations"
CREATE INDEX "idx_workspace_invitations_email" ON "workspace_invitations" ("email") WHERE (accepted_at IS NULL);
-- Create index "idx_workspace_invitations_pending" to table: "workspace_invitations"
CREATE UNIQUE INDEX "idx_workspace_invitations_pending" ON "workspace_invitations" ("workspace_id", "email") WHERE (accepted_at IS NULL);
-- Create index "idx_workspace_invitations_token" to table: "workspace_invitations"
CREATE INDEX "idx_workspace_invitations_token" ON "workspace_invitations" ("token_hash") WHERE (accepted_at IS NULL);
-- Create "workspace_members" table
CREATE TABLE "workspace_members" (
  "workspace_id" text NOT NULL,
  "user_id" text NOT NULL,
  "role" text NOT NULL DEFAULT 'member',
  "invited_by" text NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("workspace_id", "user_id"),
  CONSTRAINT "workspace_members_invited_by_fkey" FOREIGN KEY ("invited_by") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "workspace_members_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "workspace_members_workspace_id_fkey" FOREIGN KEY ("workspace_id") REFERENCES "workspaces" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "workspace_members_role_check" CHECK (role = ANY (ARRAY['owner'::text, 'admin'::text, 'member'::text, 'viewer'::text]))
);
-- Create index "idx_workspace_members_user" to table: "workspace_members"
CREATE INDEX "idx_workspace_members_user" ON "workspace_members" ("user_id");
