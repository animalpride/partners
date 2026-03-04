INSERT IGNORE INTO `permissions` (`name`, `description`, `resource`, `action`)
VALUES
  ('cms.edit', 'Can edit public partner CMS content', 'cms', 'edit'),
  ('cms.view_leads', 'Can view partner application submissions', 'cms', 'view_leads');

INSERT IGNORE INTO `role_permissions` (`role_id`, `permission_id`)
SELECT r.id, p.id
FROM roles r
JOIN permissions p ON p.name IN ('cms.edit', 'cms.view_leads')
WHERE r.name = 'admin';
