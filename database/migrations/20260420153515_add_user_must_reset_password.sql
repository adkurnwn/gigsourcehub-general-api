-- Modify "users" table
ALTER TABLE "public"."users" ADD COLUMN "must_reset_password" boolean NOT NULL DEFAULT false;
