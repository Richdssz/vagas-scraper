package models

import (
	"fmt"
	"time"
)

// LinkFonte representa um link para a vaga em um portal específico
type LinkFonte struct {
	Fonte string `json:"fonte"`
	URL   string `json:"url"`
}

// Vaga representa uma oportunidade de emprego capturada de qualquer fonte,
// com suporte a deduplicação cruzada (múltiplas fontes para a mesma vaga).
type Vaga struct {
	ID            string      `json:"id"`
	Titulo        string      `json:"titulo"`
	Empresa       string      `json:"empresa"`
	Localizacao   string      `json:"localizacao"`
	Link          string      `json:"link"`          // Link principal
	Links         []LinkFonte `json:"links"`         // Todos os links (ex: LinkedIn + Gupy)
	Fontes        []string    `json:"fontes"`        // Nomes das fontes onde apareceu
	Fonte         string      `json:"fonte"`         // Fonte primária
	Tags          []string    `json:"tags,omitempty"`
	Data          time.Time   `json:"data"`
	ChaveCanonica string      `json:"chave_canonica"` // Fingerprint empresa+cargo
}

// TempoRelativo retorna uma descrição legível de quando a vaga foi publicada
// (ex: "Há 2 horas", "Hoje", "Ontem", "Há 3 dias").
func (v Vaga) TempoRelativo() string {
	if v.Data.IsZero() {
		return "Recente"
	}
	diff := time.Since(v.Data)
	if diff < 0 {
		return "Hoje"
	}
	if diff < 1*time.Minute {
		return "Agora há pouco"
	}
	if diff < 1*time.Hour {
		min := int(diff.Minutes())
		if min <= 1 {
			return "Há 1 min"
		}
		return fmt.Sprintf("Há %d min", min)
	}
	if diff < 24*time.Hour {
		h := int(diff.Hours())
		if h <= 1 {
			return "Há 1 hora"
		}
		return fmt.Sprintf("Há %d horas", h)
	}
	dias := int(diff.Hours() / 24)
	if dias == 1 {
		return "Ontem"
	}
	if dias < 7 {
		return fmt.Sprintf("Há %d dias", dias)
	}
	if dias < 30 {
		sem := dias / 7
		if sem == 1 {
			return "Há 1 semana"
		}
		return fmt.Sprintf("Há %d semanas", sem)
	}
	return v.Data.Format("02/01/2006")
}
