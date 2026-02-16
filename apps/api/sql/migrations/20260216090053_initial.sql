-- Create "workspaces" table
CREATE TABLE "workspaces" (
  "id" text NOT NULL,
  "name" text NOT NULL,
  "slug" text NOT NULL,
  "deleted_at" timestamptz NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "workspaces_name_length" CHECK ((char_length(name) >= 1) AND (char_length(name) <= 100)),
  CONSTRAINT "workspaces_slug_format" CHECK (slug ~ '^[a-z0-9]([a-z0-9-]*[a-z0-9])?$'::text),
  CONSTRAINT "workspaces_slug_length" CHECK ((char_length(slug) >= 1) AND (char_length(slug) <= 50))
);
-- Create index "idx_workspaces_slug_active" to table: "workspaces"
CREATE UNIQUE INDEX "idx_workspaces_slug_active" ON "workspaces" ("slug") WHERE (deleted_at IS NULL);
-- Create "links" table
CREATE TABLE "links" (
  "id" text NOT NULL,
  "workspace_id" text NOT NULL,
  "short_code" text NOT NULL,
  "dest_url" text NOT NULL,
  "title" text NULL,
  "description" text NULL,
  "is_active" boolean NOT NULL DEFAULT true,
  "expires_at" timestamptz NULL,
  "click_count" bigint NOT NULL DEFAULT 0,
  "last_clicked_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "links_workspace_id_fkey" FOREIGN KEY ("workspace_id") REFERENCES "workspaces" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "links_click_count_non_negative" CHECK (click_count >= 0),
  CONSTRAINT "links_description_length" CHECK ((description IS NULL) OR (char_length(description) <= 500)),
  CONSTRAINT "links_dest_url_length" CHECK (char_length(dest_url) <= 2048),
  CONSTRAINT "links_short_code_format" CHECK (short_code ~ '^[a-zA-Z0-9_-]+$'::text),
  CONSTRAINT "links_short_code_length" CHECK ((char_length(short_code) >= 1) AND (char_length(short_code) <= 50)),
  CONSTRAINT "links_title_length" CHECK ((title IS NULL) OR (char_length(title) <= 200))
);
-- Create index "idx_links_expiring" to table: "links"
CREATE INDEX "idx_links_expiring" ON "links" ("expires_at") WHERE ((expires_at IS NOT NULL) AND (is_active = true) AND (deleted_at IS NULL));
-- Create index "idx_links_redirect" to table: "links"
CREATE INDEX "idx_links_redirect" ON "links" ("short_code") INCLUDE ("dest_url", "is_active", "expires_at") WHERE (deleted_at IS NULL);
-- Create index "idx_links_short_code_active" to table: "links"
CREATE UNIQUE INDEX "idx_links_short_code_active" ON "links" ("short_code") WHERE (deleted_at IS NULL);
-- Create index "idx_links_workspace" to table: "links"
CREATE INDEX "idx_links_workspace" ON "links" ("workspace_id", "created_at" DESC) WHERE (deleted_at IS NULL);
