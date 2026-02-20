-- Create "users" table
CREATE TABLE "users" (
  "id" text NOT NULL,
  "name" text NOT NULL,
  "email" text NOT NULL,
  "email_verified_at" timestamptz NULL,
  "avatar_url" text NULL,
  "status" text NOT NULL DEFAULT 'active',
  "onboarded_at" timestamptz NULL,
  "last_login_at" timestamptz NULL,
  "timezone" text NOT NULL DEFAULT 'UTC',
  "deleted_at" timestamptz NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "users_email_length" CHECK ((char_length(email) >= 3) AND (char_length(email) <= 254)),
  CONSTRAINT "users_name_length" CHECK ((char_length(name) >= 1) AND (char_length(name) <= 100)),
  CONSTRAINT "users_status_check" CHECK (status = ANY (ARRAY['active'::text, 'suspended'::text, 'banned'::text]))
);
-- Create index "idx_users_email" to table: "users"
CREATE UNIQUE INDEX "idx_users_email" ON "users" ((lower(email))) WHERE (deleted_at IS NULL);
-- Create "accounts" table
CREATE TABLE "accounts" (
  "id" text NOT NULL,
  "user_id" text NOT NULL,
  "provider" text NOT NULL,
  "provider_account_id" text NOT NULL,
  "password_hash" text NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "accounts_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "idx_accounts_provider" to table: "accounts"
CREATE UNIQUE INDEX "idx_accounts_provider" ON "accounts" ("provider", "provider_account_id");
-- Create index "idx_accounts_user" to table: "accounts"
CREATE INDEX "idx_accounts_user" ON "accounts" ("user_id");
-- Create "sessions" table
CREATE TABLE "sessions" (
  "id" text NOT NULL,
  "user_id" text NOT NULL,
  "token_hash" text NOT NULL,
  "expires_at" timestamptz NOT NULL,
  "ip_address" text NULL,
  "user_agent" text NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "sessions_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "idx_sessions_expires" to table: "sessions"
CREATE INDEX "idx_sessions_expires" ON "sessions" ("expires_at");
-- Create index "idx_sessions_token_hash" to table: "sessions"
CREATE UNIQUE INDEX "idx_sessions_token_hash" ON "sessions" ("token_hash") INCLUDE ("user_id", "expires_at");
-- Create index "idx_sessions_user" to table: "sessions"
CREATE INDEX "idx_sessions_user" ON "sessions" ("user_id");
-- Create "verification_tokens" table
CREATE TABLE "verification_tokens" (
  "id" text NOT NULL,
  "user_id" text NOT NULL,
  "email" text NOT NULL,
  "token_hash" text NOT NULL,
  "type" text NOT NULL,
  "expires_at" timestamptz NOT NULL,
  "used_at" timestamptz NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "verification_tokens_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "verification_tokens_type_check" CHECK (type = ANY (ARRAY['email_verification'::text, 'password_reset'::text, 'magic_link'::text]))
);
-- Create index "idx_verification_tokens_expires" to table: "verification_tokens"
CREATE INDEX "idx_verification_tokens_expires" ON "verification_tokens" ("expires_at") WHERE (used_at IS NULL);
-- Create index "idx_verification_tokens_hash" to table: "verification_tokens"
CREATE INDEX "idx_verification_tokens_hash" ON "verification_tokens" ("token_hash") WHERE (used_at IS NULL);
-- Create index "idx_verification_tokens_rate" to table: "verification_tokens"
CREATE INDEX "idx_verification_tokens_rate" ON "verification_tokens" ("email", "type", "created_at" DESC) WHERE (used_at IS NULL);
