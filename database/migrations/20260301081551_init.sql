-- Create "cvs" table
CREATE TABLE "public"."cvs" ("id" uuid NOT NULL, "filename" character varying(255) NOT NULL, "path" character varying(255) NOT NULL, "parsed_data" jsonb NULL, "status" character varying(50) NULL DEFAULT 'UPLOADED', "created_at" timestamptz NOT NULL, "updated_at" timestamptz NOT NULL, "deleted_at" timestamptz NULL, "user_id" uuid NOT NULL, PRIMARY KEY ("id"));
-- Create index "idx_cvs_deleted_at" to table: "cvs"
CREATE INDEX "idx_cvs_deleted_at" ON "public"."cvs" ("deleted_at");
-- Create index "idx_cvs_user_id" to table: "cvs"
CREATE UNIQUE INDEX "idx_cvs_user_id" ON "public"."cvs" ("user_id");
-- Create "kabupaten_kota" table
CREATE TABLE "public"."kabupaten_kota" ("id" character varying(5) NOT NULL, "name" character varying(255) NOT NULL, "provinsi_id" character varying(2) NOT NULL, "created_at" timestamptz NOT NULL DEFAULT now(), "updated_at" timestamptz NOT NULL DEFAULT now(), "deleted_at" timestamptz NULL, PRIMARY KEY ("id"));
-- Create index "idx_kabupaten_kota_deleted_at" to table: "kabupaten_kota"
CREATE INDEX "idx_kabupaten_kota_deleted_at" ON "public"."kabupaten_kota" ("deleted_at");
-- Create index "idx_kabupaten_kota_provinsi_id" to table: "kabupaten_kota"
CREATE INDEX "idx_kabupaten_kota_provinsi_id" ON "public"."kabupaten_kota" ("provinsi_id");
-- Create "provinsi" table
CREATE TABLE "public"."provinsi" ("id" character varying(2) NOT NULL, "name" character varying(255) NOT NULL, "created_at" timestamptz NOT NULL DEFAULT now(), "updated_at" timestamptz NOT NULL DEFAULT now(), "deleted_at" timestamptz NULL, PRIMARY KEY ("id"));
-- Create index "idx_provinsi_deleted_at" to table: "provinsi"
CREATE INDEX "idx_provinsi_deleted_at" ON "public"."provinsi" ("deleted_at");
-- Create "recruitment_status" table
CREATE TABLE "public"."recruitment_status" ("id" uuid NOT NULL, "name" character varying(150) NOT NULL, "hex_code" character varying(10) NULL, "is_active" boolean NOT NULL DEFAULT true, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NOT NULL, "deleted_at" timestamptz NULL, PRIMARY KEY ("id"));
-- Create index "idx_recruitment_status_deleted_at" to table: "recruitment_status"
CREATE INDEX "idx_recruitment_status_deleted_at" ON "public"."recruitment_status" ("deleted_at");
-- Create "role_applied" table
CREATE TABLE "public"."role_applied" ("id" uuid NOT NULL, "name" character varying(150) NOT NULL, "sector_id" uuid NOT NULL, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NOT NULL, "deleted_at" timestamptz NULL, PRIMARY KEY ("id"));
-- Create index "idx_role_applied_deleted_at" to table: "role_applied"
CREATE INDEX "idx_role_applied_deleted_at" ON "public"."role_applied" ("deleted_at");
-- Create index "idx_role_applied_sector_id" to table: "role_applied"
CREATE INDEX "idx_role_applied_sector_id" ON "public"."role_applied" ("sector_id");
-- Create "role_system" table
CREATE TABLE "public"."role_system" ("id" uuid NOT NULL, "name" character varying(150) NOT NULL, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NOT NULL, "deleted_at" timestamptz NULL, PRIMARY KEY ("id"));
-- Create index "idx_role_system_deleted_at" to table: "role_system"
CREATE INDEX "idx_role_system_deleted_at" ON "public"."role_system" ("deleted_at");
-- Create "sectors" table
CREATE TABLE "public"."sectors" ("id" uuid NOT NULL, "name" character varying(150) NOT NULL, "is_active" boolean NOT NULL DEFAULT true, "hex_code" character varying(10) NULL, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NOT NULL, "deleted_at" timestamptz NULL, PRIMARY KEY ("id"));
-- Create index "idx_sectors_deleted_at" to table: "sectors"
CREATE INDEX "idx_sectors_deleted_at" ON "public"."sectors" ("deleted_at");
-- Create "users" table
CREATE TABLE "public"."users" ("id" uuid NOT NULL, "name" character varying(255) NOT NULL, "email" character varying(255) NOT NULL, "password" character varying(255) NOT NULL, "birthdate" date NULL, "school_university" character varying(255) NULL, "major" character varying(255) NULL, "gpa" numeric(3,2) NULL, "cv_id" uuid NULL, "phone_number" character varying(20) NULL, "portofolio_link" text NULL, "kabupaten_kota_id" character varying(5) NULL, "years_experience" integer NULL, "tech_stack" jsonb NULL, "profile_picture" character varying(255) NULL, "candidate_level" character varying(50) NULL, "recruitment_status_id" uuid NULL, "unavailable_until" date NULL, "role_system_id" uuid NULL, "role_applied_id" uuid NULL, "account_status" character varying(10) NULL, "jabatan" character varying(100) NULL, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NOT NULL, "deleted_at" timestamptz NULL, PRIMARY KEY ("id"));
-- Create index "idx_users_deleted_at" to table: "users"
CREATE INDEX "idx_users_deleted_at" ON "public"."users" ("deleted_at");
-- Create index "idx_users_kabupaten_kota_id" to table: "users"
CREATE INDEX "idx_users_kabupaten_kota_id" ON "public"."users" ("kabupaten_kota_id");
-- Create index "idx_users_role_applied_id" to table: "users"
CREATE INDEX "idx_users_role_applied_id" ON "public"."users" ("role_applied_id");
