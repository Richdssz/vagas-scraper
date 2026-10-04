package notifier

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/smtp"
	"os"
	"strings"
	"time"
	"vagas-scraper/internal/config"
	"vagas-scraper/internal/models"
)

// Notificador gerencia o disparo de alertas por e-mail
type Notificador struct {
	cfg *config.AppConfig
}

// NovoNotificador inicializa o serviço de notificação
func NovoNotificador(cfg *config.AppConfig) *Notificador {
	return &Notificador{cfg: cfg}
}

// Enviar envia a lista de novas vagas para o destinatário configurado
func (n *Notificador) Enviar(vagas []models.Vaga) error {
	if len(vagas) == 0 {
		log.Println("ℹ️ Nenhuma nova vaga para enviar.")
		return nil
	}

	htmlBody := GerarHTMLEmail(vagas)

	// Salva arquivo HTML local para conferência visual se habilitado
	if n.cfg.SalvarPreviewHTML {
		_ = os.WriteFile("preview_email.html", []byte(htmlBody), 0644)
		log.Println("💾 Pré-visualização salva em 'preview_email.html' (abra no navegador para conferir o visual).")
	}

	// Modo de Teste Local (Dry Run) ou credenciais não configuradas
	if n.cfg.DryRun || n.cfg.EmailRemetente == "" || n.cfg.EmailSenhaApp == "" {
		log.Printf("🧪 [MODO SIMULAÇÃO / DRY RUN] %d vagas encontradas! Detalhes:", len(vagas))
		for i, v := range vagas {
			fmt.Printf("   [%d] %s (%s) - %s\n       🔗 %s\n", i+1, v.Titulo, v.Empresa, v.Localizacao, v.Link)
		}
		if n.cfg.EmailRemetente == "" || n.cfg.EmailSenhaApp == "" {
			log.Println("⚠️ Credenciais de e-mail não configuradas no .env. Configure EMAIL_REMETENTE e EMAIL_SENHA_APP para disparar e-mails reais.")
		}
		return nil
	}

	log.Printf("📧 Disparando e-mail com %d vagas para %s...", len(vagas), n.cfg.EmailDestinatario)

	assunto := fmt.Sprintf("🎯 %d Novas Vagas Encontradas!", len(vagas))
	return n.dispararMensagem(assunto, htmlBody)
}

// EnviarAvisoSemVagas envia um e-mail informando que a rotina executou com sucesso
// porém nenhuma oportunidade nova inédita foi identificada.
func (n *Notificador) EnviarAvisoSemVagas(totalEncontradas int, portais []string) error {
	htmlBody := GerarHTMLEmailSemVagas(totalEncontradas, portais)

	if n.cfg.SalvarPreviewHTML {
		_ = os.WriteFile("preview_email.html", []byte(htmlBody), 0644)
	}

	if n.cfg.DryRun || n.cfg.EmailRemetente == "" || n.cfg.EmailSenhaApp == "" {
		log.Printf("🧪 [MODO SIMULAÇÃO / DRY RUN] Nenhuma vaga nova não vista. Notificação de aviso simulada com sucesso.")
		if n.cfg.EmailRemetente == "" || n.cfg.EmailSenhaApp == "" {
			log.Println("⚠️ Credenciais de e-mail não configuradas no .env.")
		}
		return nil
	}

	log.Printf("📧 Disparando e-mail de aviso (sem novas vagas) para %s...", n.cfg.EmailDestinatario)
	assunto := "Radar de Vagas: Nenhuma nova vaga nesta rodada"
	return n.dispararMensagem(assunto, htmlBody)
}

func (n *Notificador) dispararMensagem(assunto, htmlBody string) error {
	headers := make(map[string]string)
	headers["From"] = fmt.Sprintf("Vagas Scraper <%s>", n.cfg.EmailRemetente)
	headers["To"] = n.cfg.EmailDestinatario
	headers["Subject"] = assunto
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=UTF-8"

	var message strings.Builder
	for k, v := range headers {
		message.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	message.WriteString("\r\n")
	message.WriteString(htmlBody)

	return n.enviarSMTP(message.String())
}

func (n *Notificador) enviarSMTP(mensagem string) error {
	host := n.cfg.SMTPHost
	port := n.cfg.SMTPPort
	if port == "" {
		port = "587"
	}
	addr := fmt.Sprintf("%s:%s", host, port)
	auth := smtp.PlainAuth("", n.cfg.EmailRemetente, n.cfg.EmailSenhaApp, host)

	client, err := conectarSMTP(host, port)
	if err != nil && port == "587" {
		// Tentativa de fallback automático na porta 465 (SSL direto) caso a 587 esteja bloqueada
		log.Printf("⚠️ Falha na porta 587 (%v). Tentando porta 465 (SSL direto)...", err)
		if clientFallback, errFallback := conectarSMTP(host, "465"); errFallback == nil {
			client = clientFallback
			err = nil
		}
	}
	if err != nil {
		return fmt.Errorf("falha ao conectar ao servidor SMTP %s: %w", addr, err)
	}
	defer client.Quit()

	// Autentica
	if err = client.Auth(auth); err != nil {
		return fmt.Errorf("erro de autenticação SMTP: %w (Verifique se a Senha de App de 16 caracteres sem espaços foi gerada em https://myaccount.google.com/apppasswords na conta Google %s)", err, n.cfg.EmailRemetente)
	}

	// Remetente e Destinatário
	if err = client.Mail(n.cfg.EmailRemetente); err != nil {
		return fmt.Errorf("erro ao definir remetente (%s): %w", n.cfg.EmailRemetente, err)
	}

	if err = client.Rcpt(n.cfg.EmailDestinatario); err != nil {
		return fmt.Errorf("erro ao definir destinatário (%s): %w", n.cfg.EmailDestinatario, err)
	}

	// Envia corpo da mensagem
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("erro ao abrir fluxo de dados: %w", err)
	}

	if _, err = w.Write([]byte(mensagem)); err != nil {
		return fmt.Errorf("erro ao escrever mensagem: %w", err)
	}

	if err = w.Close(); err != nil {
		return fmt.Errorf("erro ao finalizar envio: %w", err)
	}

	log.Println("✅ E-mail enviado com sucesso!")
	return nil
}

func conectarSMTP(host, port string) (*smtp.Client, error) {
	addr := fmt.Sprintf("%s:%s", host, port)
	dialer := &net.Dialer{Timeout: 15 * time.Second}

	if port == "465" {
		tlsConfig := &tls.Config{ServerName: host}
		conn, err := tls.DialWithDialer(dialer, "tcp", addr, tlsConfig)
		if err != nil {
			return nil, err
		}
		return smtp.NewClient(conn, host)
	}

	conn, err := dialer.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	client, err := smtp.NewClient(conn, host)
	if err != nil {
		conn.Close()
		return nil, err
	}

	if ok, _ := client.Extension("STARTTLS"); ok {
		tlsConfig := &tls.Config{ServerName: host}
		if err = client.StartTLS(tlsConfig); err != nil {
			client.Close()
			return nil, fmt.Errorf("falha ao iniciar STARTTLS: %w", err)
		}
	}

	return client, nil
}
