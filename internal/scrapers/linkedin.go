package scrapers

import (
	"crypto/sha256"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
	"vagas-scraper/internal/models"

	"github.com/PuerkitoBio/goquery"
)

// LinkedInScraper consulta a API pública para visitantes do LinkedIn
type LinkedInScraper struct {
	termoBusca string
}

func NovoLinkedInScraper(termoBusca string) *LinkedInScraper {
	if termoBusca == "" {
		termoBusca = "desenvolvedor"
	}
	return &LinkedInScraper{termoBusca: termoBusca}
}

func (l *LinkedInScraper) Nome() string {
	return "LinkedIn"
}

func (l *LinkedInScraper) Buscar() ([]models.Vaga, error) {
	termoEscapado := url.QueryEscape(l.termoBusca)
	apiURL := fmt.Sprintf("https://www.linkedin.com/jobs-guest/jobs/api/seeMoreJobPostings/search?keywords=%s&location=Brasil&sortBy=DD", termoEscapado)

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, err
	}

	// Cabeçalho de navegador comum para simular visitante
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36")
	req.Header.Set("Accept-Language", "pt-BR,pt;q=0.9,en-US;q=0.8,en;q=0.7")

	client := &http.Client{Timeout: 12 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("[LinkedIn] falha na requisição: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("[LinkedIn] status HTTP %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("[LinkedIn] erro ao ler HTML: %w", err)
	}

	var vagas []models.Vaga
	doc.Find("li, .base-search-card").Each(func(i int, s *goquery.Selection) {
		titulo := strings.TrimSpace(s.Find(".base-search-card__title, h3").First().Text())
		empresa := strings.TrimSpace(s.Find(".base-search-card__subtitle, h4").First().Text())
		local := strings.TrimSpace(s.Find(".job-search-card__location").First().Text())
		link, _ := s.Find("a.base-card__full-link").Attr("href")

		if titulo == "" || link == "" {
			return
		}

		// Limpa parâmetros de rastreamento do link do LinkedIn
		if idx := strings.Index(link, "?"); idx != -1 {
			link = link[:idx]
		}

		if local == "" {
			local = "Brasil"
		}
		if empresa == "" {
			empresa = "Empresa no LinkedIn"
		}

		dataVaga := time.Now()
		if timeEl := s.Find("time"); timeEl.Length() > 0 {
			if dt, ok := timeEl.Attr("datetime"); ok && dt != "" {
				if t, err := time.Parse("2006-01-02", dt); err == nil {
					dataVaga = t
				}
			}
		}

		id := fmt.Sprintf("in-%x", sha256.Sum256([]byte(link)))[:16]

		vagas = append(vagas, models.Vaga{
			ID:          id,
			Titulo:      titulo,
			Empresa:     empresa,
			Localizacao: local,
			Link:        link,
			Fonte:       "LinkedIn",
			Data:        dataVaga,
		})
	})

	return vagas, nil
}
