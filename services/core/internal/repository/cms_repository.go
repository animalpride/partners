package repository

import (
	"errors"
	"sort"
	"time"

	"github.com/animalpride/partners/services/core/internal/models"
	"gorm.io/gorm"
)

type CMSRepository struct {
	db *gorm.DB
}

func NewCMSRepository(db *gorm.DB) *CMSRepository {
	return &CMSRepository{db: db}
}

func (r *CMSRepository) EnsureDefaults() error {
	slugs := []string{
		"partnership-overview",
		"how-it-works",
		"case-studies",
		"pricing-revenue-share",
		"partner-faq",
		"application-contact",
	}

	for _, slug := range slugs {
		var count int64
		if err := r.db.Model(&models.CMSPage{}).Where("slug = ?", slug).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			continue
		}

		page := models.CMSPage{
			Slug:        slug,
			Title:       defaultTitle(slug),
			Description: defaultDescription(slug),
			ContentJSON: defaultContent(slug),
		}
		if err := r.db.Create(&page).Error; err != nil {
			return err
		}
	}
	return nil
}

func (r *CMSRepository) GetAllPages() ([]models.CMSPage, error) {
	var pages []models.CMSPage
	if err := r.db.Order("slug asc").Find(&pages).Error; err != nil {
		return nil, err
	}
	sort.SliceStable(pages, func(i, j int) bool {
		return pageOrder(pages[i].Slug) < pageOrder(pages[j].Slug)
	})
	return pages, nil
}

func (r *CMSRepository) GetPageBySlug(slug string) (*models.CMSPage, error) {
	var page models.CMSPage
	err := r.db.Where("slug = ?", slug).First(&page).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &page, nil
}

func (r *CMSRepository) UpdatePage(slug, title, description, contentJSON string, updatedBy int) (*models.CMSPage, error) {
	page, err := r.GetPageBySlug(slug)
	if err != nil {
		return nil, err
	}
	if page == nil {
		return nil, gorm.ErrRecordNotFound
	}

	page.Title = title
	page.Description = description
	page.ContentJSON = contentJSON
	if updatedBy > 0 {
		page.UpdatedBy = &updatedBy
	}
	page.UpdatedAt = time.Now().UTC()

	if err := r.db.Save(page).Error; err != nil {
		return nil, err
	}
	return page, nil
}

func pageOrder(slug string) int {
	switch slug {
	case "partnership-overview":
		return 1
	case "how-it-works":
		return 2
	case "case-studies":
		return 3
	case "pricing-revenue-share":
		return 4
	case "partner-faq":
		return 5
	case "application-contact":
		return 6
	default:
		return 999
	}
}

func defaultTitle(slug string) string {
	switch slug {
	case "partnership-overview":
		return "Partnership Overview"
	case "how-it-works":
		return "How It Works for Humane Societies"
	case "case-studies":
		return "Case Studies / Success Stories"
	case "pricing-revenue-share":
		return "Pricing & Revenue Share Details"
	case "partner-faq":
		return "Partner FAQ"
	case "application-contact":
		return "Application / Contact"
	default:
		return slug
	}
}

func defaultDescription(slug string) string {
	switch slug {
	case "partnership-overview":
		return "Clear value proposition, partner benefits, and how the embedded e-commerce solution works."
	case "how-it-works":
		return "Step-by-step explanation, real examples, and product integration snapshots."
	case "case-studies":
		return "Highlight partner performance, revenue examples, and testimonials."
	case "pricing-revenue-share":
		return "Transparent revenue share (10–25%) and ROI examples."
	case "partner-faq":
		return "Common onboarding, design support, revenue, and fulfillment questions."
	case "application-contact":
		return "Lead qualification form for new partner applications."
	default:
		return ""
	}
}

func defaultContent(slug string) string {
	switch slug {
	case "partnership-overview":
		return `{"sections":[{"type":"two_column","variant":"hero","image_side":"right","background":"","alignment":"left","heading":"Turn Every Adoption Into a Story — and a Sustainable Revenue Stream","body":"A turnkey e-commerce and printing partnership built for humane societies. Animal Pride partners with humane societies to transform real adoption moments into meaningful, on-brand retail products—while handling everything from e-commerce to printing to fulfillment.","bullets":[],"pull_quote":"","button_label":"Schedule a Partnership Demo","button_link":"/apply","image_url":"","image_label":"Hero — Happy Adopter Holding Newly Adopted Pet | Animal Pride Product Visible (Shirt, Mug, or Tote) | Warm, Candid Lifestyle Photography","image_alt":"Adopter with newly adopted pet","image_aspect_ratio":"16/9"},{"type":"two_column","variant":"default","image_side":"right","background":"","alignment":"left","heading":"The Challenge Humane Societies Face","body":"Adoptions are joyful, emotional moments—but once they happen, engagement often fades. Most humane societies struggle with:","bullets":["Fundraising fatigue from constant donation appeals","Limited staff capacity to manage merchandise programs","Generic logo merch that underperforms","Losing long-term connection with adopters after adoption day"],"pull_quote":"","button_label":"","button_link":"","image_url":"","image_label":"Split Visual — Left: Adoption Day Photo | Right: Empty Inbox / Outdated Merch / Disengaged Supporters","image_alt":"Emotional drop-off after adoption day","image_aspect_ratio":"4/3"},{"type":"two_column","variant":"default","image_side":"right","background":"#f6f7fe","alignment":"left","heading":"Adoption-Centric Retail Fundraising","body":"Animal Pride is not just a printer—it's a fundraising and storytelling partner designed specifically for humane societies. All powered by a fully managed e-commerce shop embedded directly into your website.","bullets":["Celebrate newly adopted pets with personalized, story-driven products","Generate ongoing retail revenue without inventory or overhead","Keep adopters emotionally connected to your organization","Extend your brand visibility through shareable, real-world products"],"pull_quote":"","button_label":"","button_link":"","image_url":"","image_label":"Product Grid Mockup — T-shirt with Pet Photo | Mug: Officially Adopted | Tote Bag with Rescue Branding + Pet Image","image_alt":"Animal Pride product examples","image_aspect_ratio":"4/3"},{"type":"comparison_table","background":"","alignment":"center","heading":"Why Animal Pride Is Different","left_label":"Traditional Merch","right_label":"Animal Pride","rows":[{"left":"Logo-based swag","right":"Real pet stories"},{"left":"Manual setup","right":"Turnkey, done-for-you"},{"left":"Inventory risk","right":"Print-on-demand"},{"left":"Short-lived campaigns","right":"Ongoing revenue"},{"left":"Transactional","right":"Emotional & relational"}],"note":"Animal Pride fills a unique niche: emotion-driven e-commerce fundraising designed specifically around pet adoption."},{"type":"icon_grid","background":"#f6f7fe","alignment":"center","heading":"The Impact for Humane Societies","body":"Every product becomes a conversation—and every conversation spreads your mission.","items":[{"icon":"💙","text":"Increased donor and adopter lifetime value"},{"icon":"🐾","text":"Stronger post-adoption engagement"},{"icon":"👕","text":"Brand visibility through wearable storytelling"},{"icon":"💰","text":"New revenue stream with no staff strain"},{"icon":"📣","text":"Organic marketing via proud pet parents"}],"image_label":"Photo Collage — Real Adopters Wearing or Using Animal Pride Products (Authentic, Social / UGC Style)"},{"type":"form_cta","background":"#00698f","alignment":"center","heading":"Let Your Adoptions Keep Giving","body":"Animal Pride helps humane societies turn love, pride, and storytelling into lasting impact—without adding work, cost, or complexity.","button_label":"Schedule a Partnership Demo","button_link":"/apply"}]}`
	case "how-it-works":
		return `{"sections":[{"type":"two_column","variant":"default","image_side":"left","background":"","alignment":"left","heading":"Step 1: Seamless Website Integration","body":"Animal Pride embeds a branded e-commerce shop directly into your existing website—no redirects, no third-party confusion, no loss of trust.","bullets":[],"pull_quote":"Your brand. Your domain. Our infrastructure.","button_label":"","button_link":"","image_url":"","image_label":"Website Embed Mockup — Humane Society Homepage with Embedded Shop Section | Animal Pride Products Visible Without Leaving the Domain","image_alt":"Website embed mockup","image_aspect_ratio":"16/9"},{"type":"two_column","variant":"default","image_side":"right","background":"#f6f7fe","alignment":"left","heading":"Step 2: Story-Driven Product Creation","body":"We create custom print products featuring recently adopted pets, heartfelt adoption-specific messaging, and apparel, mugs, tote bags, and expandable product lines. Each item becomes a keepsake tied to a real adoption story.","bullets":["Recently adopted pets","Heartfelt, adoption-specific messaging","Apparel, mugs, tote bags, and expandable product lines"],"pull_quote":"","button_label":"","button_link":"","image_url":"","image_label":"Close-up Product Shots — Pet Photo + Name + Adoption Message","image_alt":"Story-driven product close-up","image_aspect_ratio":"4/3"},{"type":"two_column","variant":"default","image_side":"left","background":"","alignment":"left","heading":"Step 3: Fully Managed Printing & Fulfillment","body":"Your team stays focused on your mission—not merch logistics.","bullets":["Printing on demand","Order fulfillment and shipping","Customer service","Inventory risk"],"pull_quote":"No inventory. No logistics. No staff burden.","button_label":"","button_link":"","image_url":"","image_label":"Behind-the-Scenes — Printing in Progress | Packaging with Animal Pride + Partner Branding | Orders Ready to Ship","image_alt":"Printing and fulfillment","image_aspect_ratio":"4/3"},{"type":"two_column","variant":"default","image_side":"right","background":"#f6f7fe","alignment":"left","heading":"Step 4: Ongoing Fundraising Without Campaigns","body":"Instead of one-off fundraisers, you gain passive, year-round retail revenue—products people want to buy, not feel pressured to buy.","bullets":["Passive, year-round retail revenue","Products people want to buy, not feel pressured to buy","Natural post-adoption upsell opportunities"],"pull_quote":"","button_label":"Schedule a Partnership Demo","button_link":"/apply","image_url":"","image_label":"Flow Diagram — Adoption → Product → Purchase → Funds Raised → Mission Impact","image_alt":"Fundraising flow diagram","image_aspect_ratio":"16/9"}]}`
	case "case-studies":
		return `{"sections":[{"type":"two_column","variant":"default","image_side":"right","background":"","alignment":"left","heading":"Partner Success Stories","body":"Case studies and partner results are being gathered. This page will feature real-world examples of humane society partners and the impact of the Animal Pride program.","bullets":[],"pull_quote":"","button_label":"","button_link":"","image_url":"","image_label":"[Placeholder — Partner Story Visual]","image_alt":"Partner case study","image_aspect_ratio":"4/3"}]}`
	case "pricing-revenue-share":
		return `{"sections":[{"type":"two_column","variant":"default","image_side":"right","background":"","alignment":"left","heading":"Pricing & Revenue Share","body":"Pricing details and revenue share model are being finalized. This page will outline the Animal Pride partner compensation structure, ROI examples, and what is included in the program.","bullets":[],"pull_quote":"","button_label":"","button_link":"","image_url":"","image_label":"[Placeholder — Pricing / Revenue Share Illustration]","image_alt":"Pricing illustration","image_aspect_ratio":"4/3"}]}`
	case "partner-faq":
		return `{"sections":[{"type":"two_column","variant":"default","image_side":"right","background":"","alignment":"left","heading":"Partner FAQ","body":"Answers to common partner questions are being compiled. This page will address onboarding, design support, revenue share, fulfillment, and how the program works day-to-day.","bullets":[],"pull_quote":"","button_label":"","button_link":"","image_url":"","image_label":"[Placeholder — FAQ Section Visual]","image_alt":"FAQ illustration","image_aspect_ratio":"4/3"}]}`
	case "application-contact":
		return `{"sections":[{"type":"two_column","variant":"default","image_side":"left","background":"","alignment":"left","heading":"See How Animal Pride Fits Your Humane Society","body":"Tell us about your organization and goals. We will reach out personally to discuss how the Animal Pride partnership can work for you.","bullets":[],"pull_quote":"","button_label":"","button_link":"","image_url":"","image_label":"Connection Moment — Pet + Adopter Looking at Camera (Friendly, Mission-Driven)","image_alt":"Adopter and pet connection moment","image_aspect_ratio":"4/3"},{"type":"application_form","background":"#f6f7fe","alignment":"left","heading":"Request Partnership Details","submit_label":"Request Partnership Details","fields":[{"name":"organization_name","label":"Organization Name","type":"text","required":true,"placeholder":""},{"name":"contact_name","label":"Contact Name","type":"text","required":true,"placeholder":""},{"name":"email","label":"Email Address","type":"email","required":true,"placeholder":""},{"name":"phone","label":"Phone Number","type":"tel","required":true,"placeholder":""},{"name":"address_line1","label":"Address Line 1","type":"text","required":true,"placeholder":"","col":24},{"name":"address_line2","label":"Address Line 2","type":"text","required":false,"placeholder":"","col":24},{"name":"country","label":"Country","type":"text","required":true,"placeholder":""},{"name":"city_state","label":"City / State","type":"text","required":true,"placeholder":"Start typing a city...","col":24},{"name":"postal_code","label":"Postal Code","type":"text","required":true,"placeholder":""},{"name":"website","label":"Website URL","type":"url","required":false,"placeholder":"https://"},{"name":"monthly_traffic","label":"Monthly Adoption Volume","type":"text","required":false,"placeholder":"e.g. 50-100"}]}]}`
	default:
		return `{}`
	}
}
