-- Modify "users" table
ALTER TABLE "public"."users" ADD COLUMN "verified_at" timestamptz NULL;
-- Create "user_tokens" table
CREATE TABLE "public"."user_tokens" ("id" uuid NOT NULL, "user_id" uuid NOT NULL, "type" character varying(50) NOT NULL, "token" character varying(255) NOT NULL, "expires_at" timestamptz NOT NULL, "created_at" timestamptz NOT NULL DEFAULT now(), PRIMARY KEY ("id"), CONSTRAINT "user_tokens_user_id_fk" FOREIGN KEY ("user_id") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- Create index "idx_user_tokens_token" to table: "user_tokens"
CREATE UNIQUE INDEX "idx_user_tokens_token" ON "public"."user_tokens" ("token");
-- Create index "idx_user_tokens_user_id_type" to table: "user_tokens"
CREATE INDEX "idx_user_tokens_user_id_type" ON "public"."user_tokens" ("user_id", "type");
