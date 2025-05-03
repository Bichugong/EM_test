package api

import (
	"em_test/internal/config"
	"em_test/internal/logger"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"go.uber.org/zap"
)

type AgeResponse struct {
	Age  int    `json:"age"`
	Name string `json:"name"`
}

type GenderResponse struct {
	Gender      string  `json:"gender"`
	Probability float64 `json:"probability"`
	Name        string  `json:"name"`
}

type NationalityResponse struct {
	Country []struct {
		CountryID   string  `json:"country_id"`
		Probability float64 `json:"probability"`
	} `json:"country"`
	Name string `json:"name"`
}

type EnrichmentAPI struct {
	AgeURL         string
	GenderURL      string
	NationalityURL string
}

func NewEnrichmentAPI(cfg config.EnrichmentAPIConfig) *EnrichmentAPI {
	logger.Logger.Debug("Creating new enrichment API client", zap.Any("config", cfg))
	return &EnrichmentAPI{
		AgeURL:         cfg.AgifyURL,
		GenderURL:      cfg.GenderizeURL,
		NationalityURL: cfg.NationalizeURL,
	}
}

func (e *EnrichmentAPI) GetAge(name string) (int, error) {
	logger.Logger.Debug("Calling Agify API", zap.String("name", name))

	url := fmt.Sprintf("%s/?name=%s", e.AgeURL, name)
	resp, err := http.Get(url)
	if err != nil {
		logger.Logger.Error("Agify API call failed", zap.Error(err))
		return 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	var ageResponse AgeResponse
	if err := json.Unmarshal(body, &ageResponse); err != nil {
		return 0, err
	}

	logger.Logger.Info("Age retrieved successfully", zap.String("name", name), zap.Int("age", ageResponse.Age))
	return ageResponse.Age, nil
}

func (e *EnrichmentAPI) GetGender(name string) (string, error) {
	logger.Logger.Debug("Calling Genderize API", zap.String("name", name))
	resp, err := http.Get(fmt.Sprintf("%s/?name=%s", e.GenderURL, name))
	if err != nil {
		logger.Logger.Error("Genderize API call failed", zap.Error(err))
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var genderResponse GenderResponse
	if err := json.Unmarshal(body, &genderResponse); err != nil {
		return "", err
	}

	if genderResponse.Probability > 0.9 {
		logger.Logger.Info("Gender retrieved successfully", zap.String("name", name), zap.String("gender", genderResponse.Gender))
		return genderResponse.Gender, nil
	}

	logger.Logger.Debug("Gender probability too low", zap.String("name", name), zap.Float64("probability", genderResponse.Probability))
	return "", nil
}

func (e *EnrichmentAPI) GetNationality(name string) (string, error) {
	logger.Logger.Debug("Calling Nationalize API", zap.String("name", name))
	resp, err := http.Get(fmt.Sprintf("%s/?name=%s", e.NationalityURL, name))
	if err != nil {
		logger.Logger.Error("Nationalize API call failed", zap.Error(err))
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var nationalityResponse NationalityResponse
	if err := json.Unmarshal(body, &nationalityResponse); err != nil {
		return "", err
	}

	if len(nationalityResponse.Country) == 0 {
		logger.Logger.Debug("No nationality data found", zap.String("name", name))
		return "", nil
	}

	maxProb := nationalityResponse.Country[0]
	for _, country := range nationalityResponse.Country {
		if country.Probability > maxProb.Probability {
			maxProb = country
		}
	}

	logger.Logger.Info("Nationality retrieved successfully", zap.String("name", name), zap.String("country_id", maxProb.CountryID))
	return maxProb.CountryID, nil
}
