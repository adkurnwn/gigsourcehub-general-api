-- Create enum type "message_attachment_type"
CREATE TYPE "public"."message_attachment_type" AS ENUM ('INTERVIEW', 'CONTRACT');
-- Create "contracts" table
CREATE TABLE "public"."contracts" ("id" uuid NOT NULL, "filename" character varying(255) NOT NULL, "user_id_kandidat" uuid NOT NULL, "path" character varying(255) NOT NULL, "subrequest_id" uuid NOT NULL, "start_date" date NULL, "end_date" date NULL, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NOT NULL, "deleted_at" timestamptz NULL, PRIMARY KEY ("id"), CONSTRAINT "contracts_subrequest_fk" FOREIGN KEY ("subrequest_id") REFERENCES "public"."subrequests" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "contracts_user_kandidat_fk" FOREIGN KEY ("user_id_kandidat") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- Create index "idx_contracts_deleted_at" to table: "contracts"
CREATE INDEX "idx_contracts_deleted_at" ON "public"."contracts" ("deleted_at");
-- Create index "idx_contracts_subrequest_id" to table: "contracts"
CREATE INDEX "idx_contracts_subrequest_id" ON "public"."contracts" ("subrequest_id");
-- Create index "idx_contracts_user_id_kandidat" to table: "contracts"
CREATE INDEX "idx_contracts_user_id_kandidat" ON "public"."contracts" ("user_id_kandidat");
-- Create "conversations" table
CREATE TABLE "public"."conversations" ("id" uuid NOT NULL, "admin_user_id" uuid NOT NULL, "candidate_user_id" uuid NOT NULL, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NOT NULL, "deleted_at" timestamptz NULL, PRIMARY KEY ("id"), CONSTRAINT "conversations_admin_user_fk" FOREIGN KEY ("admin_user_id") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "conversations_candidate_user_fk" FOREIGN KEY ("candidate_user_id") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- Create index "idx_conversations_admin_user_id" to table: "conversations"
CREATE INDEX "idx_conversations_admin_user_id" ON "public"."conversations" ("admin_user_id");
-- Create index "idx_conversations_candidate_user_id" to table: "conversations"
CREATE INDEX "idx_conversations_candidate_user_id" ON "public"."conversations" ("candidate_user_id");
-- Create index "idx_conversations_deleted_at" to table: "conversations"
CREATE INDEX "idx_conversations_deleted_at" ON "public"."conversations" ("deleted_at");
-- Create "messages" table
CREATE TABLE "public"."messages" ("id" uuid NOT NULL, "conversation_id" uuid NOT NULL, "content" text NOT NULL, "sender_user_id" uuid NOT NULL, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NOT NULL, "deleted_at" timestamptz NULL, PRIMARY KEY ("id"), CONSTRAINT "messages_conversation_fk" FOREIGN KEY ("conversation_id") REFERENCES "public"."conversations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "messages_sender_user_fk" FOREIGN KEY ("sender_user_id") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- Create index "idx_messages_conversation_id" to table: "messages"
CREATE INDEX "idx_messages_conversation_id" ON "public"."messages" ("conversation_id");
-- Create index "idx_messages_deleted_at" to table: "messages"
CREATE INDEX "idx_messages_deleted_at" ON "public"."messages" ("deleted_at");
-- Create index "idx_messages_sender_user_id" to table: "messages"
CREATE INDEX "idx_messages_sender_user_id" ON "public"."messages" ("sender_user_id");
-- Create "message_attachments" table
CREATE TABLE "public"."message_attachments" ("id" uuid NOT NULL, "message_id" uuid NOT NULL, "attachment_type" "public"."message_attachment_type" NOT NULL, "interview_id" uuid NULL, "contract_id" uuid NULL, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NOT NULL, "deleted_at" timestamptz NULL, PRIMARY KEY ("id"), CONSTRAINT "message_attachments_contract_fk" FOREIGN KEY ("contract_id") REFERENCES "public"."contracts" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, CONSTRAINT "message_attachments_interview_fk" FOREIGN KEY ("interview_id") REFERENCES "public"."interviews" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, CONSTRAINT "message_attachments_message_fk" FOREIGN KEY ("message_id") REFERENCES "public"."messages" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- Create index "idx_message_attachments_deleted_at" to table: "message_attachments"
CREATE INDEX "idx_message_attachments_deleted_at" ON "public"."message_attachments" ("deleted_at");
-- Create index "idx_message_attachments_message_id" to table: "message_attachments"
CREATE INDEX "idx_message_attachments_message_id" ON "public"."message_attachments" ("message_id");
-- Create "onboard_histories" table
CREATE TABLE "public"."onboard_histories" ("id" uuid NOT NULL, "candidate_user_id" uuid NOT NULL, "contract_id" uuid NOT NULL, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NOT NULL, "deleted_at" timestamptz NULL, PRIMARY KEY ("id"), CONSTRAINT "onboard_histories_candidate_user_fk" FOREIGN KEY ("candidate_user_id") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "onboard_histories_contract_fk" FOREIGN KEY ("contract_id") REFERENCES "public"."contracts" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- Create index "idx_onboard_histories_candidate_user_id" to table: "onboard_histories"
CREATE INDEX "idx_onboard_histories_candidate_user_id" ON "public"."onboard_histories" ("candidate_user_id");
-- Create index "idx_onboard_histories_contract_id" to table: "onboard_histories"
CREATE INDEX "idx_onboard_histories_contract_id" ON "public"."onboard_histories" ("contract_id");
-- Create index "idx_onboard_histories_deleted_at" to table: "onboard_histories"
CREATE INDEX "idx_onboard_histories_deleted_at" ON "public"."onboard_histories" ("deleted_at");
