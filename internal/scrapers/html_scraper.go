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

// HTMLScraperConfig define as regras de raspagem para qualquer página HTML
type HTMLScraperConfig struct {
	NomeFonte       string
	URL             string
	SeletorCard     string
	SeletorTitulo   string
	SeletorEmpresa  string
	SeletorLink     string
	SeletorLocal    string
	BaseURL         string // Para resolver links relativos
}

// GenericHTMLScraper raspa páginas HTML utilizando seletores CSS via goquery
type GenericHTMLScraper struct {
	cfg HTMLScraperConfig
}

func NovoGenericHTMLScraper(cfg HTMLScraperConfig) *GenericHTMLScraper {
	return &GenericHTMLScraper{cfg: cfg}
}

func (s *GenericHTMLScraper) Nome() string {
	return s.cfg.NomeFonte
}

func (s *GenericHTMLScraper) Buscar() ([]models.Vaga, error) {
	req, err := http.NewRequest("GET", s.cfg.URL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("[%s] falha na requisição: %w", s.cfg.NomeFonte, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("[%s] status HTTP inesperado: %d", s.cfg.NomeFonte, resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("[%s] erro ao analisar HTML: %w", s.cfg.NomeFonte, err)
	}

	var vagas []models.Vaga
	doc.Find(s.cfg.SeletorCard).Each(func(i int, card *goquery.Selection) {
		titulo := strings.TrimSpace(card.Find(s.cfg.SeletorTitulo).Text())
		if titulo == "" {
			return
		}

		empresa := "Empresa confidencial"
		if s.cfg.SeletorEmpresa != "" {
			empresa = strings.TrimSpace(card.Find(s.cfg.SeletorEmpresa).Text())
		}

		link := ""
		if s.cfg.SeletorLink != "" {
			link, _ = card.Find(s.cfg.SeletorLink).Attr("href")
		} else {
			link, _ = card.Attr("href")
		}
		link = resolverLink(link, s.cfg.BaseURL)

		local := "Brasil / Remoto"
		if s.cfg.SeletorLocal != "" {
			l := strings.TrimSpace(card.Find(s.cfg.SeletorLocal).Text())
			if l != "" {
				local = l
			}
		}

		// Gera um ID consistente baseado na URL ou título
		idHash := fmt.Sprintf("%x", sha256.Sum256([]byte(s.cfg.NomeFonte+":"+link+":"+titulo)))[:16]

		vagas = append(vagas, models.Vaga{
			ID:          idHash,
			Titulo:      titulo,
			Empresa:     empresa,
			Localizacao: local,
			Link:        link,
			Fonte:       s.cfg.NomeFonte,
			Data:        time.Now(),
		})
	})

	return vagas, nil
}

func resolverLink(href, baseURL string) string {
	href = strings.TrimSpace(href)
	if href == "" {
		return baseURL
	}
	if strings.HasPrefix(href, "http://") || strings.HasPrefix(href, "https://") {
		return href
	}
	baseURL = strings.TrimRight(baseURL, "/")
	href = strings.TrimLeft(href, "/")
	return baseURL + "/" + href
}
