-- Create enum type "review_final_recommendation"
CREATE TYPE "public"."review_final_recommendation" AS ENUM ('HIGHLY_RECOMMENDED', 'RECOMMENDED', 'CONSIDERED', 'NOT_RECOMMENDED');
-- Create "system_settings" table
CREATE TABLE "public"."system_settings" ("id" uuid NOT NULL, "is_ai_mode_enabled" boolean NOT NULL DEFAULT false, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NOT NULL, "deleted_at" timestamptz NULL, PRIMARY KEY ("id"));
-- Create index "idx_system_settings_deleted_at" to table: "system_settings"
CREATE INDEX "idx_system_settings_deleted_at" ON "public"."system_settings" ("deleted_at");
-- Create "reviews" table
CREATE TABLE "public"."reviews" ("id" uuid NOT NULL, "subrequest_id" uuid NOT NULL, "candidate_user_id" uuid NOT NULL, "employee_user_id" uuid NOT NULL, "onboard_history_id" uuid NOT NULL, "work_quality" integer NULL, "timeliness" integer NULL, "communication_collaboration" integer NULL, "problem_solving_initiative" integer NULL, "final_recommendation" "public"."review_final_recommendation" NULL, "notes" text NULL, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NOT NULL, "deleted_at" timestamptz NULL, PRIMARY KEY ("id"), CONSTRAINT "reviews_candidate_user_fk" FOREIGN KEY ("candidate_user_id") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "reviews_employee_user_fk" FOREIGN KEY ("employee_user_id") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "reviews_onboard_history_fk" FOREIGN KEY ("onboard_history_id") REFERENCES "public"."onboard_histories" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "reviews_subrequest_fk" FOREIGN KEY ("subrequest_id") REFERENCES "public"."subrequests" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- Create index "idx_reviews_candidate_user_id" to table: "reviews"
CREATE INDEX "idx_reviews_candidate_user_id" ON "public"."reviews" ("candidate_user_id");
-- Create index "idx_reviews_deleted_at" to table: "reviews"
CREATE INDEX "idx_reviews_deleted_at" ON "public"."reviews" ("deleted_at");
-- Create index "idx_reviews_employee_user_id" to table: "reviews"
CREATE INDEX "idx_reviews_employee_user_id" ON "public"."reviews" ("employee_user_id");
-- Create index "idx_reviews_onboard_history_id" to table: "reviews"
CREATE INDEX "idx_reviews_onboard_history_id" ON "public"."reviews" ("onboard_history_id");
-- Create index "idx_reviews_subrequest_id" to table: "reviews"
CREATE INDEX "idx_reviews_subrequest_id" ON "public"."reviews" ("subrequest_id");
