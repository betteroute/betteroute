-- Create "folders" table
CREATE TABLE "folders" (
  "id" text NOT NULL,
  "workspace_id" text NOT NULL,
  "name" text NOT NULL,
  "color" text NOT NULL DEFAULT '#6366f1',
  "position" integer NOT NULL DEFAULT 0,
  "deleted_at" timestamptz NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "folders_workspace_id_fkey" FOREIGN KEY ("workspace_id") REFERENCES "workspaces" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "folders_color_hex" CHECK (color ~ '^#[0-9a-fA-F]{6}$'::text),
  CONSTRAINT "folders_name_length" CHECK ((char_length(name) >= 1) AND (char_length(name) <= 100)),
  CONSTRAINT "folders_position_non_negative" CHECK ("position" >= 0)
);
-- Create index "idx_folders_workspace" to table: "folders"
CREATE INDEX "idx_folders_workspace" ON "folders" ("workspace_id", "position") WHERE (deleted_at IS NULL);
-- Create index "idx_folders_workspace_name" to table: "folders"
CREATE UNIQUE INDEX "idx_folders_workspace_name" ON "folders" ("workspace_id", "name") WHERE (deleted_at IS NULL);
-- Modify "links" table
ALTER TABLE "links" ADD COLUMN "folder_id" text NULL, ADD CONSTRAINT "fk_links_folder" FOREIGN KEY ("folder_id") REFERENCES "folders" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Create index "idx_links_folder" to table: "links"
CREATE INDEX "idx_links_folder" ON "links" ("folder_id") WHERE ((folder_id IS NOT NULL) AND (deleted_at IS NULL));
-- Create "tags" table
CREATE TABLE "tags" (
  "id" text NOT NULL,
  "workspace_id" text NOT NULL,
  "name" text NOT NULL,
  "color" text NOT NULL DEFAULT '#6366f1',
  "deleted_at" timestamptz NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "tags_workspace_id_fkey" FOREIGN KEY ("workspace_id") REFERENCES "workspaces" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "tags_color_hex" CHECK (color ~ '^#[0-9a-fA-F]{6}$'::text),
  CONSTRAINT "tags_name_length" CHECK ((char_length(name) >= 1) AND (char_length(name) <= 50))
);
-- Create index "idx_tags_workspace" to table: "tags"
CREATE INDEX "idx_tags_workspace" ON "tags" ("workspace_id") WHERE (deleted_at IS NULL);
-- Create index "idx_tags_workspace_name" to table: "tags"
CREATE UNIQUE INDEX "idx_tags_workspace_name" ON "tags" ("workspace_id", (lower(name))) WHERE (deleted_at IS NULL);
-- Create "link_tags" table
CREATE TABLE "link_tags" (
  "link_id" text NOT NULL,
  "tag_id" text NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("link_id", "tag_id"),
  CONSTRAINT "link_tags_link_id_fkey" FOREIGN KEY ("link_id") REFERENCES "links" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "link_tags_tag_id_fkey" FOREIGN KEY ("tag_id") REFERENCES "tags" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "idx_link_tags_tag" to table: "link_tags"
CREATE INDEX "idx_link_tags_tag" ON "link_tags" ("tag_id");
