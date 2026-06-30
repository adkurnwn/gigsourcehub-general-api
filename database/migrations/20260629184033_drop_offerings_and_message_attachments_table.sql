-- Modify "onboard_histories" table
ALTER TABLE "public"."onboard_histories" DROP COLUMN "offering_id";
-- Drop "message_attachments" table
DROP TABLE "public"."message_attachments";
-- Drop enum type "message_attachment_type"
DROP TYPE "public"."message_attachment_type";
-- Drop "offerings" table
DROP TABLE "public"."offerings";
