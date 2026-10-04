package web

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
	"vagas-scraper/internal/config"
	"vagas-scraper/internal/dedup"
	"vagas-scraper/internal/models"
	"vagas-scraper/internal/notifier"
	"vagas-scraper/internal/scrapers"
	"vagas-scraper/internal/storage"
)

//go:embed index.html
var htmlPainelControle []byte

type ConfigPayload struct {
	TermosBusca         []string `json:"termos_busca"`
	TermosExclusao      []string `json:"termos_exclusao"`
	FontesHabilitadas   []string `json:"fontes_habilitadas"`
	MaxVagasPorExecucao int      `json:"max_vagas_por_execucao"`
	EmailRemetente      string   `json:"email_remetente"`
	EmailSenhaApp       string   `json:"email_senha_app"`
	EmailDestinatario   string   `json:"email_destinatario"`
	DryRun              bool     `json:"dry_run"`
}

type ServidorWeb struct {
	porta         string
	caminhoConfig string
}

func NovoServidorWeb(porta, caminhoConfig string) *ServidorWeb {
	if porta == "" {
		porta = "8080"
	}
	return &ServidorWeb{
		porta:         porta,
		caminhoConfig: caminhoConfig,
	}
}

func (s *ServidorWeb) Iniciar() error {
	mux := http.NewServeMux()

	mux.HandleFunc("/", s.handleIndex)
	mux.HandleFunc("/api/config", s.handleConfig)
	mux.HandleFunc("/api/scrape", s.handleScrape)
	mux.HandleFunc("/api/schedule-windows", s.handleScheduleWindows)
	mux.HandleFunc("/api/sync-github", s.handleSyncGitHub)

	urlAcesso := fmt.Sprintf("http://localhost:%s", s.porta)
	log.Printf("🌐 Servidor Web ativo em: %s", urlAcesso)
	log.Println("💡 Abrindo painel de controle no seu navegador padrão...")

	go abrirNavegador(urlAcesso)

	return http.ListenAndServe(":"+s.porta, mux)
}

func (s *ServidorWeb) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(htmlPainelControle)
}

func (s *ServidorWeb) handleConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodGet {
		cfg, err := config.Carregar(s.caminhoConfig)
		if err != nil {
			http.Error(w, `{"erro":"falha ao ler configuracao"}`, http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(ConfigPayload{
			TermosBusca:         cfg.Filtros.TermosBusca,
			TermosExclusao:      cfg.Filtros.TermosExclusao,
			FontesHabilitadas:   cfg.FontesHabilitadas,
			MaxVagasPorExecucao: cfg.MaxVagasPorExecucao,
			EmailRemetente:      cfg.EmailRemetente,
			EmailSenhaApp:       mascararSenha(cfg.EmailSenhaApp),
			EmailDestinatario:   cfg.EmailDestinatario,
			DryRun:              cfg.DryRun,
		})
		return
	}

	if r.Method == http.MethodPost {
		var payload ConfigPayload
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, `{"erro":"dados invalidos"}`, http.StatusBadRequest)
			return
		}

		cfg, _ := config.Carregar(s.caminhoConfig)
		if cfg == nil {
			cfg = &config.AppConfig{}
		}

		cfg.Filtros.TermosBusca = payload.TermosBusca
		cfg.Filtros.TermosExclusao = payload.TermosExclusao
		cfg.FontesHabilitadas = payload.FontesHabilitadas
		if payload.MaxVagasPorExecucao > 0 {
			cfg.MaxVagasPorExecucao = payload.MaxVagasPorExecucao
		}

		configData, err := json.MarshalIndent(cfg, "", "  ")
		if err != nil {
			http.Error(w, `{"erro":"erro ao serializar config.json"}`, http.StatusInternalServerError)
			return
		}
		_ = os.WriteFile(s.caminhoConfig, configData, 0644)

		senhaParaSalvar := strings.ReplaceAll(strings.TrimSpace(payload.EmailSenhaApp), " ", "")
		if strings.Contains(senhaParaSalvar, "•••") {
			cfgAtual, _ := config.Carregar(s.caminhoConfig)
			senhaParaSalvar = cfgAtual.EmailSenhaApp
		}

		envContent := fmt.Sprintf(`EMAIL_REMETENTE=%s
EMAIL_SENHA_APP=%s
EMAIL_DESTINATARIO=%s
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
DRY_RUN=%t
`, payload.EmailRemetente, senhaParaSalvar, payload.EmailDestinatario, payload.DryRun)

		_ = os.WriteFile(".env", []byte(envContent), 0644)

		w.Write([]byte(`{"sucesso":true,"mensagem":"Configurações salvas com sucesso!"}`))
		return
	}

	http.Error(w, `{"erro":"metodo nao permitido"}`, http.StatusMethodNotAllowed)
}

func (s *ServidorWeb) handleScrape(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"erro":"metodo nao permitido"}`, http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	cfg, err := config.Carregar(s.caminhoConfig)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"erro":"%v"}`, err), http.StatusInternalServerError)
		return
	}

	enviarEmailReal := r.URL.Query().Get("email") == "true" && !cfg.DryRun

	var listaScrapers []scrapers.Scraper
	termoPrincipal := "desenvolvedor"
	if len(cfg.Filtros.TermosBusca) > 0 {
		termoPrincipal = cfg.Filtros.TermosBusca[0]
	}

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

	var wg sync.WaitGroup
	var mu sync.Mutex
	var todasVagas []models.Vaga

	for _, sc := range listaScrapers {
		wg.Add(1)
		go func(scraper scrapers.Scraper) {
			defer wg.Done()
			resultado, err := scraper.Buscar()
			if err != nil {
				log.Printf("⚠️ Erro na fonte [%s]: %v", scraper.Nome(), err)
				return
			}
			mu.Lock()
			todasVagas = append(todasVagas, resultado...)
			mu.Unlock()
		}(sc)
	}
	wg.Wait()

	vagasMescladas := dedup.MesclarVagasCruzadas(todasVagas)

	repoStorage, _ := storage.NovoStorage("vagas_vistas.json")
	var selecionadas []models.Vaga

	for _, v := range vagasMescladas {
		if scrapers.FiltroAceitaVaga(v, cfg.Filtros.TermosBusca, cfg.Filtros.TermosExclusao) {
			selecionadas = append(selecionadas, v)
			if len(selecionadas) >= cfg.MaxVagasPorExecucao {
				break
			}
		}
	}

	if enviarEmailReal && len(selecionadas) > 0 {
		notif := notifier.NovoNotificador(cfg)
		_ = notif.Enviar(selecionadas)
		ids := make(map[string]string)
		for _, v := range selecionadas {
			ids[v.ID] = v.Titulo
		}
		_ = repoStorage.Adicionar(ids)
	}

	resposta := map[string]any{
		"sucesso":         true,
		"total_bruto":     len(todasVagas),
		"total_mescladas": len(vagasMescladas),
		"total_filtradas": len(selecionadas),
		"email_disparado": enviarEmailReal,
		"vagas":           selecionadas,
	}

	json.NewEncoder(w).Encode(resposta)
}

func (s *ServidorWeb) handleSyncGitHub(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"erro":"metodo nao permitido"}`, http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	// 1. git add .
	addCmd := exec.Command("git", "add", ".")
	if out, err := addCmd.CombinedOutput(); err != nil {
		w.Write([]byte(fmt.Sprintf(`{"sucesso":false,"mensagem":"Falha no git add: %s"}`, strings.TrimSpace(string(out)))))
		return
	}

	// 2. git commit (ignora erro caso não haja nada novo para commit)
	commitCmd := exec.Command("git", "commit", "-m", "chore: atualizar filtros e preferencias via painel web")
	_ = commitCmd.Run()

	// 3. git push origin main
	pushCmd := exec.Command("git", "push", "origin", "main")
	if out, err := pushCmd.CombinedOutput(); err != nil {
		w.Write([]byte(fmt.Sprintf(`{"sucesso":false,"mensagem":"Falha no git push: %s"}`, strings.TrimSpace(string(out)))))
		return
	}

	w.Write([]byte(`{"sucesso":true,"mensagem":"Configurações e filtros enviados com sucesso para o GitHub!"}`))
}

func (s *ServidorWeb) handleScheduleWindows(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"erro":"metodo nao permitido"}`, http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	if runtime.GOOS != "windows" {
		w.Write([]byte(`{"sucesso":false,"mensagem":"O agendamento automático é exclusivo para sistemas Windows."}`))
		return
	}

	dir, err := os.Getwd()
	if err != nil {
		exePath, _ := os.Executable()
		dir = filepath.Dir(exePath)
	}
	scraperPath := filepath.Join(dir, "scraper.exe")
	if _, err := os.Stat(scraperPath); os.IsNotExist(err) {
		w.Write([]byte(fmt.Sprintf(`{"sucesso":false,"mensagem":"Arquivo scraper.exe não encontrado em '%s'. Certifique-se de que o scraper.exe está na pasta do projeto."}`, dir)))
		return
	}

	// Execução silenciosa em segundo plano via powershell -WindowStyle Hidden
	targetCmd := fmt.Sprintf(`powershell.exe -WindowStyle Hidden -Command "Start-Process -FilePath '%s' -WorkingDirectory '%s'"`, scraperPath, dir)

	horarios := []struct {
		nome string
		hora string
	}{
		{"VagasScraper_09h", "09:00"},
		{"VagasScraper_14h", "14:00"},
		{"VagasScraper_19h", "19:00"},
	}

	var erros []string
	for _, h := range horarios {
		cmd := exec.Command("schtasks", "/Create", "/SC", "DAILY", "/TN", h.nome, "/TR", targetCmd, "/ST", h.hora, "/F")
		if out, err := cmd.CombinedOutput(); err != nil {
			erros = append(erros, fmt.Sprintf("%s (%s): %s", h.nome, h.hora, strings.TrimSpace(string(out))))
		}
	}

	if len(erros) > 0 {
		msg := fmt.Sprintf("Erro ao registrar tarefas no Windows: %s", strings.Join(erros, "; "))
		log.Printf("⚠️ %s", msg)
		w.Write([]byte(fmt.Sprintf(`{"sucesso":false,"mensagem":"%s"}`, msg)))
		return
	}

	w.Write([]byte(`{"sucesso":true,"mensagem":"Tarefas agendadas com sucesso no Windows! O robô rodará automaticamente às 09:00, 14:00 e 19:00 sem abrir janelas."}`))
}

func mascararSenha(senha string) string {
	if len(senha) <= 4 {
		return "••••••••"
	}
	return senha[:2] + "••••••••" + senha[len(senha)-2:]
}

func abrirNavegador(url string) {
	time.Sleep(500 * time.Millisecond)
	switch runtime.GOOS {
	case "windows":
		_ = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		_ = exec.Command("open", url).Start()
	default:
		_ = exec.Command("xdg-open", url).Start()
	}
}
