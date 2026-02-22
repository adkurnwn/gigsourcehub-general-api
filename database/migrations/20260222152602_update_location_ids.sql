-- Modify "provinsi" table
ALTER TABLE "public"."provinsi" ALTER COLUMN "id" TYPE character varying(2);
-- Modify "users" table
ALTER TABLE "public"."users" ALTER COLUMN "kabupaten_id" TYPE character varying(5), ALTER COLUMN "provinsi_id" TYPE character varying(2);
-- Create "kabupaten_kota" table
CREATE TABLE "public"."kabupaten_kota" ("id" character varying(5) NOT NULL, "name" character varying(255) NOT NULL, "provinsi_id" character varying(2) NOT NULL, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NOT NULL, "deleted_at" timestamptz NULL, PRIMARY KEY ("id"));
-- Create index "idx_kabupaten_kota_deleted_at" to table: "kabupaten_kota"
CREATE INDEX "idx_kabupaten_kota_deleted_at" ON "public"."kabupaten_kota" ("deleted_at");
-- Create index "idx_kabupaten_kota_provinsi_id" to table: "kabupaten_kota"
CREATE INDEX "idx_kabupaten_kota_provinsi_id" ON "public"."kabupaten_kota" ("provinsi_id");
-- Drop "kabupaten" table
DROP TABLE "public"."kabupaten";
