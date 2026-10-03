package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

// ItemHistorico armazena metadados de uma vaga já vista
type ItemHistorico struct {
	ID        string    `json:"id"`
	Titulo    string    `json:"titulo,omitempty"`
	Capturado time.Time `json:"capturado"`
}

// Storage gerencia o histórico de vagas para evitar duplicatas
type Storage struct {
	caminho string
	mu      sync.RWMutex
	vistas  map[string]ItemHistorico
}

// NovoStorage inicializa o leitor/gravador de histórico
func NovoStorage(caminho string) (*Storage, error) {
	s := &Storage{
		caminho: caminho,
		vistas:  make(map[string]ItemHistorico),
	}

	if err := s.carregar(); err != nil {
		return nil, err
	}

	return s, nil
}

// carregar lê o arquivo do disco
func (s *Storage) carregar() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.caminho)
	if err != nil {
		if os.IsNotExist(err) {
			// Cria arquivo vazio se não existir
			return s.salvarEmDisco()
		}
		return fmt.Errorf("erro ao abrir histórico %s: %w", s.caminho, err)
	}

	if len(data) == 0 {
		return nil
	}

	// Tenta decodificar como lista de ItemHistorico
	var listaComplexa []ItemHistorico
	if err := json.Unmarshal(data, &listaComplexa); err == nil && len(listaComplexa) > 0 {
		for _, item := range listaComplexa {
			s.vistas[item.ID] = item
		}
		return nil
	}

	// Tenta decodificar como lista simples de strings ["id1", "id2"]
	var listaSimples []string
	if err := json.Unmarshal(data, &listaSimples); err == nil {
		for _, id := range listaSimples {
			s.vistas[id] = ItemHistorico{
				ID:        id,
				Capturado: time.Now(),
			}
		}
		return nil
	}

	return nil
}

// JaVista verifica se o ID da vaga já foi notificado anteriormente
func (s *Storage) JaVista(id string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, existe := s.vistas[id]
	return existe
}

// Adicionar adiciona múltiplos IDs ao histórico
func (s *Storage) Adicionar(idsComTitulo map[string]string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	agora := time.Now()
	for id, titulo := range idsComTitulo {
		s.vistas[id] = ItemHistorico{
			ID:        id,
			Titulo:    titulo,
			Capturado: agora,
		}
	}

	return s.salvarEmDisco()
}

// salvarEmDisco grava o mapa em formato JSON indentado
func (s *Storage) salvarEmDisco() error {
	lista := make([]ItemHistorico, 0, len(s.vistas))
	for _, item := range s.vistas {
		lista = append(lista, item)
	}

	data, err := json.MarshalIndent(lista, "", "  ")
	if err != nil {
		return fmt.Errorf("erro ao formatar histórico: %w", err)
	}

	if err := os.WriteFile(s.caminho, data, 0644); err != nil {
		return fmt.Errorf("erro ao gravar histórico em %s: %w", s.caminho, err)
	}

	return nil
}

// TotalVistas retorna a quantidade de vagas registradas
func (s *Storage) TotalVistas() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.vistas)
}
