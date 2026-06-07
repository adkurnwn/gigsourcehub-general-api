-- Modify "interviews" table
ALTER TABLE "public"."interviews" ADD COLUMN "is_24h_reminder_sent" boolean NOT NULL DEFAULT false, ADD COLUMN "is_1h_reminder_sent" boolean NOT NULL DEFAULT false;
