-- Modify "onboard_histories" table
ALTER TABLE "public"."onboard_histories" ADD COLUMN "is_expiry_notification_sent" boolean NOT NULL DEFAULT false;
