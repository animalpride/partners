CREATE TABLE `oauth_clients` (
  `id` int NOT NULL AUTO_INCREMENT,
  `client_id` varchar(120) NOT NULL,
  `name` varchar(120) NOT NULL,
  `description` text,
  `client_secret_hash` varchar(255) NOT NULL,
  `active` tinyint(1) NOT NULL DEFAULT '1',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `oauth_clients_client_id` (`client_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE `client_permissions` (
  `id` int NOT NULL AUTO_INCREMENT,
  `oauth_client_id` int NOT NULL,
  `permission_id` int NOT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `unique_client_permission` (`oauth_client_id`, `permission_id`),
  KEY `client_permissions_permission_id` (`permission_id`),
  CONSTRAINT `client_permissions_ibfk_1` FOREIGN KEY (`oauth_client_id`) REFERENCES `oauth_clients` (`id`) ON DELETE CASCADE,
  CONSTRAINT `client_permissions_ibfk_2` FOREIGN KEY (`permission_id`) REFERENCES `permissions` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

INSERT IGNORE INTO `permissions` (`name`, `description`, `resource`, `action`)
VALUES
  ('partners_applications.read', 'Can read external partner applications', 'partners_applications', 'read'),
  ('partners_applications.write', 'Can update external partner application status', 'partners_applications', 'write');
