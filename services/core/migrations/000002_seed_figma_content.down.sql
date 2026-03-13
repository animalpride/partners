-- Revert CMS pages to pre-Figma legacy JSON format.

UPDATE `cms_pages` SET `content_json` = '{"hero":"Grow mission-aligned revenue with embedded commerce","benefits":["No upfront platform cost","Brand-aligned storefront","Revenue share on every order"]}' WHERE `slug` = 'partnership-overview';

UPDATE `cms_pages` SET `content_json` = '{"steps":["Discover your audience needs","Launch branded storefront","Track revenue and engagement"],"examples":["Campaign bundles","Seasonal collections"]}' WHERE `slug` = 'how-it-works';

UPDATE `cms_pages` SET `content_json` = '{"stories":[{"partner":"Sample Humane Society","result":"Increased monthly revenue by 22%"}]}' WHERE `slug` = 'case-studies';

UPDATE `cms_pages` SET `content_json` = '{"revenue_share_min":10,"revenue_share_max":25,"roi_examples":["$5k/mo traffic can yield meaningful mission revenue"]}' WHERE `slug` = 'pricing-revenue-share';

UPDATE `cms_pages` SET `content_json` = '{"faq":[{"q":"How long is onboarding?","a":"Typically 2-4 weeks."},{"q":"Who handles fulfillment?","a":"Animal Pride handles order fulfillment."}]}' WHERE `slug` = 'partner-faq';

UPDATE `cms_pages` SET `content_json` = '{"intro":"Tell us about your organization and goals.","fields":["organization_name","contact_name","email","monthly_traffic","current_store","goals"]}' WHERE `slug` = 'application-contact';
