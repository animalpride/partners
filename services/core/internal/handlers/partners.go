package handlers

import (
	"crypto/rand"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/animalpride/partners/services/core/internal/models"
	"github.com/animalpride/partners/services/core/internal/repository"
	"github.com/gin-gonic/gin"
)

var phoneRegex = regexp.MustCompile(`^\+?[\d\s\-(). ]{7,20}$`)

type PartnerLeadMailer interface {
	SendPartnerLead(app *models.PartnerApplication) error
}

type PartnerHandler struct {
	appRepo      *repository.PartnerApplicationRepository
	locationRepo *repository.LocationRepository
	mailer       PartnerLeadMailer
}

func NewPartnerHandler(appRepo *repository.PartnerApplicationRepository, locationRepo *repository.LocationRepository, mailer PartnerLeadMailer) *PartnerHandler {
	return &PartnerHandler{appRepo: appRepo, locationRepo: locationRepo, mailer: mailer}
}

func generateUUID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	b[6] = (b[6] & 0x0f) | 0x40 // version 4
	b[8] = (b[8] & 0x3f) | 0x80 // variant
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

func (h *PartnerHandler) SubmitLead(c *gin.Context) {
	var req struct {
		OrganizationName string `json:"organization_name" binding:"required"`
		ContactName      string `json:"contact_name" binding:"required"`
		Email            string `json:"email" binding:"required,email"`
		Phone            string `json:"phone" binding:"required"`
		AddressLine1     string `json:"address_line1" binding:"required"`
		AddressLine2     string `json:"address_line2"`
		CityState        string `json:"city_state"`
		CityLookupID     uint   `json:"city_lookup_id"`
		City             string `json:"city"`
		State            string `json:"state"`
		StateCode        string `json:"state_code"`
		PostalCode       string `json:"postal_code" binding:"required"`
		Country          string `json:"country" binding:"required"`
		CountryCode      string `json:"country_code"`
		Website          string `json:"website"`
		MonthlyTraffic   string `json:"monthly_traffic"`
		CurrentStore     string `json:"current_store"`
		Goals            string `json:"goals"`
		Notes            string `json:"notes"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	if !phoneRegex.MatchString(strings.TrimSpace(req.Phone)) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid phone number format"})
		return
	}

	country, err := h.locationRepo.FindCountryByCodeOrName(firstNonEmpty(req.CountryCode, req.Country))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to validate country"})
		return
	}
	if country == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid country"})
		return
	}

	var city *models.LocationCity
	if req.CityLookupID > 0 {
		city, err = h.locationRepo.FindCityByID(country.Code, req.CityLookupID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to validate city"})
			return
		}
	} else {
		cityName, stateInput := parseCityState(firstNonEmpty(req.CityState, fmt.Sprintf("%s, %s", req.City, req.State)))
		if cityName == "" {
			cityName = req.City
		}
		if stateInput == "" {
			stateInput = firstNonEmpty(req.StateCode, req.State)
		}
		city, err = h.locationRepo.FindCityByName(country.Code, cityName, stateInput)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to validate city"})
			return
		}
	}
	if city == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid city/state selection for country"})
		return
	}

	if strings.TrimSpace(req.CityState) == "" {
		req.CityState = fmt.Sprintf("%s, %s", city.Name, city.StateCode)
	}

	app := &models.PartnerApplication{
		ID:               generateUUID(),
		OrganizationName: req.OrganizationName,
		ContactName:      req.ContactName,
		Email:            req.Email,
		Phone:            req.Phone,
		AddressLine1:     req.AddressLine1,
		AddressLine2:     req.AddressLine2,
		City:             city.Name,
		CityLookupID:     &city.ID,
		State:            city.StateName,
		StateCode:        city.StateCode,
		PostalCode:       req.PostalCode,
		Country:          country.Name,
		CountryCode:      country.Code,
		Website:          req.Website,
		MonthlyTraffic:   req.MonthlyTraffic,
		CurrentStore:     req.CurrentStore,
		Goals:            req.Goals,
		Notes:            req.Notes,
		DenopsSyncStatus: "pending",
	}

	if err := h.appRepo.Create(app); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to submit application"})
		return
	}

	if err := h.mailer.SendPartnerLead(app); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "application saved but failed to send notification"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "application submitted"})
}

func parseCityState(value string) (string, string) {
	parts := strings.Split(value, ",")
	if len(parts) < 2 {
		return strings.TrimSpace(value), ""
	}
	return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			return trimmed
		}
	}
	return ""
}
