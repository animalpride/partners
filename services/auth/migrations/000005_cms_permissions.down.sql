DELETE rp
FROM role_permissions rp
JOIN roles r ON rp.role_id = r.id
JOIN permissions p ON rp.permission_id = p.id
WHERE r.name = 'admin' AND p.name IN ('cms.edit', 'cms.view_leads');

DELETE FROM permissions
WHERE name IN ('cms.edit', 'cms.view_leads');
