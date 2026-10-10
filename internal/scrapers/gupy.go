package scrapers

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
	"vagas-scraper/internal/models"

	"github.com/PuerkitoBio/goquery"
)

// GupyScraper extrai oportunidades do portal público da Gupy
type GupyScraper struct {
	termosBusca []string
}

func NovoGupyScraper(termos ...string) *GupyScraper {
	var limpos []string
	for _, t := range termos {
		t = strings.TrimSpace(t)
		if t != "" {
			limpos = append(limpos, t)
		}
	}
	if len(limpos) == 0 {
		limpos = []string{"desenvolvedor"}
	}
	return &GupyScraper{termosBusca: limpos}
}

func (g *GupyScraper) Nome() string {
	return "Gupy"
}

func (g *GupyScraper) Buscar() ([]models.Vaga, error) {
	var todasVagas []models.Vaga
	vistos := make(map[string]bool)

	for _, termo := range g.termosBusca {
		vagasTermo, err := g.buscarPorTermo(termo)
		if err != nil {
			continue
		}
		for _, v := range vagasTermo {
			if !vistos[v.ID] && !vistos[v.Link] {
				vistos[v.ID] = true
				vistos[v.Link] = true
				todasVagas = append(todasVagas, v)
			}
		}
	}
	return todasVagas, nil
}

func (g *GupyScraper) buscarPorTermo(termoBusca string) ([]models.Vaga, error) {
	termo := url.QueryEscape(termoBusca)
	urlBusca := fmt.Sprintf("https://portal.gupy.io/job-search/term=%s", termo)

	req, err := http.NewRequest("GET", urlBusca, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36")
	req.Header.Set("Accept-Language", "pt-BR,pt;q=0.9")

	client := &http.Client{Timeout: 12 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("[Gupy] falha na requisição: %w", err)
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("[Gupy] falha ao processar HTML: %w", err)
	}

	var vagas []models.Vaga

	// Método 1: Tenta ler o JSON estruturado do Next.js (__NEXT_DATA__)
	nextDataScript := doc.Find("#__NEXT_DATA__").Text()
	if nextDataScript != "" {
		vagas = extrairVagasNextData(nextDataScript)
	}

	// Método 2: Se não vier pelo JSON do Next.js, faz varredura por links e cards no HTML
	if len(vagas) == 0 {
		doc.Find("a[href*='gupy.io/job/'], div[data-testid='job-card-wrapper']").Each(func(i int, s *goquery.Selection) {
			link, _ := s.Attr("href")
			if link == "" {
				link, _ = s.Find("a").Attr("href")
			}
			if link == "" {
				return
			}

			titulo := strings.TrimSpace(s.Find("h3, h2, [class*='title']").First().Text())
			if titulo == "" {
				titulo = strings.TrimSpace(s.Text())
			}
			if len(titulo) > 80 {
				titulo = titulo[:80] + "..."
			}

			empresa := strings.TrimSpace(s.Find("[class*='company'], [class*='careerPage']").First().Text())
			if empresa == "" {
				empresa = "Empresa via Gupy"
			}

			id := fmt.Sprintf("gupy-%x", sha256.Sum256([]byte(link)))[:16]

			vagas = append(vagas, models.Vaga{
				ID:          id,
				Titulo:      titulo,
				Empresa:     empresa,
				Localizacao: "Brasil / Gupy",
				Link:        link,
				Fonte:       "Gupy",
				Data:        time.Now(),
			})
		})
	}

	return vagas, nil
}

type GupyNextData struct {
	Props struct {
		PageProps struct {
			InitialJobList struct {
				Data []struct {
					ID             int    `json:"id"`
					Name           string `json:"name"`
					CareerPageName string `json:"careerPageName"`
					WorkplaceType  string `json:"workplaceType"`
					City           string `json:"city"`
					State          string `json:"state"`
					JobURL         string `json:"jobUrl"`
					PublishedDate  string `json:"publishedDate"`
				} `json:"data"`
			} `json:"initialJobList"`
		} `json:"pageProps"`
	} `json:"props"`
}

func extrairVagasNextData(rawJSON string) []models.Vaga {
	var gupyData GupyNextData
	if err := json.Unmarshal([]byte(rawJSON), &gupyData); err == nil && len(gupyData.Props.PageProps.InitialJobList.Data) > 0 {
		var vagas []models.Vaga
		for _, item := range gupyData.Props.PageProps.InitialJobList.Data {
			if item.Name == "" {
				continue
			}

			link := item.JobURL
			if link == "" {
				link = fmt.Sprintf("https://portal.gupy.io/jobs/%d", item.ID)
			}

			local := "Brasil"
			if item.City != "" || item.State != "" {
				local = strings.TrimSpace(fmt.Sprintf("%s - %s", item.City, item.State))
			}
			if item.WorkplaceType != "" {
				if local != "Brasil" {
					local = fmt.Sprintf("%s (%s)", local, item.WorkplaceType)
				} else {
					local = item.WorkplaceType
				}
			}

			empresa := item.CareerPageName
			if empresa == "" {
				empresa = "Empresa via Gupy"
			}

			id := fmt.Sprintf("gupy-%d", item.ID)
			vagas = append(vagas, models.Vaga{
				ID:          id,
				Titulo:      strings.TrimSpace(item.Name),
				Empresa:     strings.TrimSpace(empresa),
				Localizacao: local,
				Link:        link,
				Fonte:       "Gupy",
				Data:        time.Now(),
			})
		}
		return vagas
	}

	// Fallback por regex caso o schema mude
	re := regexp.MustCompile(`"id":\s*(\d+)[^,}]*?,\s*"name":\s*"([^"]+)"`)
	matches := re.FindAllStringSubmatch(rawJSON, 30)

	var vagas []models.Vaga
	for _, m := range matches {
		if len(m) >= 3 {
			jobID := m[1]
			nome := m[2]
			link := fmt.Sprintf("https://portal.gupy.io/jobs/%s", jobID)
			id := fmt.Sprintf("gupy-%s", jobID)

			vagas = append(vagas, models.Vaga{
				ID:          id,
				Titulo:      strings.TrimSpace(nome),
				Empresa:     "Empresa via Gupy",
				Localizacao: "Brasil / Gupy",
				Link:        link,
				Fonte:       "Gupy",
				Data:        time.Now(),
			})
		}
	}

	return vagas
}
