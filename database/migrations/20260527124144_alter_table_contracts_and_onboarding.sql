-- Add value to enum type: "message_attachment_type"
ALTER TYPE "public"."message_attachment_type" ADD VALUE 'OFFERING' AFTER 'INTERVIEW';
-- Add value to enum type: "approval_request_table_name"
ALTER TYPE "public"."approval_request_table_name" ADD VALUE 'career_departments';
-- Create "offerings" table
CREATE TABLE "public"."offerings" ("id" uuid NOT NULL, "filename" character varying(255) NOT NULL, "user_id_kandidat" uuid NOT NULL, "path" character varying(255) NOT NULL, "subrequest_id" uuid NOT NULL, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NOT NULL, "deleted_at" timestamptz NULL, PRIMARY KEY ("id"), CONSTRAINT "offerings_subrequest_fk" FOREIGN KEY ("subrequest_id") REFERENCES "public"."subrequests" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "offerings_user_kandidat_fk" FOREIGN KEY ("user_id_kandidat") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- Create index "idx_offerings_deleted_at" to table: "offerings"
CREATE INDEX "idx_offerings_deleted_at" ON "public"."offerings" ("deleted_at");
-- Create index "idx_offerings_subrequest_id" to table: "offerings"
CREATE INDEX "idx_offerings_subrequest_id" ON "public"."offerings" ("subrequest_id");
-- Create index "idx_offerings_user_id_kandidat" to table: "offerings"
CREATE INDEX "idx_offerings_user_id_kandidat" ON "public"."offerings" ("user_id_kandidat");
-- Rename a column from "contract_id" to "offering_id"
ALTER TABLE "public"."message_attachments" RENAME COLUMN "contract_id" TO "offering_id";
-- Modify "message_attachments" table
ALTER TABLE "public"."message_attachments" DROP CONSTRAINT "message_attachments_contract_fk", ADD CONSTRAINT "message_attachments_offering_fk" FOREIGN KEY ("offering_id") REFERENCES "public"."offerings" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "onboard_histories" table
ALTER TABLE "public"."onboard_histories" DROP COLUMN "contract_id", ADD COLUMN "offering_id" uuid NULL, ADD COLUMN "start_date" date NULL, ADD COLUMN "end_date" date NULL, ADD COLUMN "snapshot" jsonb NULL, ADD CONSTRAINT "onboard_histories_offering_fk" FOREIGN KEY ("offering_id") REFERENCES "public"."offerings" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Create index "idx_onboard_histories_offering_id" to table: "onboard_histories"
CREATE INDEX "idx_onboard_histories_offering_id" ON "public"."onboard_histories" ("offering_id");
-- Drop "contracts" table
DROP TABLE "public"."contracts";
