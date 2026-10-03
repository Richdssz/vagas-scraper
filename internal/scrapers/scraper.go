package scrapers

import (
	"strings"
	"vagas-scraper/internal/models"
)

// Scraper é a interface que todos os rastreadores de vagas devem implementar.
type Scraper interface {
	Nome() string
	Buscar() ([]models.Vaga, error)
}

// FiltroAceitaVaga avalia se o título ou tags da vaga combinam com os termos desejados
// e não contêm nenhum dos termos proibidos (ex: senior, pleno).
func FiltroAceitaVaga(vaga models.Vaga, termosBusca, termosExclusao []string) bool {
	texto := strings.ToLower(vaga.Titulo + " " + strings.Join(vaga.Tags, " ") + " " + vaga.Localizacao)

	// 1. Verifica exclusões primeiro
	for _, excl := range termosExclusao {
		excl = strings.TrimSpace(strings.ToLower(excl))
		if excl == "" {
			continue
		}
		if contemPalavraExataOuSubstr(texto, excl) {
			return false
		}
	}

	// 2. Se não houver termos de busca especificados, aceita todas
	if len(termosBusca) == 0 {
		return true
	}

	// 3. Verifica termos desejados
	for _, busca := range termosBusca {
		busca = strings.TrimSpace(strings.ToLower(busca))
		if busca == "" {
			continue
		}
		if contemPalavraExataOuSubstr(texto, busca) {
			return true
		}
	}

	return false
}

func contemPalavraExataOuSubstr(texto, termo string) bool {
	return strings.Contains(texto, termo)
}
