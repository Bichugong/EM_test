package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
	"em_test/pkg/logging"
)

// EnricherService обогащает данные о человеке из внешних API
type EnricherService interface {
	Enrich(ctx context.Context, name string) (age int, gender string, nationality string, err error)
}

type enricher struct {
	client    *http.Client
	logger    logging.Logger
	timeout   time.Duration
}

func NewEnricherService(logger logging.Logger) EnricherService {
	return &enricher{
		client:  &http.Client{Timeout: 5 * time.Second},
		logger:  logger,
		timeout: 3 * time.Second,
	}
}

// Enrich получает данные из трех API параллельно
func (e *enricher) Enrich(ctx context.Context, name string) (int, string, string, error) {
	ctx, cancel := context.WithTimeout(ctx, e.timeout)
	defer cancel()

	type result struct {
		age         int
		gender      string
		nationality string
		err         error
	}

	ch := make(chan result, 3)

	// Запускаем все запросы параллельно
	go func() {
		age, err := e.getAge(ctx, name)
		ch <- result{age: age, err: err}
	}()

	go func() {
		gender, err := e.getGender(ctx, name)
		ch <- result{gender: gender, err: err}
	}()

	go func() {
		nationality, err := e.getNationality(ctx, name)
		ch <- result{nationality: nationality, err: err}
	}()

	var (
		age int
		gender, nationality string
		count                    int
		finalErr                 error
	)

	// Собираем результаты
	for i := 0; i < 3; i++ {
		select {
		case <-ctx.Done():
			return 0, "", "", fmt.Errorf("enrichment timeout exceeded")
		case res := <-ch:
			if res.err != nil {
				finalErr = res.err
				continue
			}
			if res.age > 0 {
				age = res.age
				count++
			}
			if res.gender != "" {
				gender = res.gender
				count++
			}
			if res.nationality != "" {
				nationality = res.nationality
				count++
			}
		}
	}

	if count == 0 && finalErr != nil {
		return 0, "", "", fmt.Errorf("all enrichment APIs failed: %v", finalErr)
	}

	return age, gender, nationality, nil
}

func (e *enricher) getAge(ctx context.Context, name string) (int, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		"GET", 
		fmt.Sprintf("https://api.agify.io/?name=%s", name),
		nil,
	)
	if err != nil {
		return 0, fmt.Errorf("create age request failed: %w", err)
	}

	resp, err := e.client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("age API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("age API returned status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("read age response failed: %w", err)
	}

	var data struct {
		Age  int    `json:"age"`
		Name string `json:"name"`
	}
	if err := json.Unmarshal(body, &data); err != nil {
		return 0, fmt.Errorf("parse age response failed: %w", err)
	}

	e.logger.Debugf("Received age %d for name %s", data.Age, name)
	return data.Age, nil
}

func (e *enricher) getGender(ctx context.Context, name string) (string, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		"GET",
		fmt.Sprintf("https://api.genderize.io/?name=%s", name),
		nil,
	)
	if err != nil {
		return "", fmt.Errorf("create gender request failed: %w", err)
	}

	resp, err := e.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("gender API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("gender API returned status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read gender response failed: %w", err)
	}

	var data struct {
		Gender string `json:"gender"`
	}
	if err := json.Unmarshal(body, &data); err != nil {
		return "", fmt.Errorf("parse gender response failed: %w", err)
	}

	e.logger.Debugf("Received gender %s for name %s", data.Gender, name)
	return strings.ToLower(data.Gender), nil
}

func (e *enricher) getNationality(ctx context.Context, name string) (string, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		"GET",
		fmt.Sprintf("https://api.nationalize.io/?name=%s", name),
		nil,
	)
	if err != nil {
		return "", fmt.Errorf("create nationality request failed: %w", err)
	}

	resp, err := e.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("nationality API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("nationality API returned status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read nationality response failed: %w", err)
	}

	var data struct {
		Country []struct {
			CountryID string `json:"country_id"`
		} `json:"country"`
	}
	if err := json.Unmarshal(body, &data); err != nil {
		return "", fmt.Errorf("parse nationality response failed: %w", err)
	}

	if len(data.Country) == 0 {
		return "", nil
	}

	nationality := strings.ToLower(data.Country[0].CountryID)
	e.logger.Debugf("Received nationality %s for name %s", nationality, name)
	return nationality, nil
}