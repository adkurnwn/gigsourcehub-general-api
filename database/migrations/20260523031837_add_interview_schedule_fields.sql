-- Create enum type "interview_method"
CREATE TYPE "public"."interview_method" AS ENUM ('Online', 'Offline');
-- Modify "interviews" table
ALTER TABLE "public"."interviews" ALTER COLUMN "method" TYPE "public"."interview_method" USING CASE WHEN "method" IN ('Online', 'Offline') THEN "method"::"public"."interview_method" ELSE NULL END, ADD COLUMN "admin_user_id" uuid NULL, ADD COLUMN "title" character varying(255) NULL, ADD COLUMN "description" text NULL, ADD CONSTRAINT "interviews_admin_user_fk" FOREIGN KEY ("admin_user_id") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Create index "idx_interviews_admin_user_id" to table: "interviews"
CREATE INDEX "idx_interviews_admin_user_id" ON "public"."interviews" ("admin_user_id");
