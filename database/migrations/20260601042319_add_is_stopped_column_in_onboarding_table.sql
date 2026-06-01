-- Modify "onboard_histories" table
ALTER TABLE "public"."onboard_histories" ADD COLUMN "is_stopped" boolean NOT NULL DEFAULT false;
