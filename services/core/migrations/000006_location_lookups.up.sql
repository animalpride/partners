CREATE TABLE `location_countries` (
  `code` char(2) NOT NULL,
  `name` varchar(100) NOT NULL,
  `search_name` varchar(120) NOT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`code`),
  KEY `ix_location_countries_search_name` (`search_name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE `location_states` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `country_code` char(2) NOT NULL,
  `state_code` varchar(20) NOT NULL,
  `name` varchar(120) NOT NULL,
  `search_name` varchar(160) NOT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `ux_location_states_country_state_code` (`country_code`, `state_code`),
  KEY `ix_location_states_country_name` (`country_code`, `search_name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE `location_cities` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `country_code` char(2) NOT NULL,
  `state_code` varchar(20) NOT NULL,
  `state_name` varchar(120) NOT NULL,
  `name` varchar(120) NOT NULL,
  `search_name` varchar(200) NOT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `ix_location_cities_country_state` (`country_code`, `state_code`),
  KEY `ix_location_cities_country_search` (`country_code`, `search_name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE `location_city_aliases` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `city_id` bigint NOT NULL,
  `alias` varchar(160) NOT NULL,
  `search_alias` varchar(160) NOT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `ux_location_city_aliases_city_search_alias` (`city_id`, `search_alias`),
  KEY `ix_location_city_aliases_search_alias` (`search_alias`),
  CONSTRAINT `fk_location_city_aliases_city_id` FOREIGN KEY (`city_id`) REFERENCES `location_cities` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

ALTER TABLE `partner_applications`
  ADD COLUMN `country_code` char(2) DEFAULT NULL AFTER `country`,
  ADD COLUMN `state_code` varchar(20) DEFAULT NULL AFTER `state`,
  ADD COLUMN `city_lookup_id` bigint DEFAULT NULL AFTER `city`,
  ADD KEY `ix_partner_applications_country_code` (`country_code`),
  ADD KEY `ix_partner_applications_state_code` (`state_code`),
  ADD KEY `ix_partner_applications_city_lookup_id` (`city_lookup_id`);
