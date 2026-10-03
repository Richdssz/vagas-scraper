package scrapers

import (
	"crypto/sha256"
	"fmt"
	"net/http"
	"strings"
	"time"
	"vagas-scraper/internal/models"

	"github.com/PuerkitoBio/goquery"
)

// ProgramaThorScraper busca vagas no portal ProgramaThor
type ProgramaThorScraper struct{}

func NovoProgramaThorScraper() *ProgramaThorScraper {
	return &ProgramaThorScraper{}
}

func (p *ProgramaThorScraper) Nome() string {
	return "ProgramaThor"
}

func (p *ProgramaThorScraper) Buscar() ([]models.Vaga, error) {
	urlSite := "https://programathor.com.br/jobs"

	req, err := http.NewRequest("GET", urlSite, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	client := &http.Client{Timeout: 12 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("[ProgramaThor] falha na requisição: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("[ProgramaThor] status HTTP %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("[ProgramaThor] falha ao ler HTML: %w", err)
	}

	var vagas []models.Vaga
	doc.Find(".cell-list, div[class*='cell-list']").Each(func(i int, s *goquery.Selection) {
		link, _ := s.Find("a[href*='/jobs/']").Attr("href")
		if link == "" {
			link, _ = s.Attr("href")
		}
		if link == "" {
			return
		}

		if !strings.HasPrefix(link, "http") {
			link = "https://programathor.com.br" + link
		}

		titulo := strings.TrimSpace(s.Find("h3, .cell-list-content-title").First().Text())
		if titulo == "" {
			return
		}

		empresa := strings.TrimSpace(s.Find("span[class*='company'], .cell-list-content-subtitle").First().Text())
		if empresa == "" {
			empresa = "Empresa no ProgramaThor"
		}

		local := "Remoto / Brasil"
		if l := strings.TrimSpace(s.Find(".cell-list-content-icon, span:contains('Remoto')").First().Text()); l != "" {
			local = l
		}

		var tags []string
		s.Find(".tag-list span, .badge").Each(func(j int, t *goquery.Selection) {
			tagText := strings.TrimSpace(t.Text())
			if tagText != "" {
				tags = append(tags, tagText)
			}
		})

		id := fmt.Sprintf("pgt-%x", sha256.Sum256([]byte(link)))[:16]

		vagas = append(vagas, models.Vaga{
			ID:          id,
			Titulo:      titulo,
			Empresa:     empresa,
			Localizacao: local,
			Link:        link,
			Fonte:       "ProgramaThor",
			Tags:        tags,
			Data:        time.Now(),
		})
	})

	return vagas, nil
}
