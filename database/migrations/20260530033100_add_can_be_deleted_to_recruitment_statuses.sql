-- Modify "recruitment_statuses" table
ALTER TABLE "public"."recruitment_statuses" ADD COLUMN "can_be_deleted" boolean NOT NULL DEFAULT true;
-- Seed / Update default recruitment statuses
INSERT INTO "recruitment_statuses" ("id", "name", "hex_code", "is_active", "can_be_deleted", "created_at", "updated_at")
VALUES
  ('1db9ec40-fdc0-4357-9d7a-1ed6f38fe1cb', 'Accepted', '#4CAF50', true, false, NOW(), NOW()),
  ('20a9a4b3-c12e-4b48-9c16-2917e761df8f', 'Assigned', '#2196F3', true, false, NOW(), NOW()),
  ('b1e8463c-3bb9-45e8-8db9-b88304107ef4', 'Contacted', '#FF9800', true, false, NOW(), NOW()),
  ('d7195e34-58e1-4545-9835-7ffc5c2d1ad2', 'Decline', '#F44336', true, false, NOW(), NOW())
ON CONFLICT ("id") DO UPDATE SET
  "name" = EXCLUDED."name",
  "can_be_deleted" = false,
  "updated_at" = NOW();