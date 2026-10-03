package dedup

import (
	"crypto/sha256"
	"fmt"
	"regexp"
	"strings"
	"vagas-scraper/internal/models"
)

var replacerAcentos = strings.NewReplacer(
	"á", "a", "à", "a", "ã", "a", "â", "a", "ä", "a",
	"é", "e", "è", "e", "ê", "e", "ë", "e",
	"í", "i", "ì", "i", "î", "i", "ï", "i",
	"ó", "o", "ò", "o", "õ", "o", "ô", "o", "ö", "o",
	"ú", "u", "ù", "u", "û", "u", "ü", "u",
	"ç", "c", "ñ", "n",
)

// GerarChaveCanonica cria uma impressão digital única (fingerprint) cruzando Empresa + Cargo
func GerarChaveCanonica(empresa, cargo string) string {
	cLimpo := normalizarTexto(cargo)
	eLimpa := normalizarEmpresa(empresa)

	if eLimpa == "" || eLimpa == "confidencial" || eLimpa == "naoinformada" || eLimpa == "vernoanuncio" {
		return fmt.Sprintf("sem_empresa:%x", sha256.Sum256([]byte(cLimpo)))[:16]
	}

	return fmt.Sprintf("%s:%s", eLimpa, cLimpo)
}

// MesclarVagasCruzadas identifica vagas idênticas vindas de portais diferentes
// (ex: mesma vaga na Gupy e no LinkedIn) e junta os links num único card enriquecido.
func MesclarVagasCruzadas(vagas []models.Vaga) []models.Vaga {
	mapa := make(map[string]*models.Vaga)
	ordem := make([]string, 0)

	for _, v := range vagas {
		if v.ChaveCanonica == "" {
			v.ChaveCanonica = GerarChaveCanonica(v.Empresa, v.Titulo)
		}

		existente, jaExiste := mapa[v.ChaveCanonica]
		if !jaExiste {
			nova := v
			if len(nova.Links) == 0 && nova.Link != "" {
				nova.Links = []models.LinkFonte{{Fonte: nova.Fonte, URL: nova.Link}}
			}
			if len(nova.Fontes) == 0 && nova.Fonte != "" {
				nova.Fontes = []string{nova.Fonte}
			}
			mapa[v.ChaveCanonica] = &nova
			ordem = append(ordem, v.ChaveCanonica)
		} else {
			// A mesma vaga foi encontrada em outro portal! Adiciona o novo link
			linkNovo := v.Link
			fonteNova := v.Fonte

			// Evita duplicar o mesmo link se a fonte for igual
			linkJaAdicionado := false
			for _, lf := range existente.Links {
				if lf.URL == linkNovo {
					linkJaAdicionado = true
					break
				}
			}

			if !linkJaAdicionado && linkNovo != "" {
				existente.Links = append(existente.Links, models.LinkFonte{
					Fonte: fonteNova,
					URL:   linkNovo,
				})
			}

			fonteJaPresente := false
			for _, f := range existente.Fontes {
				if f == fonteNova {
					fonteJaPresente = true
					break
				}
			}
			if !fonteJaPresente && fonteNova != "" {
				existente.Fontes = append(existente.Fontes, fonteNova)
			}

			if len(v.Empresa) > len(existente.Empresa) && (existente.Empresa == "Ver no anúncio" || existente.Empresa == "") {
				existente.Empresa = v.Empresa
			}
			if len(v.Tags) > len(existente.Tags) {
				existente.Tags = v.Tags
			}
		}
	}

	resultado := make([]models.Vaga, 0, len(ordem))
	for _, chave := range ordem {
		resultado = append(resultado, *mapa[chave])
	}

	return resultado
}

func normalizarTexto(s string) string {
	s = removerAcentos(strings.ToLower(s))

	s = strings.ReplaceAll(s, "back-end", "backend")
	s = strings.ReplaceAll(s, "front-end", "frontend")
	s = strings.ReplaceAll(s, "full-stack", "fullstack")
	s = strings.ReplaceAll(s, "software engineer", "dev")
	s = strings.ReplaceAll(s, "engenheiro de software", "dev")
	s = strings.ReplaceAll(s, "desenvolvedor", "dev")
	s = strings.ReplaceAll(s, "desenvolvedora", "dev")
	s = strings.ReplaceAll(s, "programador", "dev")
	s = strings.ReplaceAll(s, "junior", "jr")
	s = strings.ReplaceAll(s, "estagiario", "estagio")
	s = strings.ReplaceAll(s, "intern", "estagio")

	re := regexp.MustCompile(`[^a-z0-9]`)
	return re.ReplaceAllString(s, "")
}

func normalizarEmpresa(e string) string {
	e = removerAcentos(strings.ToLower(e))
	e = strings.ReplaceAll(e, "s.a.", "")
	e = strings.ReplaceAll(e, "ltda", "")
	e = strings.ReplaceAll(e, "inc.", "")
	e = strings.ReplaceAll(e, "tecnologia", "")
	e = strings.ReplaceAll(e, "brasil", "")

	re := regexp.MustCompile(`[^a-z0-9]`)
	return re.ReplaceAllString(e, "")
}

func removerAcentos(s string) string {
	return replacerAcentos.Replace(strings.ToLower(s))
}
