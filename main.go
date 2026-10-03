package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"sync"
	"time"
	"vagas-scraper/internal/config"
	"vagas-scraper/internal/dedup"
	"vagas-scraper/internal/models"
	"vagas-scraper/internal/notifier"
	"vagas-scraper/internal/scrapers"
	"vagas-scraper/internal/storage"
	"vagas-scraper/internal/web"
)

const versao = "1.2.0"

func main() {
	modoScrape := flag.Bool("scrape", false, "Executa uma varredura silenciosa e encerra (usado no agendador/cron)")
	portaWeb := flag.String("port", "8080", "Porta para o painel web local")
	caminhoConfig := flag.String("config", "config.json", "Caminho do arquivo de configuração JSON")
	dryRunFlag := flag.Bool("dry-run", false, "Executa em modo simulação sem disparar e-mails reais")
	limparHistorico := flag.Bool("limpar-historico", false, "Limpa o histórico de vagas vistas")
	exibirVersao := flag.Bool("version", false, "Exibe a versão do programa")
	flag.Parse()

	if *exibirVersao {
		fmt.Printf("Vagas Scraper v%s • Ultraleve em Go Puro\n", versao)
		return
	}

	// SE NÃO FOR MODO SCRAPE HEADLESS: Inicia o Servidor Web e abre no navegador!
	if !*modoScrape {
		log.Printf("🚀 Iniciando Painel Visual do Vagas Scraper v%s...", versao)
		servidor := web.NovoServidorWeb(*portaWeb, *caminhoConfig)
		if err := servidor.Iniciar(); err != nil {
			log.Fatalf("❌ Erro ao iniciar servidor web: %v", err)
		}
		return
	}

	// MODO SCRAPE HEADLESS (Executado pelo Agendador do Windows ou GitHub Actions)
	log.Printf("⚡ Executando ciclo agendado do Vagas Scraper v%s...", versao)

	cfg, err := config.Carregar(*caminhoConfig)
	if err != nil {
		log.Fatalf("❌ Erro de configuração: %v", err)
	}

	if *dryRunFlag {
		cfg.DryRun = true
	}

	caminhoHistorico := "vagas_vistas.json"
	if *limparHistorico {
		_ = os.Remove(caminhoHistorico)
		log.Println("🧹 Histórico de vagas limpo.")
	}

	repoStorage, err := storage.NovoStorage(caminhoHistorico)
	if err != nil {
		log.Fatalf("❌ Erro ao abrir histórico: %v", err)
	}

	// 1. Instancia fontes habilitadas
	termoPrincipal := "desenvolvedor"
	if len(cfg.Filtros.TermosBusca) > 0 {
		termoPrincipal = cfg.Filtros.TermosBusca[0]
	}

	var listaScrapers []scrapers.Scraper
	for _, fonte := range cfg.FontesHabilitadas {
		switch fonte {
		case "linkedin":
			listaScrapers = append(listaScrapers, scrapers.NovoLinkedInScraper(termoPrincipal))
		case "gupy":
			listaScrapers = append(listaScrapers, scrapers.NovoGupyScraper(termoPrincipal))
		case "backend_br":
			listaScrapers = append(listaScrapers, scrapers.NovoGitHubScraper("backend-br/vagas", "Backend-BR"))
		case "frontend_br":
			listaScrapers = append(listaScrapers, scrapers.NovoGitHubScraper("frontendbr/vagas", "Frontend-BR"))
		case "react_brasil":
			listaScrapers = append(listaScrapers, scrapers.NovoGitHubScraper("react-brasil/vagas", "React-Brasil"))
		case "qa_brasil":
			listaScrapers = append(listaScrapers, scrapers.NovoGitHubScraper("qa-brasil/vagas", "QA-Brasil"))
		case "programathor":
			listaScrapers = append(listaScrapers, scrapers.NovoProgramaThorScraper())
		case "remoteok":
			listaScrapers = append(listaScrapers, scrapers.NovoRemoteOKScraper())
		case "weworkremotely":
			listaScrapers = append(listaScrapers, scrapers.NovoWeWorkRemotelyScraper())
		}
	}

	if len(listaScrapers) == 0 {
		log.Fatal("❌ Nenhuma fonte habilitada no config.json.")
	}

	// 2. Consulta em paralelo
	log.Printf("🔍 Consultando %d fontes em paralelo...", len(listaScrapers))
	inicio := time.Now()

	var wg sync.WaitGroup
	var mu sync.Mutex
	var todasVagas []models.Vaga

	for _, s := range listaScrapers {
		wg.Add(1)
		go func(scraper scrapers.Scraper) {
			defer wg.Done()
			resultado, err := scraper.Buscar()
			if err != nil {
				log.Printf("⚠️ [%s]: %v", scraper.Nome(), err)
				return
			}
			log.Printf("✔️ [%s] retornou %d oportunidades.", scraper.Nome(), len(resultado))
			mu.Lock()
			todasVagas = append(todasVagas, resultado...)
			mu.Unlock()
		}(s)
	}

	wg.Wait()
	duracao := time.Since(inicio)
	log.Printf("⚡ Varredura concluída em %v. Total bruto: %d vagas.", duracao.Round(time.Millisecond), len(todasVagas))

	// 3. Deduplicação Cruzada Multi-Plataforma
	vagasMescladas := dedup.MesclarVagasCruzadas(todasVagas)
	log.Printf("🧬 Após unificação de vagas presentes em múltiplos portais: %d vagas únicas.", len(vagasMescladas))

	// 4. Filtra vagas
	novasVagas := make([]models.Vaga, 0)
	idsParaRegistrar := make(map[string]string)

	for _, vaga := range vagasMescladas {
		if repoStorage.JaVista(vaga.ID) || repoStorage.JaVista(vaga.ChaveCanonica) {
			continue
		}

		if scrapers.FiltroAceitaVaga(vaga, cfg.Filtros.TermosBusca, cfg.Filtros.TermosExclusao) {
			novasVagas = append(novasVagas, vaga)
			idsParaRegistrar[vaga.ID] = vaga.Titulo
			idsParaRegistrar[vaga.ChaveCanonica] = vaga.Titulo

			if len(novasVagas) >= cfg.MaxVagasPorExecucao {
				break
			}
		}
	}

	if len(novasVagas) == 0 {
		log.Println("✨ Nenhuma nova vaga não vista encontrada nesta rodada.")
		return
	}

	log.Printf("🎯 %d novas vagas selecionadas e prontas para notificação!", len(novasVagas))

	notif := notifier.NovoNotificador(cfg)
	if err := notif.Enviar(novasVagas); err != nil {
		log.Fatalf("❌ Erro ao enviar notificação: %v", err)
	}

	_ = repoStorage.Adicionar(idsParaRegistrar)
	log.Printf("💾 Histórico atualizado com novas entradas.")
	log.Println("🏁 Execução concluída.")
}
