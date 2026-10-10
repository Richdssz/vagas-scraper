package scrapers

import (
	"strings"
	"unicode"
	"vagas-scraper/internal/models"
)

// Scraper é a interface que todos os rastreadores de vagas devem implementar.
type Scraper interface {
	Nome() string
	Buscar() ([]models.Vaga, error)
}

// Termos comuns que indicam apenas nível de senioridade ou formato de contratação (sem indicar nicho de TI).
var termosNivel = map[string]bool{
	"estagio":    true,
	"estagios":   true,
	"estagiario": true,
	"estagiaria": true,
	"intern":     true,
	"internship": true,
	"trainee":    true,
	"junior":     true,
	"jr":         true,
	"iniciante":  true,
	"aprendiz":   true,
}

// Indicadores que confirmam que a oportunidade pertence à área técnica/TI.
var indicadoresTech = []string{
	"desenvolvedor", "desenvolvedora", "desenvolvimento", "dev",
	"software", "programador", "programadora", "programacao",
	"ti", "tecnologia", "computacao", "informatica", "sistemas",
	"backend", "frontend", "fullstack", "mobile", "web",
	"cloud", "devops", "qa", "dados", "redes", "seguranca da informacao",
	"engenharia de software", "engenheiro de software", "engenheira de software",
	"analista de sistemas", "analista de ti", "suporte ti",
	"java", "python", "node", "react", "spring", "flutter", "dart", "sql",
}

// NormalizarTexto converte para minúsculas, remove acentos e limpa caracteres especiais,
// preservando termos técnicos como c# e c++, e separando palavras coladas (ex: SêniorNOVA -> senior nova).
func NormalizarTexto(s string) string {
	var sb strings.Builder
	var prev rune
	for _, r := range s {
		// Separa palavras coladas no padrão minúsculaMaiúscula (ex: SêniorNOVA, PlenoCLT, BackEnd)
		if unicode.IsUpper(r) && unicode.IsLower(prev) {
			sb.WriteRune(' ')
		}
		prev = r

		rLow := unicode.ToLower(r)
		switch rLow {
		case 'á', 'à', 'â', 'ã', 'ä', 'å':
			sb.WriteRune('a')
		case 'é', 'è', 'ê', 'ë':
			sb.WriteRune('e')
		case 'í', 'ì', 'î', 'ï':
			sb.WriteRune('i')
		case 'ó', 'ò', 'ô', 'õ', 'ö':
			sb.WriteRune('o')
		case 'ú', 'ù', 'û', 'ü':
			sb.WriteRune('u')
		case 'ç':
			sb.WriteRune('c')
		case 'ñ':
			sb.WriteRune('n')
		default:
			if (rLow >= 'a' && rLow <= 'z') || (rLow >= '0' && rLow <= '9') || rLow == '#' || rLow == '+' {
				sb.WriteRune(rLow)
			} else {
				sb.WriteRune(' ')
			}
		}
	}
	return strings.Join(strings.Fields(sb.String()), " ")
}

// ContemTermo verifica se termoNorm ocorre com limites de palavra em textoNorm.
// Isso impede falsos positivos (ex: "sr" casando dentro de "disrupcao", ou "ti" dentro de "assistente").
func ContemTermo(textoNorm, termoNorm string) bool {
	if termoNorm == "" || textoNorm == "" {
		return false
	}
	alvo := " " + textoNorm + " "
	padrao := " " + termoNorm + " "
	return strings.Contains(alvo, padrao)
}

// obterVariantes retorna sinônimos ou flexões comuns em títulos de vagas
// (ex: estagio -> estagiario, junior -> jr, senior -> sr).
func obterVariantes(termoNorm string) []string {
	switch termoNorm {
	case "estagio", "estagios", "estagiario", "estagiaria":
		return []string{"estagio", "estagios", "estagiario", "estagiaria", "intern", "internship"}
	case "junior", "jr", "jrs":
		return []string{"junior", "jr", "jrs"}
	case "desenvolvedor", "desenvolvedora", "desenvolvedores", "dev", "devs":
		return []string{"desenvolvedor", "desenvolvedora", "desenvolvedores", "dev", "devs"}
	case "programador", "programadora", "programadores":
		return []string{"programador", "programadora", "programadores"}
	case "pleno", "plenos", "mid":
		return []string{"pleno", "plenos", "mid"}
	case "senior", "seniors", "sr", "srs":
		return []string{"senior", "seniors", "sr", "srs"}
	default:
		return []string{termoNorm}
	}
}

// ContemTermoOuVariante verifica se o termo ou suas flexões ocorrem em textoNorm.
func ContemTermoOuVariante(textoNorm, termoNorm string) bool {
	for _, v := range obterVariantes(termoNorm) {
		if ContemTermo(textoNorm, v) {
			return true
		}
	}
	return false
}

// ehFonteDedicadaTech retorna se a fonte é um canal 100% voltado para desenvolvedores/TI
// (onde qualquer vaga de estágio ou júnior já é nativamente de tecnologia).
func ehFonteDedicadaTech(fonte string) bool {
	f := strings.ToLower(fonte)
	return strings.Contains(f, "backend") ||
		strings.Contains(f, "frontend") ||
		strings.Contains(f, "react") ||
		strings.Contains(f, "qa") ||
		strings.Contains(f, "programathor") ||
		strings.Contains(f, "github")
}

// FiltroAceitaVaga avalia se o título ou tags da vaga combinam com os termos desejados
// e não contêm nenhum dos termos proibidos (ex: senior, pleno, pedagogia, contabilidade).
func FiltroAceitaVaga(vaga models.Vaga, termosBusca, termosExclusao []string) bool {
	textoNorm := NormalizarTexto(vaga.Titulo + " " + strings.Join(vaga.Tags, " "))
	empresaNorm := NormalizarTexto(vaga.Empresa)

	// 1. Verifica exclusões primeiro (se casar termo proibido, descarta na hora)
	for _, excl := range termosExclusao {
		exclNorm := NormalizarTexto(excl)
		if exclNorm == "" {
			continue
		}
		if ContemTermoOuVariante(textoNorm, exclNorm) || ContemTermoOuVariante(empresaNorm, exclNorm) {
			return false
		}
	}

	// 2. Se não houver termos de busca especificados, aceita todas que passaram na exclusão
	if len(termosBusca) == 0 {
		return true
	}

	// 3. Verifica termos desejados
	casouTermoTechOuEspecifico := false
	casouTermoNivel := false

	for _, busca := range termosBusca {
		buscaNorm := NormalizarTexto(busca)
		if buscaNorm == "" {
			continue
		}
		if ContemTermoOuVariante(textoNorm, buscaNorm) {
			if termosNivel[buscaNorm] {
				casouTermoNivel = true
			} else {
				casouTermoTechOuEspecifico = true
			}
		}
	}

	// Se casou com termo de tecnologia/nicho específico (ex: "backend", "java", "estagio ti"), aceita
	if casouTermoTechOuEspecifico {
		return true
	}

	// Se casou apenas com termo genérico de nível (ex: "estagio", "junior", "trainee"):
	if casouTermoNivel {
		// Em portais dedicados de TI (GitHub Backend-BR, ProgramaThor, etc.), já sabemos que o nicho é tech
		if ehFonteDedicadaTech(vaga.Fonte) {
			return true
		}

		// Em portais generalistas (Gupy, LinkedIn, etc.), exige contexto técnico no texto
		for _, ind := range indicadoresTech {
			if ContemTermoOuVariante(textoNorm, ind) {
				return true
			}
		}
	}

	return false
}

// ResolverTermosPortais seleciona termos inteligentes e contextuais para pesquisa em portais
// abertos como Gupy e LinkedIn, evitando pesquisas genéricas como apenas "estagio" que retornam
// vagas de áreas médicas, pedagógicas, administrativas ou financeiras.
func ResolverTermosPortais(termosBusca []string) []string {
	if len(termosBusca) == 0 {
		return []string{"desenvolvedor"}
	}

	var selecionados []string
	adicionado := make(map[string]bool)

	adicionar := func(termo string) {
		termo = strings.TrimSpace(termo)
		if termo != "" && !adicionado[strings.ToLower(termo)] {
			adicionado[strings.ToLower(termo)] = true
			selecionados = append(selecionados, termo)
		}
	}

	temNivelEstagio := false
	temNivelJunior := false
	var termosEspecificos []string

	for _, t := range termosBusca {
		tNorm := NormalizarTexto(t)
		if tNorm == "estagio" || tNorm == "estagiario" || tNorm == "trainee" {
			temNivelEstagio = true
		} else if tNorm == "junior" || tNorm == "jr" {
			temNivelJunior = true
		} else if len(strings.Fields(tNorm)) > 1 || (!termosNivel[tNorm] && len(tNorm) >= 3) {
			termosEspecificos = append(termosEspecificos, t)
		}
	}

	// 1. Se busca estágio, prioriza busca qualificada de estágio de TI/desenvolvimento
	if temNivelEstagio {
		adicionar("estagio ti")
	}

	// 2. Se busca júnior, prioriza desenvolvedor júnior
	if temNivelJunior {
		adicionar("desenvolvedor junior")
	}

	// 3. Adiciona termos técnicos do usuário (ex: backend, fullstack)
	for _, te := range termosEspecificos {
		adicionar(te)
	}

	// Se sobrou apenas estágio e mais nada, adiciona estagio desenvolvedor como segunda opção
	if temNivelEstagio && len(selecionados) < 2 {
		adicionar("estagio desenvolvedor")
	}

	if len(selecionados) == 0 {
		adicionar("desenvolvedor")
	}

	// Limita a no máximo 2 termos para consultas rápidas e sem rate limit
	if len(selecionados) > 2 {
		return selecionados[:2]
	}
	return selecionados
}
