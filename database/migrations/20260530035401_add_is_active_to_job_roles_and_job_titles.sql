-- Modify "job_roles" table
ALTER TABLE "public"."job_roles" ADD COLUMN "is_active" boolean NOT NULL DEFAULT true;
-- Modify "job_titles" table
ALTER TABLE "public"."job_titles" ADD COLUMN "is_active" boolean NOT NULL DEFAULT true;
