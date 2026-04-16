-- Create "company_profiles" table
CREATE TABLE "public"."company_profiles" ("id" uuid NOT NULL, "address" text NULL, "phone" character varying(20) NULL, "email" character varying(255) NULL, "facebook_url" text NULL, "instagram_url" text NULL, "linkedin_url" text NULL, "twitter_url" text NULL, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NOT NULL, "deleted_at" timestamptz NULL, PRIMARY KEY ("id"));
-- Create index "idx_company_profiles_deleted_at" to table: "company_profiles"
CREATE INDEX "idx_company_profiles_deleted_at" ON "public"."company_profiles" ("deleted_at");
-- Create "faqs" table
CREATE TABLE "public"."faqs" ("id" uuid NOT NULL, "question" text NOT NULL, "answer" text NOT NULL, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NOT NULL, "deleted_at" timestamptz NULL, PRIMARY KEY ("id"));
-- Create index "idx_faqs_deleted_at" to table: "faqs"
CREATE INDEX "idx_faqs_deleted_at" ON "public"."faqs" ("deleted_at");
-- Create enum type "approval_request_table_name"
CREATE TYPE "public"."approval_request_table_name" AS ENUM ('faqs', 'job_vacancies', 'company_profiles');
-- Create enum type "approval_request_status"
CREATE TYPE "public"."approval_request_status" AS ENUM ('PENDING', 'APPROVED', 'REJECTED');
-- Create enum type "job_vacancy_status"
CREATE TYPE "public"."job_vacancy_status" AS ENUM ('DRAFT', 'ARCHIVED', 'PUBLISHED');
-- Modify "interviews" table
ALTER TABLE "public"."interviews" ALTER COLUMN "meeting_link" TYPE text, ADD COLUMN "meeting_location" text NULL;
-- Modify "requests" table
ALTER TABLE "public"."requests" ADD COLUMN "rejected_reason" text NULL;
-- Create enum type "job_vacancy_schema"
CREATE TYPE "public"."job_vacancy_schema" AS ENUM ('ONSITE', 'REMOTE', 'HYBRID');
-- Create enum type "approval_request_action"
CREATE TYPE "public"."approval_request_action" AS ENUM ('CREATE', 'UPDATE', 'DELETE');
-- Create "admin_notes" table
CREATE TABLE "public"."admin_notes" ("id" uuid NOT NULL, "admin_user_id" uuid NOT NULL, "candidate_user_id" uuid NOT NULL, "content" text NOT NULL, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NOT NULL, "deleted_at" timestamptz NULL, PRIMARY KEY ("id"), CONSTRAINT "admin_notes_admin_fk" FOREIGN KEY ("admin_user_id") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "admin_notes_candidate_fk" FOREIGN KEY ("candidate_user_id") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- Create index "idx_admin_notes_deleted_at" to table: "admin_notes"
CREATE INDEX "idx_admin_notes_deleted_at" ON "public"."admin_notes" ("deleted_at");
-- Create "ai_chats" table
CREATE TABLE "public"."ai_chats" ("id" uuid NOT NULL, "admin_user_id" uuid NOT NULL, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NOT NULL, "deleted_at" timestamptz NULL, PRIMARY KEY ("id"), CONSTRAINT "ai_chats_admin_fk" FOREIGN KEY ("admin_user_id") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- Create index "idx_ai_chats_deleted_at" to table: "ai_chats"
CREATE INDEX "idx_ai_chats_deleted_at" ON "public"."ai_chats" ("deleted_at");
-- Create "ai_messages" table
CREATE TABLE "public"."ai_messages" ("id" uuid NOT NULL, "ai_chat_id" uuid NOT NULL, "content" json NOT NULL, "is_last_message" boolean NOT NULL DEFAULT false, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NOT NULL, "deleted_at" timestamptz NULL, PRIMARY KEY ("id"), CONSTRAINT "ai_messages_chat_fk" FOREIGN KEY ("ai_chat_id") REFERENCES "public"."ai_chats" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- Create index "idx_ai_messages_deleted_at" to table: "ai_messages"
CREATE INDEX "idx_ai_messages_deleted_at" ON "public"."ai_messages" ("deleted_at");
-- Create "approval_requests" table
CREATE TABLE "public"."approval_requests" ("id" uuid NOT NULL, "requested_by_admin_id" uuid NOT NULL, "table_name" "public"."approval_request_table_name" NOT NULL, "record_id" uuid NOT NULL, "action" "public"."approval_request_action" NOT NULL, "proposed_data" jsonb NULL, "status" "public"."approval_request_status" NOT NULL DEFAULT 'PENDING', "rejected_reason" text NULL, "reviewed_by_superadmin_id" uuid NULL, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NOT NULL, "deleted_at" timestamptz NULL, PRIMARY KEY ("id"), CONSTRAINT "approval_req_admin_fk" FOREIGN KEY ("requested_by_admin_id") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "approval_req_superadmin_fk" FOREIGN KEY ("reviewed_by_superadmin_id") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL);
-- Create index "idx_approval_requests_deleted_at" to table: "approval_requests"
CREATE INDEX "idx_approval_requests_deleted_at" ON "public"."approval_requests" ("deleted_at");
-- Modify "subrequests" table
ALTER TABLE "public"."subrequests" ADD COLUMN "overview" text NULL;
-- Create "job_vacancies" table
CREATE TABLE "public"."job_vacancies" ("id" uuid NOT NULL, "subrequest_id" uuid NOT NULL, "name" character varying(255) NOT NULL, "takedown_date" date NULL, "fulfillment_date" date NULL, "schema" "public"."job_vacancy_schema" NULL, "status" "public"."job_vacancy_status" NULL, "description" character varying(50) NULL, "overview" text NULL, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NOT NULL, "deleted_at" timestamptz NULL, PRIMARY KEY ("id"), CONSTRAINT "job_vacancies_subrequest_fk" FOREIGN KEY ("subrequest_id") REFERENCES "public"."subrequests" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- Create index "idx_job_vacancies_deleted_at" to table: "job_vacancies"
CREATE INDEX "idx_job_vacancies_deleted_at" ON "public"."job_vacancies" ("deleted_at");
-- Create "log_activities" table
CREATE TABLE "public"."log_activities" ("id" uuid NOT NULL, "actor_id" uuid NOT NULL, "action_type" character varying(100) NOT NULL, "module" character varying(100) NOT NULL, "description" text NULL, "metadata" jsonb NULL, "ip_address" character varying(45) NULL, "is_success" boolean NOT NULL, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NOT NULL, "deleted_at" timestamptz NULL, PRIMARY KEY ("id"), CONSTRAINT "log_activities_actor_fk" FOREIGN KEY ("actor_id") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- Create index "idx_log_activities_deleted_at" to table: "log_activities"
CREATE INDEX "idx_log_activities_deleted_at" ON "public"."log_activities" ("deleted_at");
-- Create "notifications" table
CREATE TABLE "public"."notifications" ("id" uuid NOT NULL, "admin_user_id_owner" uuid NOT NULL, "title" character varying(255) NOT NULL, "description" text NULL, "is_read" boolean NOT NULL DEFAULT false, "is_admin_broadcast" boolean NOT NULL DEFAULT false, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NOT NULL, "deleted_at" timestamptz NULL, PRIMARY KEY ("id"), CONSTRAINT "notifications_admin_fk" FOREIGN KEY ("admin_user_id_owner") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- Create index "idx_notifications_deleted_at" to table: "notifications"
CREATE INDEX "idx_notifications_deleted_at" ON "public"."notifications" ("deleted_at");
-- Create "subrequest_candidates" table
CREATE TABLE "public"."subrequest_candidates" ("id" uuid NOT NULL, "subrequest_id" uuid NOT NULL, "candidate_user_id" uuid NOT NULL, "name" character varying(150) NOT NULL, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NOT NULL, "deleted_at" timestamptz NULL, PRIMARY KEY ("id"), CONSTRAINT "subrequest_candidates_subrequest_fk" FOREIGN KEY ("subrequest_id") REFERENCES "public"."subrequests" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "subrequest_candidates_user_fk" FOREIGN KEY ("candidate_user_id") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- Create index "idx_subrequest_candidates_deleted_at" to table: "subrequest_candidates"
CREATE INDEX "idx_subrequest_candidates_deleted_at" ON "public"."subrequest_candidates" ("deleted_at");
