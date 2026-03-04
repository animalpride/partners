-- Up Migration: Add invitations and audit events tables

CREATE TABLE `invitations` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `email` varchar(255) NOT NULL,
  `role_id` int NOT NULL,
  `status` varchar(20) NOT NULL,
  `expires_at` datetime NOT NULL,
  `token_hash` varchar(64) NOT NULL,
  `token_nonce` varchar(64) NOT NULL,
  `invited_by_user_id` bigint DEFAULT NULL,
  `accepted_at` datetime DEFAULT NULL,
  `revoked_at` datetime DEFAULT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_invitations_email` (`email`),
  KEY `idx_invitations_token_hash` (`token_hash`),
  KEY `idx_invitations_status` (`status`),
  CONSTRAINT `invitations_role_fk` FOREIGN KEY (`role_id`) REFERENCES `roles` (`id`) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE `audit_events` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `event_type` varchar(100) NOT NULL,
  `actor_user_id` bigint DEFAULT NULL,
  `target_user_id` bigint DEFAULT NULL,
  `target_email` varchar(255) DEFAULT NULL,
  `metadata` json DEFAULT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_audit_actor` (`actor_user_id`),
  KEY `idx_audit_target_user` (`target_user_id`),
  KEY `idx_audit_target_email` (`target_email`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

ALTER TABLE `users`
  MODIFY `invited_at` datetime DEFAULT NULL;
