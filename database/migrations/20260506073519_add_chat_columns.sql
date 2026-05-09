-- Modify "messages" table
ALTER TABLE "public"."messages" ADD COLUMN "read_at" timestamptz NULL;
-- Modify "conversations" table
ALTER TABLE "public"."conversations" ADD COLUMN "subrequest_id" uuid NOT NULL, ADD CONSTRAINT "conversations_subrequest_fk" FOREIGN KEY ("subrequest_id") REFERENCES "public"."subrequests" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Create index "idx_conversations_subrequest_id" to table: "conversations"
CREATE INDEX "idx_conversations_subrequest_id" ON "public"."conversations" ("subrequest_id");
