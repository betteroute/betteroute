-- Add created_by to links, folders, tags for user attribution.
-- Add created_by to api_keys covering index for index-only auth scans.

-- Links: who created this link (web, API, import).
ALTER TABLE "links" ADD COLUMN "created_by" text NULL;
ALTER TABLE "links" ADD CONSTRAINT "links_created_by_fkey"
    FOREIGN KEY ("created_by") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;

-- Folders: who created this folder.
ALTER TABLE "folders" ADD COLUMN "created_by" text NULL;
ALTER TABLE "folders" ADD CONSTRAINT "folders_created_by_fkey"
    FOREIGN KEY ("created_by") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;

-- Tags: who created this tag.
ALTER TABLE "tags" ADD COLUMN "created_by" text NULL;
ALTER TABLE "tags" ADD CONSTRAINT "tags_created_by_fkey"
    FOREIGN KEY ("created_by") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;

-- Rebuild api_keys covering index to include created_by (index-only auth scan).
DROP INDEX IF EXISTS "idx_api_keys_hash";
CREATE UNIQUE INDEX "idx_api_keys_hash" ON "api_keys" ("key_hash")
    INCLUDE ("workspace_id", "created_by", "expires_at", "permission", "scopes")
    WHERE (deleted_at IS NULL);
