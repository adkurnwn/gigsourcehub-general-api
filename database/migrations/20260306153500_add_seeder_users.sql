-- Seeder for USERS (one per system role)
-- Password for all users: "securepassword" (bcrypt hashed)
-- You may want to rehash the password manually before applying this migration.

INSERT INTO users (id, name, email, password, system_role_id, account_status, created_at, updated_at) VALUES
-- Candidate
('a3d7e84b-1f29-4c6a-b8e3-5c92d0f17a64', 'Kandidat Test', 'kandidat@test.com', '$2a$10$lXNVx7vkXObWmqqMvBXIY.QNv4qIvlYp6.CZg9to7MsvSlmtcry2a', '593fc0c5-e51c-4b55-aab7-aeaf8e8ece2a', 'Active', NOW(), NOW()),
-- Admin (HR)
('7b2f9c16-e4a8-43d5-9071-6d83a4e5bf02', 'HR Test', 'hr@test.com', '$2a$10$lXNVx7vkXObWmqqMvBXIY.QNv4qIvlYp6.CZg9to7MsvSlmtcry2a', 'd2c7302f-b4de-4d43-9dd1-b0e6e22fb2ea', 'Active', NOW(), NOW()),
-- Superadmin
('e1c84f63-5a0d-4b72-ae97-3f8b21d6c4e9', 'Superadmin Test', 'superadmin@test.com', '$2a$10$lXNVx7vkXObWmqqMvBXIY.QNv4qIvlYp6.CZg9to7MsvSlmtcry2a', 'eac7b8e1-e123-4567-bd04-ae6305aabfa4', 'Active', NOW(), NOW()),
-- Employee (Pegawai)
('4f6d82a1-c9b3-47e0-8a15-d27e3b94f5c8', 'Pegawai Test', 'pegawai@test.com', '$2a$10$lXNVx7vkXObWmqqMvBXIY.QNv4qIvlYp6.CZg9to7MsvSlmtcry2a', '5a72fd34-4b95-46c9-ae5f-b51f8a842880', 'Active', NOW(), NOW());
