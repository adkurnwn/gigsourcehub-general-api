-- Modify "cvs" table
ALTER TABLE "public"."cvs" ADD CONSTRAINT "cv_user_fk" FOREIGN KEY ("user_id") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "job_roles" table
ALTER TABLE "public"."job_roles" ADD CONSTRAINT "job_roles_sector_fk" FOREIGN KEY ("sector_id") REFERENCES "public"."sectors" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "users" table
ALTER TABLE "public"."users" ADD CONSTRAINT "user_kabupaten_kota_fk" FOREIGN KEY ("kabupaten_kota_id") REFERENCES "public"."kabupaten_kota" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "user_recruitment_status_fk" FOREIGN KEY ("recruitment_status_id") REFERENCES "public"."recruitment_statuses" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "user_assigned_role_fk" FOREIGN KEY ("assigned_role_id") REFERENCES "public"."job_roles" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "user_role_system_fk" FOREIGN KEY ("role_system_id") REFERENCES "public"."role_systems" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "user_has_job_roles" table
ALTER TABLE "public"."user_has_job_roles" ADD CONSTRAINT "user_has_job_roles_user_fk" FOREIGN KEY ("user_id") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "user_has_job_roles_job_role_fk" FOREIGN KEY ("job_role_id") REFERENCES "public"."job_roles" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
