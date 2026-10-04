package notifier

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"
	"os"
	"strings"
	"vagas-scraper/internal/models"
)

// AssinarID gera a assinatura HMAC-SHA256 (hex) de um ID de vaga.
// O Worker valida a mesma assinatura para impedir que terceiros marquem vagas.
func AssinarID(segredo, id string) string {
	mac := hmac.New(sha256.New, []byte(segredo))
	mac.Write([]byte(id))
	return hex.EncodeToString(mac.Sum(nil))
}

// botaoMarcarVista devolve o HTML do botão "Já vi esta vaga".
// Retorna vazio se MARCAR_URL / MARCAR_SECRET não estiverem configurados.
func botaoMarcarVista(v models.Vaga) string {
	base := strings.TrimRight(strings.TrimSpace(os.Getenv("MARCAR_URL")), "/")
	segredo := strings.TrimSpace(os.Getenv("MARCAR_SECRET"))
	if base == "" || segredo == "" {
		return ""
	}

	q := url.Values{}
	q.Set("id", v.ID)
	q.Set("t", v.Titulo)
	q.Set("sig", AssinarID(segredo, v.ID))
	link := fmt.Sprintf("%s/marcar?%s", base, q.Encode())

	return fmt.Sprintf(`
				<a href="%s" target="_blank" style="display: inline-block; background-color: rgba(16, 185, 129, 0.1); border: 1px solid rgba(16, 185, 129, 0.3); color: #34d399; font-size: 12px; font-weight: 600; text-decoration: none; padding: 7px 14px; border-radius: 6px; margin-bottom: 6px;">
					&#10003; J&aacute; vi esta vaga
				</a>`, escapeHTML(link))
}
