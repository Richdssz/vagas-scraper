package models

import "time"

// Vaga representa uma oportunidade de emprego capturada de qualquer fonte.
type Vaga struct {
	ID          string    `json:"id"`
	Titulo      string    `json:"titulo"`
	Empresa     string    `json:"empresa"`
	Localizacao string    `json:"localizacao"`
	Link        string    `json:"link"`
	Fonte       string    `json:"fonte"`
	Tags        []string  `json:"tags,omitempty"`
	Data        time.Time `json:"data"`
}
