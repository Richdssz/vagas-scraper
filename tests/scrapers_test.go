package scrapers_test

import (
	"fmt"
	"testing"
	"time"
	"vagas-scraper/internal/scrapers"
)

func TestProvedoresDeVagas(t *testing.T) {
	provedores := []struct {
		nome    string
		scraper scrapers.Scraper
	}{
		{
			nome:    "LinkedIn (Geral)",
			scraper: scrapers.NovoLinkedInScraper("desenvolvedor"),
		},
		{
			nome:    "Gupy (Geral)",
			scraper: scrapers.NovoGupyScraper("desenvolvedor"),
		},
		{
			nome:    "GitHub Backend-BR",
			scraper: scrapers.NovoGitHubScraper("backend-br/vagas", "Backend-BR"),
		},
		{
			nome:    "GitHub Frontend-BR",
			scraper: scrapers.NovoGitHubScraper("frontendbr/vagas", "Frontend-BR"),
		},
		{
			nome:    "GitHub React-Brasil",
			scraper: scrapers.NovoGitHubScraper("react-brasil/vagas", "React-Brasil"),
		},
		{
			nome:    "GitHub QA-Brasil",
			scraper: scrapers.NovoGitHubScraper("qa-brasil/vagas", "QA-Brasil"),
		},
		{
			nome:    "ProgramaThor",
			scraper: scrapers.NovoProgramaThorScraper(),
		},
		{
			nome:    "RemoteOK",
			scraper: scrapers.NovoRemoteOKScraper(),
		},
		{
			nome:    "WeWorkRemotely",
			scraper: scrapers.NovoWeWorkRemotelyScraper(),
		},
	}

	for _, p := range provedores {
		p := p
		t.Run(p.nome, func(t *testing.T) {
			inicio := time.Now()
			vagas, err := p.scraper.Buscar()
			duracao := time.Since(inicio)

			if err != nil {
				t.Fatalf("❌ [%s] Falhou na requisição/extração: %v (levou %v)", p.nome, err, duracao)
			}

			if len(vagas) == 0 {
				t.Errorf("⚠️ [%s] Nenhuma vaga retornada em %v (possível alteração de HTML/API ou bloqueio)", p.nome, duracao)
				return
			}

			// Validar integridade dos dados da primeira vaga
			primeira := vagas[0]
			if primeira.Titulo == "" {
				t.Errorf("❌ [%s] Vaga sem título retornado", p.nome)
			}
			if primeira.Link == "" {
				t.Errorf("❌ [%s] Vaga sem link retornado", p.nome)
			}

			fmt.Printf("✅ [%s] OK: %d vagas capturadas em %v (Ex: \"%s\" na %s)\n",
				p.nome, len(vagas), duracao.Round(time.Millisecond), primeira.Titulo, primeira.Empresa)
		})
	}
}
