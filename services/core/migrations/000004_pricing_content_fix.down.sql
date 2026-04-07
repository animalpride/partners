-- Revert pricing-revenue-share to placeholder content (pre-000003 state).

UPDATE `cms_pages` SET `content_json` = '{"sections":[{"type":"two_column","variant":"default","image_side":"right","background":"","alignment":"left","heading":"Pricing & Revenue Share","body":"Pricing details and revenue share model are being finalized.","bullets":[],"pull_quote":"","button_label":"","button_link":"","image_url":"","image_label":"[Placeholder — Pricing / Revenue Share Illustration]","image_alt":"Pricing illustration","image_aspect_ratio":"4/3"}]}' WHERE `slug` = 'pricing-revenue-share';
