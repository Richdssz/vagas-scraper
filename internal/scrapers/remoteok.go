package scrapers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
	"vagas-scraper/internal/models"
)

type remoteOKItem struct {
	ID       any      `json:"id"`
	Position string   `json:"position"`
	Company  string   `json:"company"`
	Location string   `json:"location"`
	URL      string   `json:"url"`
	ApplyURL string   `json:"apply_url"`
	Tags     []string `json:"tags"`
	Date     string   `json:"date"`
}

// RemoteOKScraper busca vagas remotas no feed público do RemoteOK
type RemoteOKScraper struct{}

func NovoRemoteOKScraper() *RemoteOKScraper {
	return &RemoteOKScraper{}
}

func (r *RemoteOKScraper) Nome() string {
	return "RemoteOK"
}

func (r *RemoteOKScraper) Buscar() ([]models.Vaga, error) {
	req, err := http.NewRequest("GET", "https://remoteok.com/api", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "VagasScraper-Bot/1.0 (Mozilla/5.0)")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("[RemoteOK] erro na requisição: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("[RemoteOK] HTTP status %d", resp.StatusCode)
	}

	var items []remoteOKItem
	if err := json.NewDecoder(resp.Body).Decode(&items); err != nil {
		return nil, fmt.Errorf("[RemoteOK] erro ao decodificar JSON: %w", err)
	}

	var vagas []models.Vaga
	for _, it := range items {
		// O primeiro item do feed do RemoteOK costuma ser um aviso legal sem position
		if it.Position == "" || it.Company == "" {
			continue
		}

		link := it.ApplyURL
		if link == "" {
			link = it.URL
		}
		if link == "" {
			continue
		}

		idStr := fmt.Sprintf("rok-%v", it.ID)
		local := it.Location
		if local == "" {
			local = "100% Remoto (Global)"
		}

		dataVaga := time.Now()
		if it.Date != "" {
			if parsed, err := time.Parse(time.RFC3339, it.Date); err == nil {
				dataVaga = parsed
			}
		}

		vagas = append(vagas, models.Vaga{
			ID:          idStr,
			Titulo:      it.Position,
			Empresa:     it.Company,
			Localizacao: local,
			Link:        link,
			Fonte:       "RemoteOK",
			Tags:        it.Tags,
			Data:        dataVaga,
		})
	}

	return vagas, nil
}
