-- Down Migration: Remove partner role
DELETE FROM `roles` WHERE `name` = 'partner';
