package repository

import (
	"errors"
	"sort"
	"time"

	"github.com/animalpride/animalpride-core/services/core/internal/models"
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
		return `{"hero":"Grow mission-aligned revenue with embedded commerce","benefits":["No upfront platform cost","Brand-aligned storefront","Revenue share on every order"]}`
	case "how-it-works":
		return `{"steps":["Discover your audience needs","Launch branded storefront","Track revenue and engagement"],"examples":["Campaign bundles","Seasonal collections"]}`
	case "case-studies":
		return `{"stories":[{"partner":"Sample Humane Society","result":"Increased monthly revenue by 22%"}]}`
	case "pricing-revenue-share":
		return `{"revenue_share_min":10,"revenue_share_max":25,"roi_examples":["$5k/mo traffic can yield meaningful mission revenue"]}`
	case "partner-faq":
		return `{"faq":[{"q":"How long is onboarding?","a":"Typically 2–4 weeks."},{"q":"Who handles fulfillment?","a":"Animal Pride handles order fulfillment."}]}`
	case "application-contact":
		return `{"intro":"Tell us about your organization and goals.","fields":["organization_name","contact_name","email","monthly_traffic","current_store","goals"]}`
	default:
		return `{}`
	}
}
