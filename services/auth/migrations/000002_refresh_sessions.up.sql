-- Up Migration: Create refresh sessions

CREATE TABLE `refresh_sessions` (
  `id` int NOT NULL AUTO_INCREMENT,
  `user_id` bigint NOT NULL,
  `family_id` varchar(64) NOT NULL,
  `current_token_hash` varchar(128) NOT NULL,
  `previous_token_hash` varchar(128) DEFAULT NULL,
  `last_rotated_at` datetime NOT NULL,
  `last_used_at` datetime DEFAULT NULL,
  `expires_at` datetime NOT NULL,
  `revoked_at` datetime DEFAULT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `unique_user` (`user_id`),
  UNIQUE KEY `unique_family` (`family_id`),
  KEY `idx_user` (`user_id`),
  CONSTRAINT `refresh_sessions_ibfk_1` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
