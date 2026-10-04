package web

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
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
	FrequenciaExecucao  string   `json:"frequencia_execucao"`
	HorasMaximas        int      `json:"horas_maximas"`
	DiasMaximos         int      `json:"dias_maximos"`
	Localizacoes        []string `json:"localizacoes"`
	Modalidades         []string `json:"modalidades"`
	Jornadas            []string `json:"jornadas"`
	AceitarRemotoSempre bool     `json:"aceitar_remoto_sempre"`
	EmailRemetente      string   `json:"email_remetente"`
	EmailSenhaApp       string   `json:"email_senha_app"`
	EmailDestinatario   string   `json:"email_destinatario"`
	DryRun              bool     `json:"dry_run"`
}

type SchedulePayload struct {
	Frequencia string `json:"frequencia"`
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
	mux.HandleFunc("/preview-email", s.handlePreviewEmail)
	mux.HandleFunc("/api/config", s.handleConfig)
	mux.HandleFunc("/api/scrape", s.handleScrape)
	mux.HandleFunc("/api/schedule-windows", s.handleScheduleWindows)
	mux.HandleFunc("/api/schedule-windows/status", s.handleScheduleWindowsStatus)
	mux.HandleFunc("/api/schedule-windows/cancel", s.handleScheduleWindowsCancel)
	mux.HandleFunc("/api/sync-github", s.handleSyncGitHub)
	mux.HandleFunc("/api/workflow-frequency", s.handleWorkflowFrequency)
	mux.HandleFunc("/api/health", s.handleHealth)

	urlAcesso := fmt.Sprintf("http://localhost:%s", s.porta)

	l, err := net.Listen("tcp", ":"+s.porta)
	if err != nil {
		log.Printf("O Painel de Controle já está ativo na porta %s! Abrindo seu navegador...", s.porta)
		abrirNavegador(urlAcesso)
		time.Sleep(1 * time.Second)
		return nil
	}

	log.Printf("Servidor Web ativo em: %s", urlAcesso)
	log.Println("Abrindo painel de controle no seu navegador padrão...")
	go abrirNavegador(urlAcesso)

	return http.Serve(l, mux)
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
			FrequenciaExecucao:  cfg.FrequenciaExecucao,
			HorasMaximas:        cfg.Filtros.HorasMaximas,
			DiasMaximos:         cfg.Filtros.DiasMaximos,
			Localizacoes:        cfg.Filtros.Localizacoes,
			Modalidades:         cfg.Filtros.Modalidades,
			Jornadas:            cfg.Filtros.Jornadas,
			AceitarRemotoSempre: cfg.Filtros.AceitarRemotoSempre,
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
		cfg.Filtros.HorasMaximas = payload.HorasMaximas
		cfg.Filtros.DiasMaximos = payload.DiasMaximos
		cfg.Filtros.Localizacoes = payload.Localizacoes
		cfg.Filtros.Modalidades = payload.Modalidades
		cfg.Filtros.Jornadas = payload.Jornadas
		cfg.Filtros.AceitarRemotoSempre = payload.AceitarRemotoSempre
		if payload.MaxVagasPorExecucao > 0 {
			cfg.MaxVagasPorExecucao = payload.MaxVagasPorExecucao
		}
		if payload.FrequenciaExecucao != "" {
			cfg.FrequenciaExecucao = payload.FrequenciaExecucao
			_ = atualizarCronWorkflow(payload.FrequenciaExecucao)
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

	filtrosAv := scrapers.FiltrosAvancados{
		HorasMaximas:        cfg.Filtros.HorasMaximas,
		DiasMaximos:         cfg.Filtros.DiasMaximos,
		Localizacoes:        cfg.Filtros.Localizacoes,
		Modalidades:         cfg.Filtros.Modalidades,
		Jornadas:            cfg.Filtros.Jornadas,
		AceitarRemotoSempre: cfg.Filtros.AceitarRemotoSempre,
	}

	for _, v := range vagasMescladas {
		if repoStorage.JaVista(v.ID) || repoStorage.JaVista(v.ChaveCanonica) {
			continue
		}
		if !scrapers.FiltroAvancadoAceita(v, filtrosAv) {
			continue
		}
		if scrapers.FiltroAceitaVaga(v, cfg.Filtros.TermosBusca, cfg.Filtros.TermosExclusao) {
			selecionadas = append(selecionadas, v)
			if len(selecionadas) >= cfg.MaxVagasPorExecucao {
				break
			}
		}
	}

	// Sempre atualiza o preview_email.html com o padrão idêntico de envio
	htmlBody := notifier.GerarHTMLEmail(selecionadas)
	_ = os.WriteFile("preview_email.html", []byte(htmlBody), 0644)

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

func (s *ServidorWeb) handlePreviewEmail(w http.ResponseWriter, r *http.Request) {
	data, err := os.ReadFile("preview_email.html")
	if err != nil {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte("<html><body style='font-family:sans-serif;padding:30px;text-align:center;'><h2>Nenhum preview gerado ainda</h2><p>Clique em 'Buscar vagas' no painel para gerar a pré-visualização mais recente.</p></body></html>"))
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(data)
}

func (s *ServidorWeb) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Write([]byte(fmt.Sprintf(`{"status":"online","porta":"%s","timestamp":%d}`, s.porta, time.Now().Unix())))
}

func (s *ServidorWeb) handleScheduleWindowsStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if runtime.GOOS != "windows" {
		w.Write([]byte(`{"agendado":false,"mensagem":"Apenas Windows"}`))
		return
	}

	out, err := exec.Command("schtasks", "/Query", "/FO", "LIST").CombinedOutput()
	if err != nil {
		w.Write([]byte(`{"agendado":false}`))
		return
	}

	texto := string(out)
	var horarios []string
	if strings.Contains(texto, "VagasScraper_08h") {
		horarios = append(horarios, "08:00")
	}
	if strings.Contains(texto, "VagasScraper_09h") {
		horarios = append(horarios, "09:00")
	}
	if strings.Contains(texto, "VagasScraper_12h") {
		horarios = append(horarios, "12:00")
	}
	if strings.Contains(texto, "VagasScraper_14h") {
		horarios = append(horarios, "14:00")
	}
	if strings.Contains(texto, "VagasScraper_16h") {
		horarios = append(horarios, "16:00")
	}
	if strings.Contains(texto, "VagasScraper_18h") {
		horarios = append(horarios, "18:00")
	}
	if strings.Contains(texto, "VagasScraper_19h") {
		horarios = append(horarios, "19:00")
	}
	if strings.Contains(texto, "VagasScraper_20h") {
		horarios = append(horarios, "20:00")
	}

	if len(horarios) > 0 {
		freq := fmt.Sprintf("%dx ao dia", len(horarios))
		json.NewEncoder(w).Encode(map[string]any{
			"agendado":   true,
			"frequencia": freq,
			"tarefas":    horarios,
		})
		return
	}

	w.Write([]byte(`{"agendado":false}`))
}

func (s *ServidorWeb) handleScheduleWindowsCancel(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"erro":"metodo nao permitido"}`, http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	tarefas := []string{
		"VagasScraper_08h", "VagasScraper_09h", "VagasScraper_12h", "VagasScraper_14h",
		"VagasScraper_16h", "VagasScraper_18h", "VagasScraper_19h", "VagasScraper_20h",
		"VagasScraperDiario", "VagasScraperTeste",
	}

	for _, t := range tarefas {
		_ = exec.Command("schtasks", "/Delete", "/TN", t, "/F").Run()
	}

	w.Write([]byte(`{"sucesso":true,"mensagem":"Agendamento cancelado com sucesso no Windows! Todas as tarefas foram removidas."}`))
}

func (s *ServidorWeb) handleWorkflowFrequency(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"erro":"metodo nao permitido"}`, http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	var payload SchedulePayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil || payload.Frequencia == "" {
		payload.Frequencia = "3x"
	}

	if err := atualizarCronWorkflow(payload.Frequencia); err != nil {
		w.Write([]byte(fmt.Sprintf(`{"sucesso":false,"mensagem":"Falha ao atualizar workflow: %v"}`, err)))
		return
	}

	cfg, _ := config.Carregar(s.caminhoConfig)
	if cfg != nil {
		cfg.FrequenciaExecucao = payload.Frequencia
		configData, _ := json.MarshalIndent(cfg, "", "  ")
		_ = os.WriteFile(s.caminhoConfig, configData, 0644)
	}

	w.Write([]byte(fmt.Sprintf(`{"sucesso":true,"mensagem":"Frequência do GitHub Actions atualizada para %s! Clique em 'Subir pro GitHub' para aplicar no repositório."}`, payload.Frequencia)))
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

	var payload SchedulePayload
	_ = json.NewDecoder(r.Body).Decode(&payload)
	frequencia := payload.Frequencia
	if frequencia == "" {
		frequencia = "3x"
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

	// 1. Remove tarefas antigas primeiro
	tarefasAntigas := []string{
		"VagasScraper_08h", "VagasScraper_09h", "VagasScraper_12h", "VagasScraper_14h",
		"VagasScraper_16h", "VagasScraper_18h", "VagasScraper_19h", "VagasScraper_20h",
		"VagasScraperDiario", "VagasScraperTeste",
	}
	for _, t := range tarefasAntigas {
		_ = exec.Command("schtasks", "/Delete", "/TN", t, "/F").Run()
	}

	// 2. Define os horários de acordo com a frequência escolhida
	type Horario struct {
		nome string
		hora string
	}
	var horarios []Horario

	switch frequencia {
	case "1x":
		horarios = []Horario{
			{"VagasScraper_09h", "09:00"},
		}
	case "2x":
		horarios = []Horario{
			{"VagasScraper_09h", "09:00"},
			{"VagasScraper_18h", "18:00"},
		}
	case "4x":
		horarios = []Horario{
			{"VagasScraper_08h", "08:00"},
			{"VagasScraper_12h", "12:00"},
			{"VagasScraper_16h", "16:00"},
			{"VagasScraper_20h", "20:00"},
		}
	default: // "3x"
		horarios = []Horario{
			{"VagasScraper_09h", "09:00"},
			{"VagasScraper_14h", "14:00"},
			{"VagasScraper_19h", "19:00"},
		}
	}

	targetCmd := fmt.Sprintf(`powershell.exe -WindowStyle Hidden -Command "Start-Process -FilePath '%s' -WorkingDirectory '%s'"`, scraperPath, dir)

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

	// Atualiza config.json com a frequência
	cfg, _ := config.Carregar(s.caminhoConfig)
	if cfg != nil {
		cfg.FrequenciaExecucao = frequencia
		configData, _ := json.MarshalIndent(cfg, "", "  ")
		_ = os.WriteFile(s.caminhoConfig, configData, 0644)
	}

	var horasList []string
	for _, h := range horarios {
		horasList = append(horasList, h.hora)
	}

	w.Write([]byte(fmt.Sprintf(`{"sucesso":true,"mensagem":"Tarefas agendadas com sucesso no Windows (%s: %s)! O robô rodará automaticamente sem abrir janelas."}`, frequencia, strings.Join(horasList, ", "))))
}

func atualizarCronWorkflow(frequencia string) error {
	caminho := filepath.Join(".github", "workflows", "scraper.yml")
	data, err := os.ReadFile(caminho)
	if err != nil {
		return err
	}
	cronMap := map[string]string{
		"1x": "'0 12 * * *'",
		"2x": "'0 12,21 * * *'",
		"3x": "'0 12,17,22 * * *'",
		"4x": "'0 11,15,19,23 * * *'",
	}
	cronVal, ok := cronMap[frequencia]
	if !ok {
		cronVal = "'0 12,17,22 * * *'"
	}
	re := regexp.MustCompile(`(?m)^\s*-\s*cron:\s*['"][^'"]+['"]`)
	novo := re.ReplaceAllString(string(data), fmt.Sprintf("    - cron: %s", cronVal))
	return os.WriteFile(caminho, []byte(novo), 0644)
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
