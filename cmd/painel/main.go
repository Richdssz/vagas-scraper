package main

import (
	"flag"
	"fmt"
	"log"
	"vagas-scraper/internal/web"
)

const versao = "1.2.0"

func main() {
	porta := flag.String("port", "8080", "Porta para o painel web local")
	caminhoConfig := flag.String("config", "config.json", "Caminho do arquivo de configuração JSON")
	exibirVersao := flag.Bool("version", false, "Exibe a versão do painel")
	flag.Parse()

	if *exibirVersao {
		fmt.Printf("Vagas Scraper • Painel de Controle v%s\n", versao)
		return
	}

	log.Printf("Iniciando Painel Visual de Configurações (v%s)...", versao)
	log.Printf("Servidor abrindo na porta %s...", *porta)

	servidor := web.NovoServidorWeb(*porta, *caminhoConfig)
	if err := servidor.Iniciar(); err != nil {
		log.Fatalf("Erro ao iniciar servidor do painel: %v", err)
	}
}
