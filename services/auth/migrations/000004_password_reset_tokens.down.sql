-- Down Migration: Drop password reset tokens

DROP TABLE IF EXISTS `password_reset_tokens`;
