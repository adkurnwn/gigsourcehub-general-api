-- Seeder for RECRUITMENT_STATUSES
INSERT INTO recruitment_statuses (id, name, is_active, created_at, updated_at) VALUES
('b1e8463c-3bb9-45e8-8db9-b88304107ef4', 'Contacted', true, NOW(), NOW()),
('648348a2-f8c5-4aba-9f37-18fbbdb2fd20', 'Offering', true, NOW(), NOW()),
('1db9ec40-fdc0-4357-9d7a-1ed6f38fe1cb', 'Accepted', true, NOW(), NOW()),
('3c6f8fb9-63ba-45f8-8422-eec2e38c35b4', 'Unavailable', true, NOW(), NOW()),
('cd7e0892-0b16-43d8-b5e1-0613203f15c7', 'HR Interview', true, NOW(), NOW());

-- Seeder for SECTOR
INSERT INTO sectors (id, name, is_active, created_at, updated_at) VALUES
('9d07bd24-2c67-4bd9-948f-36cd66ea8d65', 'Human Resources', true, NOW(), NOW()),
('ff744111-e2e7-49bc-9528-ebbbfa43d607', 'Information Technology', true, NOW(), NOW());

-- Seeder for ROLE_SYSTEMS
INSERT INTO role_systems (id, name, created_at, updated_at) VALUES
('593fc0c5-e51c-4b55-aab7-aeaf8e8ece2a', 'Candidate', NOW(), NOW()),
('d2c7302f-b4de-4d43-9dd1-b0e6e22fb2ea', 'Admin', NOW(), NOW()),
('eac7b8e1-e123-4567-bd04-ae6305aabfa4', 'Superadmin', NOW(), NOW()),
('5a72fd34-4b95-46c9-ae5f-b51f8a842880', 'Employee', NOW(), NOW());

-- Seeder for JOB_ROLES (Linking to SECTOR 'Information Technology')
INSERT INTO job_roles (id, name, sector_id, created_at, updated_at) VALUES
('d1a93ca3-c283-4a87-a8ac-fbee426685f0', 'Frontend Developer', 'ff744111-e2e7-49bc-9528-ebbbfa43d607', NOW(), NOW());
