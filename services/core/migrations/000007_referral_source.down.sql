ALTER TABLE `partner_applications`
  DROP KEY `ix_partner_applications_referral_source`,
  DROP COLUMN `referral_detail`,
  DROP COLUMN `referral_source`;
