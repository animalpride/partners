ALTER TABLE `partner_applications`
  ADD COLUMN `referral_source` varchar(100) DEFAULT NULL AFTER `notes`,
  ADD COLUMN `referral_detail` text DEFAULT NULL AFTER `referral_source`,
  ADD KEY `ix_partner_applications_referral_source` (`referral_source`);
