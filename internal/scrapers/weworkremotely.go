package scrapers

import (
	"crypto/sha256"
	"encoding/xml"
	"fmt"
	"net/http"
	"strings"
	"time"
	"vagas-scraper/internal/models"
)

type rssItem struct {
	Title   string `xml:"title"`
	Link    string `xml:"link"`
	PubDate string `xml:"pubDate"`
	Region  string `xml:"region"`
}

type rssChannel struct {
	Items []rssItem `xml:"item"`
}

type rssFeed struct {
	Channel rssChannel `xml:"channel"`
}

// WeWorkRemotelyScraper consome o feed RSS aberto do WeWorkRemotely
type WeWorkRemotelyScraper struct{}

func NovoWeWorkRemotelyScraper() *WeWorkRemotelyScraper {
	return &WeWorkRemotelyScraper{}
}

func (w *WeWorkRemotelyScraper) Nome() string {
	return "WeWorkRemotely"
}

func (w *WeWorkRemotelyScraper) Buscar() ([]models.Vaga, error) {
	feedURL := "https://weworkremotely.com/categories/remote-programming-jobs.rss"

	req, err := http.NewRequest("GET", feedURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "VagasScraper/1.0")

	client := &http.Client{Timeout: 12 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("[WWR] falha na requisição: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("[WWR] status HTTP %d", resp.StatusCode)
	}

	var feed rssFeed
	if err := xml.NewDecoder(resp.Body).Decode(&feed); err != nil {
		return nil, fmt.Errorf("[WWR] erro ao decodificar RSS: %w", err)
	}

	var vagas []models.Vaga
	for _, item := range feed.Channel.Items {
		// O formato padrão do título no WWR é: "Nome da Empresa: Cargo da Vaga"
		partes := strings.SplitN(item.Title, ":", 2)
		empresa := "Empresa Remota"
		titulo := item.Title

		if len(partes) == 2 {
			empresa = strings.TrimSpace(partes[0])
			titulo = strings.TrimSpace(partes[1])
		}

		local := item.Region
		if local == "" {
			local = "100% Remoto (Global)"
		}

		id := fmt.Sprintf("wwr-%x", sha256.Sum256([]byte(item.Link)))[:16]

		vagas = append(vagas, models.Vaga{
			ID:          id,
			Titulo:      titulo,
			Empresa:     empresa,
			Localizacao: local,
			Link:        item.Link,
			Fonte:       "WeWorkRemotely",
			Data:        time.Now(),
		})
	}

	return vagas, nil
}
