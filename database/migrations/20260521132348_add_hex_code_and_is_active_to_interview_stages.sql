-- Modify "interview_stages" table
ALTER TABLE "public"."interview_stages" ADD COLUMN "hex_code" character varying(10) NULL, ADD COLUMN "is_active" boolean NOT NULL DEFAULT true;
