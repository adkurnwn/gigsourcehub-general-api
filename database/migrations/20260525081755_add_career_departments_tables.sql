-- Create "career_departments" table
CREATE TABLE "public"."career_departments" ("id" uuid NOT NULL, "name" character varying(50) NULL, "description" text NULL, "image_path" text NULL, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NOT NULL, "deleted_at" timestamptz NULL, PRIMARY KEY ("id"));
-- Create index "idx_career_departments_deleted_at" to table: "career_departments"
CREATE INDEX "idx_career_departments_deleted_at" ON "public"."career_departments" ("deleted_at");
