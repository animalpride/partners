package handlers

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

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

func (h *LocationHandler) RefreshLocations(c *gin.Context) {
	var req struct {
		CombinedURL  string `json:"combined_url"`
		CountriesURL string `json:"countries_url"`
		StatesURL    string `json:"states_url"`
		CitiesURL    string `json:"cities_url"`
		WaitSeconds  int    `json:"wait_seconds"`
	}

	if c.Request.ContentLength > 0 {
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
			return
		}
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 15*time.Minute)
	defer cancel()

	imported, err := h.repo.RefreshFromGitHub(ctx, req.CombinedURL, req.CountriesURL, req.StatesURL, req.CitiesURL, req.WaitSeconds)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to refresh locations"})
		return
	}
	if !imported {
		c.JSON(http.StatusConflict, gin.H{"error": "location refresh already in progress"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "location refresh completed"})
}
