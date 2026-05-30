-- Create enum type "review_indicator"
CREATE TYPE "public"."review_indicator" AS ENUM (
  'WORK_QUALITY',
  'TIMELINESS',
  'COMMUNICATION_COLLABORATION',
  'PROBLEM_SOLVING_INITIATIVE'
);

-- Create "review_questions" table
CREATE TABLE "public"."review_questions" (
  "id" uuid NOT NULL,
  "indicator" "public"."review_indicator" NOT NULL,
  "question_text" text NOT NULL,
  "question_order" integer NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id")
);

-- Create indexes for "review_questions"
CREATE INDEX "idx_review_questions_deleted_at"
  ON "public"."review_questions" ("deleted_at");

CREATE INDEX "idx_review_questions_indicator"
  ON "public"."review_questions" ("indicator");

CREATE UNIQUE INDEX "uq_review_questions_indicator_order"
  ON "public"."review_questions" ("indicator", "question_order");

-- Create "review_answers" table
CREATE TABLE "public"."review_answers" (
  "id" uuid NOT NULL,
  "review_id" uuid NOT NULL,
  "question_id" uuid NOT NULL,
  "score" integer NOT NULL CHECK ("score" BETWEEN 1 AND 5),
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "review_answers_review_fk"
    FOREIGN KEY ("review_id")
    REFERENCES "public"."reviews" ("id")
    ON DELETE CASCADE,
  CONSTRAINT "review_answers_question_fk"
    FOREIGN KEY ("question_id")
    REFERENCES "public"."review_questions" ("id")
    ON DELETE CASCADE
);

-- Create indexes for "review_answers"
CREATE INDEX "idx_review_answers_deleted_at"
  ON "public"."review_answers" ("deleted_at");

CREATE INDEX "idx_review_answers_review_id"
  ON "public"."review_answers" ("review_id");

CREATE INDEX "idx_review_answers_question_id"
  ON "public"."review_answers" ("question_id");

CREATE UNIQUE INDEX "uq_review_answers_review_question"
  ON "public"."review_answers" ("review_id", "question_id");

-- Update "reviews" table to remove old indicator columns
ALTER TABLE "public"."reviews" DROP COLUMN IF EXISTS "work_quality";
ALTER TABLE "public"."reviews" DROP COLUMN IF EXISTS "timeliness";
ALTER TABLE "public"."reviews" DROP COLUMN IF EXISTS "communication_collaboration";
ALTER TABLE "public"."reviews" DROP COLUMN IF EXISTS "problem_solving_initiative";

-- Add unique constraint for one review per onboarding per employee
CREATE UNIQUE INDEX "uq_reviews_onboard_employee"
  ON "public"."reviews" ("onboard_history_id", "employee_user_id");

-- Seed review questions
INSERT INTO "public"."review_questions"
  ("id", "indicator", "question_text", "question_order", "created_at", "updated_at")
VALUES
  ('11111111-1111-1111-1111-111111111101', 'WORK_QUALITY',
   'Sejauh mana hasil pekerjaan freelancer sesuai dengan requirement dan spesifikasi yang telah ditentukan?', 1, now(), now()),
  ('11111111-1111-1111-1111-111111111102', 'WORK_QUALITY',
   'Seberapa rendah tingkat bug atau kesalahan yang ditemukan pada hasil pekerjaan freelancer?', 2, now(), now()),
  ('11111111-1111-1111-1111-111111111103', 'WORK_QUALITY',
   'Seberapa rapi dan mudah dipahami struktur hasil kerja (misalnya kode, dokumen, atau output lainnya) yang dihasilkan freelancer?', 3, now(), now()),
  ('11111111-1111-1111-1111-111111111104', 'WORK_QUALITY',
   'Sejauh mana hasil akhir pekerjaan freelancer memenuhi standar kualitas yang ditetapkan perusahaan?', 4, now(), now()),

  ('11111111-1111-1111-1111-111111111105', 'TIMELINESS',
   'Sejauh mana freelancer menyelesaikan tugas sesuai dengan deadline yang telah disepakati?', 1, now(), now()),
  ('11111111-1111-1111-1111-111111111106', 'TIMELINESS',
   'Seberapa cepat freelancer menindaklanjuti revisi atau permintaan perbaikan yang diberikan?', 2, now(), now()),
  ('11111111-1111-1111-1111-111111111107', 'TIMELINESS',
   'Sejauh mana freelancer mampu mengatur prioritas kerja ketika menangani beberapa tugas sekaligus?', 3, now(), now()),
  ('11111111-1111-1111-1111-111111111108', 'TIMELINESS',
   'Seberapa jarang freelancer mengalami keterlambatan tanpa alasan yang jelas?', 4, now(), now()),

  ('11111111-1111-1111-1111-111111111109', 'COMMUNICATION_COLLABORATION',
   'Sejauh mana freelancer responsif dalam berkomunikasi melalui chat, email, atau meeting?', 1, now(), now()),
  ('11111111-1111-1111-1111-111111111110', 'COMMUNICATION_COLLABORATION',
   'Seberapa proaktif freelancer mengajukan pertanyaan ketika terdapat instruksi atau kebutuhan yang belum jelas?', 2, now(), now()),
  ('11111111-1111-1111-1111-111111111111', 'COMMUNICATION_COLLABORATION',
   'Sejauh mana freelancer terbuka dan menerima feedback atau saran perbaikan yang diberikan?', 3, now(), now()),
  ('11111111-1111-1111-1111-111111111112', 'COMMUNICATION_COLLABORATION',
   'Seberapa baik freelancer dapat bekerja sama dan berkoordinasi dengan anggota tim lainnya?', 4, now(), now()),

  ('11111111-1111-1111-1111-111111111113', 'PROBLEM_SOLVING_INITIATIVE',
   'Sejauh mana freelancer berusaha mencari solusi secara mandiri sebelum meminta bantuan?', 1, now(), now()),
  ('11111111-1111-1111-1111-111111111114', 'PROBLEM_SOLVING_INITIATIVE',
   'Seberapa besar inisiatif freelancer dalam mengambil tindakan untuk menyelesaikan masalah yang muncul?', 2, now(), now()),
  ('11111111-1111-1111-1111-111111111115', 'PROBLEM_SOLVING_INITIATIVE',
   'Sejauh mana freelancer memberikan saran atau improvement yang relevan terhadap project yang dikerjakan?', 3, now(), now()),
  ('11111111-1111-1111-1111-111111111116', 'PROBLEM_SOLVING_INITIATIVE',
   'Seberapa baik freelancer menangani kendala teknis secara tenang, sistematis, dan terstruktur?', 4, now(), now());