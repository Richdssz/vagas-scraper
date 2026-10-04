package scrapers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
	"vagas-scraper/internal/models"
)

type githubLabel struct {
	Name string `json:"name"`
}

type githubIssue struct {
	ID        int64         `json:"id"`
	Title     string        `json:"title"`
	HTMLURL   string        `json:"html_url"`
	CreatedAt time.Time     `json:"created_at"`
	Labels    []githubLabel `json:"labels"`
}

// GitHubVagasScraper consulta repositórios de vagas da comunidade brasileira no GitHub
type GitHubVagasScraper struct {
	nomeRepositorio string // ex: "backend-br/vagas" ou "frontend-br/vagas"
	alias           string // ex: "Backend-BR" ou "Frontend-BR"
}

// NovoGitHubScraper instancia o rastreador para um repositório específico
func NovoGitHubScraper(repo, alias string) *GitHubVagasScraper {
	return &GitHubVagasScraper{
		nomeRepositorio: repo,
		alias:           alias,
	}
}

func (g *GitHubVagasScraper) Nome() string {
	return g.alias
}

func (g *GitHubVagasScraper) Buscar() ([]models.Vaga, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/issues?state=open&per_page=40", g.nomeRepositorio)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// GitHub API requer User-Agent
	req.Header.Set("User-Agent", "VagasScraper-Bot/1.0")
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	client := &http.Client{Timeout: 12 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("[%s] falha na requisição: %w", g.alias, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("[%s] API retornou status HTTP %d", g.alias, resp.StatusCode)
	}

	var issues []githubIssue
	if err := json.NewDecoder(resp.Body).Decode(&issues); err != nil {
		return nil, fmt.Errorf("[%s] falha ao decodificar JSON: %w", g.alias, err)
	}

	var vagas []models.Vaga
	for _, issue := range issues {
		// Ignora issues que não são anúncios de vagas (ex: discussões ou regras)
		if strings.HasPrefix(strings.ToLower(issue.Title), "aviso") || strings.HasPrefix(strings.ToLower(issue.Title), "regra") {
			continue
		}

		tags := make([]string, 0, len(issue.Labels))
		localizacao := "Não especificada"
		for _, l := range issue.Labels {
			tags = append(tags, l.Name)
			if strings.Contains(strings.ToLower(l.Name), "remoto") {
				localizacao = "Remoto"
			}
		}

		// Tenta extrair empresa ou localização do formato comum de título: [Localização/Remoto] Cargo na Empresa
		if localizacao == "Não especificada" {
			if ini := strings.Index(issue.Title, "["); ini != -1 {
				if fim := strings.Index(issue.Title, "]"); fim > ini {
					conteudo := strings.TrimSpace(issue.Title[ini+1 : fim])
					if conteudo != "" && len(conteudo) <= 40 {
						localizacao = conteudo
					}
				}
			}
		}

		empresa := extrairEmpresa(issue.Title)

		vagas = append(vagas, models.Vaga{
			ID:          fmt.Sprintf("gh-%s-%d", strings.ReplaceAll(g.nomeRepositorio, "/", "-"), issue.ID),
			Titulo:      limparTitulo(issue.Title),
			Empresa:     empresa,
			Localizacao: localizacao,
			Link:        issue.HTMLURL,
			Fonte:       g.alias,
			Tags:        tags,
			Data:        issue.CreatedAt,
		})
	}

	return vagas, nil
}

func extrairEmpresa(titulo string) string {
	partes := strings.Split(titulo, " na ")
	if len(partes) > 1 {
		return strings.TrimSpace(partes[len(partes)-1])
	}
	partesEm := strings.Split(titulo, " em ")
	if len(partesEm) > 1 {
		return strings.TrimSpace(partesEm[len(partesEm)-1])
	}
	partesAt := strings.Split(titulo, " @ ")
	if len(partesAt) > 1 {
		return strings.TrimSpace(partesAt[len(partesAt)-1])
	}
	return "Ver no anúncio"
}

func limparTitulo(t string) string {
	return strings.TrimSpace(t)
}
