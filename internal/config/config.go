package config

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// FiltrosConfig define palavras-chave aceitas e proibidas.
type FiltrosConfig struct {
	TermosBusca         []string `json:"termos_busca"`
	TermosExclusao      []string `json:"termos_exclusao"`
	HorasMaximas        int      `json:"horas_maximas,omitempty"`
	DiasMaximos         int      `json:"dias_maximos"`
	Localizacoes        []string `json:"localizacoes"`
	Modalidades         []string `json:"modalidades"`
	Jornadas            []string `json:"jornadas"`
	AceitarRemotoSempre bool     `json:"aceitar_remoto_sempre"`
}

// AppConfig armazena as configurações do sistema.
type AppConfig struct {
	Filtros             FiltrosConfig `json:"filtros"`
	FontesHabilitadas   []string      `json:"fontes_habilitadas"`
	MaxVagasPorExecucao int           `json:"max_vagas_por_execucao"`
	SalvarPreviewHTML   bool          `json:"salvar_preview_html"`
	FrequenciaExecucao  string        `json:"frequencia_execucao"`

	// Dados carregados de variáveis de ambiente / .env
	EmailRemetente    string `json:"-"`
	EmailSenhaApp     string `json:"-"`
	EmailDestinatario string `json:"-"`
	SMTPHost          string `json:"-"`
	SMTPPort          string `json:"-"`
	DryRun            bool   `json:"-"`
}

// Carregar lê o config.json e as variáveis de ambiente (.env ou SO).
func Carregar(caminhoConfig string) (*AppConfig, error) {
	// 1. Carrega arquivo .env se existir (sem sobrepor variáveis já exportadas)
	carregarArquivoDotEnv(".env")

	// 2. Lê config.json
	data, err := os.ReadFile(caminhoConfig)
	if err != nil {
		return nil, fmt.Errorf("falha ao ler %s: %w", caminhoConfig, err)
	}

	var cfg AppConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("falha ao decodificar %s: %w", caminhoConfig, err)
	}

	// 3. Aplica variáveis de ambiente
	cfg.EmailRemetente = strings.TrimSpace(os.Getenv("EMAIL_REMETENTE"))
	cfg.EmailSenhaApp = strings.ReplaceAll(strings.TrimSpace(os.Getenv("EMAIL_SENHA_APP")), " ", "")
	cfg.EmailDestinatario = strings.TrimSpace(os.Getenv("EMAIL_DESTINATARIO"))

	// Destinatário padrão pode ser o próprio remetente se não informado
	if cfg.EmailDestinatario == "" && cfg.EmailRemetente != "" {
		cfg.EmailDestinatario = cfg.EmailRemetente
	}

	cfg.SMTPHost = os.Getenv("SMTP_HOST")
	if cfg.SMTPHost == "" {
		cfg.SMTPHost = "smtp.gmail.com"
	}

	cfg.SMTPPort = os.Getenv("SMTP_PORT")
	if cfg.SMTPPort == "" {
		cfg.SMTPPort = "587"
	}

	dryRunEnv := strings.ToLower(strings.TrimSpace(os.Getenv("DRY_RUN")))
	if dryRunEnv == "true" || dryRunEnv == "1" || dryRunEnv == "yes" {
		cfg.DryRun = true
	}

	if cfg.MaxVagasPorExecucao <= 0 {
		cfg.MaxVagasPorExecucao = 25
	}

	if cfg.FrequenciaExecucao == "" {
		cfg.FrequenciaExecucao = "3x"
	}

	return &cfg, nil
}

// carregarArquivoDotEnv lê linhas chave=valor de um arquivo .env simples
func carregarArquivoDotEnv(caminho string) {
	file, err := os.Open(caminho)
	if err != nil {
		return // Arquivo opcional
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		linha := strings.TrimSpace(scanner.Text())
		if linha == "" || strings.HasPrefix(linha, "#") {
			continue
		}
		partes := strings.SplitN(linha, "=", 2)
		if len(partes) != 2 {
			continue
		}
		chave := strings.TrimSpace(partes[0])
		valor := strings.TrimSpace(partes[1])
		// Remove aspas simples ou duplas ao redor do valor
		if len(valor) >= 2 && ((valor[0] == '"' && valor[len(valor)-1] == '"') || (valor[0] == '\'' && valor[len(valor)-1] == '\'')) {
			valor = valor[1 : len(valor)-1]
		}
		// Apenas define se a variável ainda não existir no ambiente
		if _, existe := os.LookupEnv(chave); !existe {
			os.Setenv(chave, valor)
		}
	}
}

// StringFormatedPort retorna a porta como inteiro se necessário
func (c *AppConfig) PortInt() int {
	p, err := strconv.Atoi(c.SMTPPort)
	if err != nil {
		return 587
	}
	return p
}
