-- Down Migration: Drop invitations and audit events tables

ALTER TABLE `users`
  MODIFY `invited_at` datetime NOT NULL;

DROP TABLE IF EXISTS `audit_events`;
DROP TABLE IF EXISTS `invitations`;
