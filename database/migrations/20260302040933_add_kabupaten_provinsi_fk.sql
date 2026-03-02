-- Modify "kabupaten_kota" table
ALTER TABLE "public"."kabupaten_kota" ADD CONSTRAINT "kabupaten_kota_provinsi_id_fk" FOREIGN KEY ("provinsi_id") REFERENCES "public"."provinsi" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
