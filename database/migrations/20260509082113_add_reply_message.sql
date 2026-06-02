-- Modify "messages" table
ALTER TABLE "public"."messages" ADD COLUMN "reply_to_message_id" uuid NULL, ADD CONSTRAINT "messages_reply_to_message_fk" FOREIGN KEY ("reply_to_message_id") REFERENCES "public"."messages" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Create index "idx_messages_reply_to_message_id" to table: "messages"
CREATE INDEX "idx_messages_reply_to_message_id" ON "public"."messages" ("reply_to_message_id");
