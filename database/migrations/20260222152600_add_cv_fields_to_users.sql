ALTER TABLE "public"."users" 
  ADD COLUMN "pendidikan_terakhir" character varying(255) NULL,
  ADD COLUMN "instansi_pendidikan" character varying(255) NULL,
  ADD COLUMN "jurusan" character varying(255) NULL,
  ADD COLUMN "ipk" character varying(50) NULL,
  ADD COLUMN "kabupaten" character varying(255) NULL,
  ADD COLUMN "provinsi" character varying(255) NULL,
  ADD COLUMN "lama_pengalaman_kerja" character varying(255) NULL,
  ADD COLUMN "bidang_minat" character varying(255) NULL,
  ADD COLUMN "applied_role" character varying(255) NULL,
  ADD COLUMN "skills" jsonb NULL,
  ADD COLUMN "link_portofolio" character varying(255) NULL;
