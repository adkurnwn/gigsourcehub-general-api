-- Modify "users" table
ALTER TABLE "public"."users" DROP COLUMN "kabupaten", DROP COLUMN "provinsi", ADD COLUMN "kabupaten_id" uuid NULL, ADD COLUMN "provinsi_id" uuid NULL;
-- Create index "idx_users_kabupaten_id" to table: "users"
CREATE INDEX "idx_users_kabupaten_id" ON "public"."users" ("kabupaten_id");
-- Create index "idx_users_provinsi_id" to table: "users"
CREATE INDEX "idx_users_provinsi_id" ON "public"."users" ("provinsi_id");
-- Create "kabupaten" table
CREATE TABLE "public"."kabupaten" ("id" uuid NOT NULL, "name" character varying(255) NOT NULL, "provinsi_id" uuid NOT NULL, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NOT NULL, "deleted_at" timestamptz NULL, PRIMARY KEY ("id"));
-- Create index "idx_kabupaten_deleted_at" to table: "kabupaten"
CREATE INDEX "idx_kabupaten_deleted_at" ON "public"."kabupaten" ("deleted_at");
-- Create index "idx_kabupaten_provinsi_id" to table: "kabupaten"
CREATE INDEX "idx_kabupaten_provinsi_id" ON "public"."kabupaten" ("provinsi_id");
-- Create "provinsi" table
CREATE TABLE "public"."provinsi" ("id" uuid NOT NULL, "name" character varying(255) NOT NULL, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NOT NULL, "deleted_at" timestamptz NULL, PRIMARY KEY ("id"));
-- Create index "idx_provinsi_deleted_at" to table: "provinsi"
CREATE INDEX "idx_provinsi_deleted_at" ON "public"."provinsi" ("deleted_at");
