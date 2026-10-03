package models

import "time"

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
