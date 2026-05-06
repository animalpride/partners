package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/animalpride/partners/services/core/internal/repository"
	"github.com/gin-gonic/gin"
)

type LocationHandler struct {
	repo *repository.LocationRepository
}

func NewLocationHandler(repo *repository.LocationRepository) *LocationHandler {
	return &LocationHandler{repo: repo}
}

func (h *LocationHandler) GetCountries(c *gin.Context) {
	countries, err := h.repo.ListCountries()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch countries"})
		return
	}
	c.JSON(http.StatusOK, countries)
}

func (h *LocationHandler) SearchCityStates(c *gin.Context) {
	countryCode := strings.TrimSpace(c.Query("country_code"))
	query := strings.TrimSpace(c.Query("q"))
	limit := 25

	if rawLimit := strings.TrimSpace(c.Query("limit")); rawLimit != "" {
		parsed, err := strconv.Atoi(rawLimit)
		if err != nil || parsed <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid limit"})
			return
		}
		if parsed > 50 {
			parsed = 50
		}
		limit = parsed
	}

	if len([]rune(query)) < 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query must be at least 2 characters"})
		return
	}

	results, err := h.repo.SearchCityStates(countryCode, query, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to search cities"})
		return
	}
	c.JSON(http.StatusOK, results)
}
