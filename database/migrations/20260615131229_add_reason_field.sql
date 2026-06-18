-- Modify "onboard_histories" table
ALTER TABLE "public"."onboard_histories" ADD COLUMN "cancelled_reason" text NULL;
-- Modify "subrequest_candidates" table
ALTER TABLE "public"."subrequest_candidates" ADD COLUMN "declined_reason" text NULL;
