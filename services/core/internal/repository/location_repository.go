package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"unicode"

	"github.com/animalpride/partners/services/core/internal/models"
	"gorm.io/gorm"
)

const (
	defaultCombinedURL  = "https://raw.githubusercontent.com/dr5hn/countries-states-cities-database/master/json/countries+states+cities.json"
	defaultCountriesURL = "https://raw.githubusercontent.com/dr5hn/countries-states-cities-database/master/json/countries.json"
	defaultStatesURL    = "https://raw.githubusercontent.com/dr5hn/countries-states-cities-database/master/json/states.json"
	defaultCitiesURL    = "https://raw.githubusercontent.com/dr5hn/countries-states-cities-database/master/json/cities.json"
	importLockName      = "partners:location_import"
)

type LocationRepository struct {
	db *gorm.DB
}

type CountryOption struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type CityStateOption struct {
	CityLookupID uint   `json:"city_lookup_id"`
	City         string `json:"city"`
	State        string `json:"state"`
	StateCode    string `json:"state_code"`
	Country      string `json:"country"`
	CountryCode  string `json:"country_code"`
	Label        string `json:"label"`
}

func NewLocationRepository(db *gorm.DB) *LocationRepository {
	return &LocationRepository{db: db}
}

func normalizeSearchTerm(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		return ""
	}
	var b strings.Builder
	lastSpace := false
	for _, r := range value {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			b.WriteRune(r)
			lastSpace = false
			continue
		}
		if !lastSpace {
			b.WriteRune(' ')
			lastSpace = true
		}
	}
	return strings.TrimSpace(b.String())
}

func fallbackStateCode(primaryCode, stateName string) string {
	code := strings.ToUpper(strings.TrimSpace(primaryCode))
	if code != "" {
		return code
	}
	normalized := normalizeSearchTerm(stateName)
	normalized = strings.ReplaceAll(normalized, " ", "_")
	if len(normalized) > 20 {
		normalized = normalized[:20]
	}
	return strings.ToUpper(normalized)
}

func (r *LocationRepository) ListCountries() ([]CountryOption, error) {
	var rows []models.LocationCountry
	if err := r.db.Order("name asc").Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]CountryOption, 0, len(rows))
	for _, row := range rows {
		out = append(out, CountryOption{Code: row.Code, Name: row.Name})
	}
	return out, nil
}

func (r *LocationRepository) SearchCityStates(countryCode, query string, limit int) ([]CityStateOption, error) {
	countryCode = strings.ToUpper(strings.TrimSpace(countryCode))
	query = normalizeSearchTerm(query)
	if countryCode == "" || query == "" {
		return []CityStateOption{}, nil
	}
	if limit <= 0 || limit > 50 {
		limit = 25
	}
	like := query + "%"

	var rows []struct {
		ID          uint
		CityName    string
		StateName   string
		StateCode   string
		CountryName string
		CountryCode string
	}

	err := r.db.Table("location_cities lc").
		Select("lc.id as id, lc.name as city_name, lc.state_name as state_name, lc.state_code as state_code, c.name as country_name, lc.country_code as country_code").
		Joins("JOIN location_countries c ON c.code = lc.country_code").
		Where("lc.country_code = ?", countryCode).
		Where("lc.search_name LIKE ? OR EXISTS (SELECT 1 FROM location_city_aliases a WHERE a.city_id = lc.id AND a.search_alias LIKE ?)", like, like).
		Order("lc.search_name asc").
		Limit(limit).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}

	results := make([]CityStateOption, 0, len(rows))
	for _, row := range rows {
		stateLabel := row.StateCode
		if strings.TrimSpace(stateLabel) == "" {
			stateLabel = row.StateName
		}
		results = append(results, CityStateOption{
			CityLookupID: row.ID,
			City:         row.CityName,
			State:        row.StateName,
			StateCode:    row.StateCode,
			Country:      row.CountryName,
			CountryCode:  row.CountryCode,
			Label:        fmt.Sprintf("%s, %s", row.CityName, stateLabel),
		})
	}
	return results, nil
}

func (r *LocationRepository) FindCountryByCodeOrName(codeOrName string) (*models.LocationCountry, error) {
	needle := strings.TrimSpace(codeOrName)
	if needle == "" {
		return nil, nil
	}
	var country models.LocationCountry
	if len(needle) == 2 {
		err := r.db.Where("code = ?", strings.ToUpper(needle)).First(&country).Error
		if err == nil {
			return &country, nil
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
	}

	normalized := normalizeSearchTerm(needle)
	err := r.db.Where("search_name = ?", normalized).First(&country).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &country, nil
}

func (r *LocationRepository) FindCityByID(countryCode string, cityID uint) (*models.LocationCity, error) {
	if cityID == 0 {
		return nil, nil
	}
	var city models.LocationCity
	err := r.db.Where("id = ? AND country_code = ?", cityID, strings.ToUpper(strings.TrimSpace(countryCode))).First(&city).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &city, nil
}

func (r *LocationRepository) FindCityByName(countryCode, cityName, state string) (*models.LocationCity, error) {
	countryCode = strings.ToUpper(strings.TrimSpace(countryCode))
	cityNeedle := normalizeSearchTerm(cityName)
	stateNeedle := normalizeSearchTerm(state)
	if countryCode == "" || cityNeedle == "" || stateNeedle == "" {
		return nil, nil
	}

	var city models.LocationCity
	err := r.db.Where("country_code = ? AND search_name = ? AND (LOWER(state_code) = ? OR LOWER(state_name) = ?)",
		countryCode,
		cityNeedle,
		strings.ToLower(strings.TrimSpace(state)),
		stateNeedle,
	).First(&city).Error
	if err == nil {
		return &city, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	return nil, nil
}

type rawCountry struct {
	Name string `json:"name"`
	ISO2 string `json:"iso2"`
}

type rawState struct {
	Name        string `json:"name"`
	CountryCode string `json:"country_code"`
	StateCode   string `json:"state_code"`
}

type rawCity struct {
	Name        string `json:"name"`
	CountryCode string `json:"country_code"`
	StateCode   string `json:"state_code"`
	StateName   string `json:"state_name"`
}

type rawCombinedCity struct {
	Name string `json:"name"`
}

type rawCombinedState struct {
	Name      string            `json:"name"`
	StateCode string            `json:"state_code"`
	Cities    []rawCombinedCity `json:"cities"`
}

type rawCombinedCountry struct {
	Name   string             `json:"name"`
	ISO2   string             `json:"iso2"`
	States []rawCombinedState `json:"states"`
}

func buildRowsFromCombined(data []rawCombinedCountry) ([]models.LocationCountry, []models.LocationState, []models.LocationCity) {
	countryRows := make([]models.LocationCountry, 0, len(data))
	stateRows := make([]models.LocationState, 0, len(data)*8)
	cityRows := make([]models.LocationCity, 0, len(data)*300)
	countrySeen := map[string]struct{}{}
	stateSeen := map[string]struct{}{}

	for _, country := range data {
		countryCode := strings.ToUpper(strings.TrimSpace(country.ISO2))
		countryName := strings.TrimSpace(country.Name)
		if countryCode == "" || countryName == "" {
			continue
		}

		if _, exists := countrySeen[countryCode]; !exists {
			countrySeen[countryCode] = struct{}{}
			countryRows = append(countryRows, models.LocationCountry{
				Code:       countryCode,
				Name:       countryName,
				SearchName: normalizeSearchTerm(countryName),
			})
		}

		for _, state := range country.States {
			stateName := strings.TrimSpace(state.Name)
			stateCode := fallbackStateCode(state.StateCode, stateName)
			if stateCode == "" || stateName == "" {
				continue
			}

			stateKey := countryCode + ":" + stateCode
			if _, exists := stateSeen[stateKey]; !exists {
				stateSeen[stateKey] = struct{}{}
				stateRows = append(stateRows, models.LocationState{
					CountryCode: countryCode,
					StateCode:   stateCode,
					Name:        stateName,
					SearchName:  normalizeSearchTerm(stateName),
				})
			}

			for _, city := range state.Cities {
				cityName := strings.TrimSpace(city.Name)
				if cityName == "" {
					continue
				}

				cityRows = append(cityRows, models.LocationCity{
					CountryCode: countryCode,
					StateCode:   stateCode,
					StateName:   stateName,
					Name:        cityName,
					SearchName:  normalizeSearchTerm(cityName),
				})
			}
		}
	}

	return countryRows, stateRows, cityRows
}

func buildRowsFromFlat(countries []rawCountry, states []rawState, cities []rawCity) ([]models.LocationCountry, []models.LocationState, []models.LocationCity) {
	countryRows := make([]models.LocationCountry, 0, len(countries))
	countrySeen := map[string]struct{}{}
	for _, country := range countries {
		code := strings.ToUpper(strings.TrimSpace(country.ISO2))
		name := strings.TrimSpace(country.Name)
		if code == "" || name == "" {
			continue
		}
		if _, exists := countrySeen[code]; exists {
			continue
		}
		countrySeen[code] = struct{}{}
		countryRows = append(countryRows, models.LocationCountry{
			Code:       code,
			Name:       name,
			SearchName: normalizeSearchTerm(name),
		})
	}

	stateRows := make([]models.LocationState, 0, len(states))
	stateSeen := map[string]struct{}{}
	for _, state := range states {
		countryCode := strings.ToUpper(strings.TrimSpace(state.CountryCode))
		name := strings.TrimSpace(state.Name)
		stateCode := fallbackStateCode(state.StateCode, name)
		if countryCode == "" || stateCode == "" || name == "" {
			continue
		}
		stateKey := countryCode + ":" + stateCode
		if _, exists := stateSeen[stateKey]; exists {
			continue
		}
		stateSeen[stateKey] = struct{}{}
		stateRows = append(stateRows, models.LocationState{
			CountryCode: countryCode,
			StateCode:   stateCode,
			Name:        name,
			SearchName:  normalizeSearchTerm(name),
		})
	}

	cityRows := make([]models.LocationCity, 0, len(cities))
	for _, city := range cities {
		countryCode := strings.ToUpper(strings.TrimSpace(city.CountryCode))
		stateName := strings.TrimSpace(city.StateName)
		stateCode := fallbackStateCode(city.StateCode, stateName)
		name := strings.TrimSpace(city.Name)
		if countryCode == "" || stateCode == "" || stateName == "" || name == "" {
			continue
		}
		cityRows = append(cityRows, models.LocationCity{
			CountryCode: countryCode,
			StateCode:   stateCode,
			StateName:   stateName,
			Name:        name,
			SearchName:  normalizeSearchTerm(name),
		})
	}

	return countryRows, stateRows, cityRows
}

func fetchJSON[T any](ctx context.Context, url string) ([]T, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("failed to fetch %s: status %d", url, resp.StatusCode)
	}

	var out []T
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *LocationRepository) acquireImportLock(ctx context.Context, waitSeconds int) (bool, error) {
	if waitSeconds < 0 {
		waitSeconds = 0
	}
	var acquired int
	err := r.db.WithContext(ctx).Raw("SELECT GET_LOCK(?, ?)", importLockName, waitSeconds).Scan(&acquired).Error
	if err != nil {
		return false, err
	}
	return acquired == 1, nil
}

func (r *LocationRepository) releaseImportLock(ctx context.Context) {
	var released int
	_ = r.db.WithContext(ctx).Raw("SELECT RELEASE_LOCK(?)", importLockName).Scan(&released).Error
}

func (r *LocationRepository) RefreshFromGitHub(ctx context.Context, combinedURL, countriesURL, statesURL, citiesURL string, waitSeconds int) (bool, error) {
	acquired, err := r.acquireImportLock(ctx, waitSeconds)
	if err != nil {
		return false, err
	}
	if !acquired {
		return false, nil
	}
	defer r.releaseImportLock(ctx)

	if err := r.ImportFromGitHub(ctx, combinedURL, countriesURL, statesURL, citiesURL); err != nil {
		return false, err
	}
	return true, nil
}

func (r *LocationRepository) ImportFromGitHub(ctx context.Context, combinedURL, countriesURL, statesURL, citiesURL string) error {
	if strings.TrimSpace(combinedURL) == "" {
		combinedURL = defaultCombinedURL
	}

	combined, combinedErr := fetchJSON[rawCombinedCountry](ctx, combinedURL)

	var countryRows []models.LocationCountry
	var stateRows []models.LocationState
	var cityRows []models.LocationCity

	if combinedErr == nil && len(combined) > 0 {
		countryRows, stateRows, cityRows = buildRowsFromCombined(combined)
	} else {
		if strings.TrimSpace(countriesURL) == "" {
			countriesURL = defaultCountriesURL
		}
		if strings.TrimSpace(statesURL) == "" {
			statesURL = defaultStatesURL
		}
		if strings.TrimSpace(citiesURL) == "" {
			citiesURL = defaultCitiesURL
		}

		countries, err := fetchJSON[rawCountry](ctx, countriesURL)
		if err != nil {
			return err
		}
		states, err := fetchJSON[rawState](ctx, statesURL)
		if err != nil {
			return err
		}
		cities, err := fetchJSON[rawCity](ctx, citiesURL)
		if err != nil {
			return err
		}

		countryRows, stateRows, cityRows = buildRowsFromFlat(countries, states, cities)
	}

	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("DELETE FROM location_city_aliases").Error; err != nil {
			return err
		}
		if err := tx.Exec("DELETE FROM location_cities").Error; err != nil {
			return err
		}
		if err := tx.Exec("DELETE FROM location_states").Error; err != nil {
			return err
		}
		if err := tx.Exec("DELETE FROM location_countries").Error; err != nil {
			return err
		}

		if len(countryRows) > 0 {
			if err := tx.CreateInBatches(countryRows, 500).Error; err != nil {
				return err
			}
		}
		if len(stateRows) > 0 {
			if err := tx.CreateInBatches(stateRows, 1000).Error; err != nil {
				return err
			}
		}
		if len(cityRows) > 0 {
			if err := tx.CreateInBatches(cityRows, 2000).Error; err != nil {
				return err
			}
		}

		if err := tx.Exec("INSERT INTO location_city_aliases (city_id, alias, search_alias) SELECT id, name, search_name FROM location_cities").Error; err != nil {
			return err
		}
		if err := tx.Exec("INSERT IGNORE INTO location_city_aliases (city_id, alias, search_alias) SELECT id, name, REPLACE(search_name, 'saint ', 'st ') FROM location_cities WHERE search_name LIKE 'saint %'").Error; err != nil {
			return err
		}
		if err := tx.Exec("INSERT IGNORE INTO location_city_aliases (city_id, alias, search_alias) SELECT id, name, REPLACE(search_name, 'st ', 'saint ') FROM location_cities WHERE search_name LIKE 'st %'").Error; err != nil {
			return err
		}

		return nil
	})
}
