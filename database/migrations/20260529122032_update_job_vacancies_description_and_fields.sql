-- Modify "job_vacancies" table
ALTER TABLE "public"."job_vacancies" ALTER COLUMN "description" TYPE jsonb USING to_jsonb(description), ADD COLUMN "candidate_qualification" jsonb NULL, ADD COLUMN "benefit" jsonb NULL;
