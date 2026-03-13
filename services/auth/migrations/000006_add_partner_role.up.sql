-- Up Migration: Add partner role
INSERT IGNORE INTO `roles` (`name`, `description`, `active`, `created_at`, `updated_at`)
VALUES ('partner', 'Partner organization user', 1, NOW(), NOW());
