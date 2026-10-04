package scrapers

import (
	"strings"
	"time"
	"vagas-scraper/internal/models"
)

// FiltrosAvancados define critérios extras de seleção de vagas.
// Campos vazios significam "sem restrição".
type FiltrosAvancados struct {
	HorasMaximas        int      // Publicadas há no máximo N horas (ex: 6, 12, 24, 48)
	DiasMaximos         int      // Publicadas há no máximo N dias (0 = sem limite)
	Localizacoes        []string // Cidades/estados/países aceitos (ex: "recife", "pe", "brasil", "sp")
	Modalidades         []string // "remoto", "hibrido", "presencial"
	Jornadas            []string // "integral", "meio_periodo", "estagio", "trainee"
	AceitarRemotoSempre bool     // Vagas remotas passam mesmo que o endereço não coincida
}

var palavrasModalidade = map[string][]string{
	"remoto":     {"remoto", "remote", "home office", "anywhere", "100% remoto", "trabalho remoto", "worldwide"},
	"hibrido":    {"híbrido", "hibrido", "hybrid"},
	"presencial": {"presencial", "on-site", "onsite", "in-office", "in office"},
}

var palavrasJornada = map[string][]string{
	"integral":     {"integral", "full-time", "full time", "fulltime", "40h"},
	"meio_periodo": {"meio período", "meio periodo", "part-time", "part time", "parttime", "20h", "30h", "4h", "6h"},
	"estagio":      {"estágio", "estagio", "intern", "internship", "estagiário", "estagiario"},
	"trainee":      {"trainee", "recém-formado", "recem-formado"},
}

// FiltroAvancadoAceita aplica os filtros de data, localização, modalidade e jornada.
// Quando a vaga não informa um atributo (ex: modalidade), ela é aceita para não perder oportunidades.
func FiltroAvancadoAceita(v models.Vaga, f FiltrosAvancados) bool {
	texto := strings.ToLower(v.Titulo + " " + strings.Join(v.Tags, " ") + " " + v.Localizacao)

	// 1. Data e horário de publicação
	if f.HorasMaximas > 0 && !v.Data.IsZero() {
		limite := time.Now().Add(-time.Duration(f.HorasMaximas) * time.Hour)
		if v.Data.Before(limite) {
			return false
		}
	} else if f.DiasMaximos > 0 && !v.Data.IsZero() {
		limite := time.Now().AddDate(0, 0, -f.DiasMaximos)
		if v.Data.Before(limite) {
			return false
		}
	}

	// 2. Modalidade
	if len(f.Modalidades) > 0 {
		detectadas := detectar(texto, palavrasModalidade)
		if len(detectadas) > 0 && !intersecta(detectadas, f.Modalidades) {
			return false
		}
	}

	// 3. Jornada
	if len(f.Jornadas) > 0 {
		detectadas := detectar(texto, palavrasJornada)
		if len(detectadas) > 0 && !intersecta(detectadas, f.Jornadas) {
			return false
		}
	}

	// 4. Localização / Endereço (vagas remotas passam se aceitar remoto ou se modalidade permitir)
	if len(f.Localizacoes) > 0 {
		local := strings.ToLower(v.Localizacao + " " + v.Titulo)
		ok := false
		for _, l := range f.Localizacoes {
			l = strings.TrimSpace(strings.ToLower(l))
			if l != "" && strings.Contains(local, l) {
				ok = true
				break
			}
		}
		if !ok {
			ehRemota := contem(detectar(texto, palavrasModalidade), "remoto") || strings.Contains(local, "remoto")
			aceitaRemoto := f.AceitarRemotoSempre || len(f.Modalidades) == 0 || contem(f.Modalidades, "remoto")
			if !(ehRemota && aceitaRemoto) {
				return false
			}
		}
	}

	return true
}

func detectar(texto string, mapa map[string][]string) []string {
	var achados []string
	for chave, palavras := range mapa {
		for _, p := range palavras {
			if strings.Contains(texto, p) {
				achados = append(achados, chave)
				break
			}
		}
	}
	return achados
}

func intersecta(a, b []string) bool {
	for _, x := range a {
		if contem(b, x) {
			return true
		}
	}
	return false
}

func contem(lista []string, valor string) bool {
	for _, x := range lista {
		if x == valor {
			return true
		}
	}
	return false
}
