-- Modify "cvs" table
ALTER TABLE "public"."cvs" ADD COLUMN "parsed_data" jsonb NULL, ADD COLUMN "status" character varying(50) NULL DEFAULT 'UPLOADED';
-- Create "roles" table
CREATE TABLE "public"."roles" ("id" uuid NOT NULL, "name" character varying(255) NOT NULL, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NOT NULL, "deleted_at" timestamptz NULL, "sector_id" uuid NOT NULL, PRIMARY KEY ("id"));
-- Create index "idx_roles_deleted_at" to table: "roles"
CREATE INDEX "idx_roles_deleted_at" ON "public"."roles" ("deleted_at");
-- Create index "idx_roles_sector_id" to table: "roles"
CREATE INDEX "idx_roles_sector_id" ON "public"."roles" ("sector_id");
-- Create "sectors" table
CREATE TABLE "public"."sectors" ("id" uuid NOT NULL, "name" character varying(255) NOT NULL, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NOT NULL, "deleted_at" timestamptz NULL, PRIMARY KEY ("id"));
-- Create index "idx_sectors_deleted_at" to table: "sectors"
CREATE INDEX "idx_sectors_deleted_at" ON "public"."sectors" ("deleted_at");
