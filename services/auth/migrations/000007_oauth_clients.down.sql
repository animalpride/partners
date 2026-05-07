DELETE FROM `client_permissions`;
DELETE FROM `permissions` WHERE `resource` = 'partners_applications' AND `action` IN ('read', 'write');
DROP TABLE IF EXISTS `client_permissions`;
DROP TABLE IF EXISTS `oauth_clients`;
