-- Modify "kabupaten_kota" table
ALTER TABLE "public"."kabupaten_kota" ALTER COLUMN "created_at" SET DEFAULT now(), ALTER COLUMN "updated_at" SET DEFAULT now();
-- Modify "provinsi" table
ALTER TABLE "public"."provinsi" ALTER COLUMN "created_at" SET DEFAULT now(), ALTER COLUMN "updated_at" SET DEFAULT now();
