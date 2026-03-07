-- Create enum type "request_status"
CREATE TYPE "public"."request_status" AS ENUM ('PENDING', 'ACCEPTED', 'PROCESSING', 'REJECTED', 'DONE');
-- Modify "requests" table
ALTER TABLE "public"."requests" ALTER COLUMN "status" TYPE "public"."request_status" USING "status"::"public"."request_status", ALTER COLUMN "status" SET NOT NULL, ALTER COLUMN "status" SET DEFAULT 'PENDING';
