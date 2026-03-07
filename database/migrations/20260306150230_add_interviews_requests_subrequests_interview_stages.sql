-- Create enum type "interview_status"
CREATE TYPE "public"."interview_status" AS ENUM ('SCHEDULED', 'CANCELLED', 'RESCHEDULED', 'NO_SHOW');
-- Create enum type "request_urgency"
CREATE TYPE "public"."request_urgency" AS ENUM ('HIGH', 'MIDDLE', 'LOW');
-- Create "interview_stages" table
CREATE TABLE "public"."interview_stages" ("id" uuid NOT NULL, "name" character varying(255) NOT NULL, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NOT NULL, "deleted_at" timestamptz NULL, PRIMARY KEY ("id"));
-- Create index "idx_interview_stages_deleted_at" to table: "interview_stages"
CREATE INDEX "idx_interview_stages_deleted_at" ON "public"."interview_stages" ("deleted_at");
-- Create "requests" table
CREATE TABLE "public"."requests" ("id" uuid NOT NULL, "project_name" character varying(255) NOT NULL, "due_date" date NULL, "admin_user_id" uuid NULL, "employee_user_id" uuid NOT NULL, "required_headcount" integer NOT NULL, "status" character varying(50) NULL, "urgency" "public"."request_urgency" NULL, "fulfillment_date" date NULL, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NOT NULL, "deleted_at" timestamptz NULL, PRIMARY KEY ("id"), CONSTRAINT "requests_admin_user_fk" FOREIGN KEY ("admin_user_id") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "requests_employee_user_fk" FOREIGN KEY ("employee_user_id") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- Create index "idx_requests_admin_user_id" to table: "requests"
CREATE INDEX "idx_requests_admin_user_id" ON "public"."requests" ("admin_user_id");
-- Create index "idx_requests_deleted_at" to table: "requests"
CREATE INDEX "idx_requests_deleted_at" ON "public"."requests" ("deleted_at");
-- Create index "idx_requests_employee_user_id" to table: "requests"
CREATE INDEX "idx_requests_employee_user_id" ON "public"."requests" ("employee_user_id");
-- Create "subrequests" table
CREATE TABLE "public"."subrequests" ("id" uuid NOT NULL, "request_id" uuid NOT NULL, "min_years_experience" integer NULL, "job_role_id" uuid NULL, "tech_stack" jsonb NULL, "notes" text NULL, "is_filled" boolean NOT NULL DEFAULT false, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NOT NULL, "deleted_at" timestamptz NULL, PRIMARY KEY ("id"), CONSTRAINT "subrequests_job_role_fk" FOREIGN KEY ("job_role_id") REFERENCES "public"."job_roles" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, CONSTRAINT "subrequests_request_fk" FOREIGN KEY ("request_id") REFERENCES "public"."requests" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- Create index "idx_subrequests_deleted_at" to table: "subrequests"
CREATE INDEX "idx_subrequests_deleted_at" ON "public"."subrequests" ("deleted_at");
-- Create index "idx_subrequests_job_role_id" to table: "subrequests"
CREATE INDEX "idx_subrequests_job_role_id" ON "public"."subrequests" ("job_role_id");
-- Create index "idx_subrequests_request_id" to table: "subrequests"
CREATE INDEX "idx_subrequests_request_id" ON "public"."subrequests" ("request_id");
-- Create "interviews" table
CREATE TABLE "public"."interviews" ("id" uuid NOT NULL, "candidate_user_id" uuid NOT NULL, "subrequest_id" uuid NOT NULL, "stage_id" uuid NOT NULL, "scheduled_at" timestamptz NULL, "method" character varying(50) NULL, "status" "public"."interview_status" NULL, "meeting_link" character varying(255) NULL, "is_email_sent" boolean NOT NULL DEFAULT false, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NOT NULL, "deleted_at" timestamptz NULL, PRIMARY KEY ("id"), CONSTRAINT "interviews_candidate_user_fk" FOREIGN KEY ("candidate_user_id") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "interviews_stage_fk" FOREIGN KEY ("stage_id") REFERENCES "public"."interview_stages" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, CONSTRAINT "interviews_subrequest_fk" FOREIGN KEY ("subrequest_id") REFERENCES "public"."subrequests" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- Create index "idx_interviews_candidate_user_id" to table: "interviews"
CREATE INDEX "idx_interviews_candidate_user_id" ON "public"."interviews" ("candidate_user_id");
-- Create index "idx_interviews_deleted_at" to table: "interviews"
CREATE INDEX "idx_interviews_deleted_at" ON "public"."interviews" ("deleted_at");
-- Create index "idx_interviews_stage_id" to table: "interviews"
CREATE INDEX "idx_interviews_stage_id" ON "public"."interviews" ("stage_id");
-- Create index "idx_interviews_subrequest_id" to table: "interviews"
CREATE INDEX "idx_interviews_subrequest_id" ON "public"."interviews" ("subrequest_id");
