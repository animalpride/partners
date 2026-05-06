ALTER TABLE `partner_applications`
  DROP KEY `ix_partner_applications_city_lookup_id`,
  DROP KEY `ix_partner_applications_state_code`,
  DROP KEY `ix_partner_applications_country_code`,
  DROP COLUMN `city_lookup_id`,
  DROP COLUMN `state_code`,
  DROP COLUMN `country_code`;

DROP TABLE IF EXISTS `location_city_aliases`;
DROP TABLE IF EXISTS `location_cities`;
DROP TABLE IF EXISTS `location_states`;
DROP TABLE IF EXISTS `location_countries`;
