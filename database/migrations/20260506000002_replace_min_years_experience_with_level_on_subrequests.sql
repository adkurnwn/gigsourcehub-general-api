-- Create enum type "subrequest_level"
CREATE TYPE "public"."subrequest_level" AS ENUM ('Junior', 'Middle', 'Senior');
-- Modify "subrequests" table
ALTER TABLE "public"."subrequests" DROP COLUMN "min_years_experience", ADD COLUMN "level" "public"."subrequest_level" NULL;
