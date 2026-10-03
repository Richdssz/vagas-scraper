package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"sync"
	"time"
	"vagas-scraper/internal/config"
	"vagas-scraper/internal/models"
	"vagas-scraper/internal/notifier"
	"vagas-scraper/internal/scrapers"
	"vagas-scraper/internal/storage"
)

const versao = "1.0.0"

func main() {
	caminhoConfig := flag.String("config", "config.json", "Caminho do arquivo de configuração JSON")
	dryRunFlag := flag.Bool("dry-run", false, "Executa sem disparar e-mails reais (gera preview HTML)")
	limparHistorico := flag.Bool("limpar-historico", false, "Limpa o arquivo de histórico de vagas vistas")
	exibirVersao := flag.Bool("version", false, "Exibe a versão do programa")
	flag.Parse()

	if *exibirVersao {
		fmt.Printf("Vagas Scraper v%s (compilado em Go puro)\n", versao)
		return
	}

	log.Printf("🚀 Iniciando Vagas Scraper v%s...", versao)

	// 1. Carrega configurações
	cfg, err := config.Carregar(*caminhoConfig)
	if err != nil {
		log.Fatalf("❌ Erro de configuração: %v", err)
	}

	if *dryRunFlag {
		cfg.DryRun = true
	}

	// 2. Gerenciador de histórico (Deduplicação)
	caminhoHistorico := "vagas_vistas.json"
	if *limparHistorico {
		_ = os.Remove(caminhoHistorico)
		log.Println("🧹 Histórico de vagas limpo com sucesso.")
	}

	repoStorage, err := storage.NovoStorage(caminhoHistorico)
	if err != nil {
		log.Fatalf("❌ Erro ao inicializar armazenamento: %v", err)
	}
	log.Printf("📦 Histórico atual: %d vagas já registradas.", repoStorage.TotalVistas())

	// 3. Monta lista de rastreadores ativos
	var listaScrapers []scrapers.Scraper
	for _, fonte := range cfg.FontesHabilitadas {
		switch fonte {
		case "backend_br":
			listaScrapers = append(listaScrapers, scrapers.NovoGitHubScraper("backend-br/vagas", "Backend-BR"))
		case "frontend_br":
			listaScrapers = append(listaScrapers, scrapers.NovoGitHubScraper("frontendbr/vagas", "Frontend-BR"))
		case "remoteok":
			listaScrapers = append(listaScrapers, scrapers.NovoRemoteOKScraper())
		default:
			log.Printf("⚠️ Fonte desconhecida ignorada: %s", fonte)
		}
	}

	if len(listaScrapers) == 0 {
		log.Fatal("❌ Nenhuma fonte de vagas habilitada no config.json.")
	}

	// 4. Executa os scrapers em paralelo (Goroutines)
	log.Printf("🔍 Consultando %d fontes de vagas em paralelo...", len(listaScrapers))
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
				log.Printf("⚠️ Erro na fonte [%s]: %v", scraper.Nome(), err)
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
	log.Printf("⚡ Varredura concluída em %v. Total bruto coletado: %d vagas.", duracao.Round(time.Millisecond), len(todasVagas))

	// 5. Filtra vagas de interesse e remove duplicatas
	novasVagas := make([]models.Vaga, 0)
	idsParaRegistrar := make(map[string]string)
	idsNestaRodada := make(map[string]bool)

	for _, vaga := range todasVagas {
		// Evita duplicatas da mesma rodada
		if idsNestaRodada[vaga.ID] {
			continue
		}

		// Verifica se já vimos em execuções anteriores
		if repoStorage.JaVista(vaga.ID) {
			continue
		}

		// Aplica filtros de termos desejados e exclusões
		if scrapers.FiltroAceitaVaga(vaga, cfg.Filtros.TermosBusca, cfg.Filtros.TermosExclusao) {
			idsNestaRodada[vaga.ID] = true
			novasVagas = append(novasVagas, vaga)
			idsParaRegistrar[vaga.ID] = vaga.Titulo

			if len(novasVagas) >= cfg.MaxVagasPorExecucao {
				log.Printf("🛑 Limite máximo de %d vagas por execução atingido.", cfg.MaxVagasPorExecucao)
				break
			}
		}
	}

	// 6. Notificação e atualização do histórico
	if len(novasVagas) == 0 {
		log.Println("✨ Nenhuma nova vaga correspondente encontrada nesta execução.")
		return
	}

	log.Printf("🎯 %d novas vagas selecionadas após filtros e deduplicação!", len(novasVagas))

	notif := notifier.NovoNotificador(cfg)
	if err := notif.Enviar(novasVagas); err != nil {
		log.Fatalf("❌ Erro ao enviar notificação: %v", err)
	}

	// Salva as novas vagas no histórico para não repetir
	if err := repoStorage.Adicionar(idsParaRegistrar); err != nil {
		log.Printf("⚠️ Falha ao atualizar histórico de vagas vistas: %v", err)
	} else {
		log.Printf("💾 Histórico atualizado com %d novas entradas.", len(idsParaRegistrar))
	}

	log.Println("🏁 Execução concluída com sucesso.")
}
