package notifier

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/smtp"
	"os"
	"strings"
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

	// Montagem dos cabeçalhos MIME
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

	// Envio seguro via SMTP com suporte a STARTTLS e SSL direto
	return n.enviarSMTP(message.String())
}

func (n *Notificador) enviarSMTP(mensagem string) error {
	addr := fmt.Sprintf("%s:%s", n.cfg.SMTPHost, n.cfg.SMTPPort)
	auth := smtp.PlainAuth("", n.cfg.EmailRemetente, n.cfg.EmailSenhaApp, n.cfg.SMTPHost)

	// Conecta ao servidor SMTP
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("falha ao conectar ao servidor SMTP %s: %w", addr, err)
	}

	client, err := smtp.NewClient(conn, n.cfg.SMTPHost)
	if err != nil {
		conn.Close()
		return fmt.Errorf("falha ao criar cliente SMTP: %w", err)
	}
	defer client.Quit()

	// Se o servidor suportar STARTTLS, ativa TLS
	if ok, _ := client.Extension("STARTTLS"); ok {
		tlsConfig := &tls.Config{
			ServerName: n.cfg.SMTPHost,
		}
		if err = client.StartTLS(tlsConfig); err != nil {
			return fmt.Errorf("falha ao iniciar STARTTLS: %w", err)
		}
	}

	// Autentica
	if err = client.Auth(auth); err != nil {
		return fmt.Errorf("erro de autenticação SMTP (verifique sua senha de app): %w", err)
	}

	// Remetente e Destinatário
	if err = client.Mail(n.cfg.EmailRemetente); err != nil {
		return fmt.Errorf("erro ao definir remetente: %w", err)
	}

	if err = client.Rcpt(n.cfg.EmailDestinatario); err != nil {
		return fmt.Errorf("erro ao definir destinatário: %w", err)
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
