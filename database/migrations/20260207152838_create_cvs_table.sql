-- Create "cvs" table
CREATE TABLE "public"."cvs" ("id" uuid NOT NULL, "name" character varying(255) NOT NULL, "path" character varying(255) NOT NULL, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NOT NULL, "deleted_at" timestamptz NULL, "user_id" uuid NOT NULL, PRIMARY KEY ("id"));
-- Create index "idx_cvs_deleted_at" to table: "cvs"
CREATE INDEX "idx_cvs_deleted_at" ON "public"."cvs" ("deleted_at");
-- Create index "idx_cvs_user_id" to table: "cvs"
CREATE INDEX "idx_cvs_user_id" ON "public"."cvs" ("user_id");
