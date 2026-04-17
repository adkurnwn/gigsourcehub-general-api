-- Modify "ai_messages" table
ALTER TABLE "public"."ai_messages" ALTER COLUMN "content" TYPE text, ADD COLUMN "role" character varying(20) NOT NULL;
