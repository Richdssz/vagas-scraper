# ⚡ Vagas Scraper (Go)

> Rastreador de vagas ultraleve e automatizado desenvolvido em **Go**, projetado para rodar até várias vezes ao dia e notificar você por **e-mail com layout moderno em HTML**.

---

## 🎯 Por que em Go?

- **Ultraleve:** Consome apenas **~8MB de memória RAM** durante a execução.
- **Veloz:** Faz varredura concorrente de dezenas de vagas em **menos de 1 segundo**.
- **Zero Interpretadores:** Diferente do Python ou Node.js, compila para um único binário estático independente.
- **Anti-Duplicatas Confiável:** Mantém um registro inteligente (`vagas_vistas.json`) para nunca enviar a mesma vaga duas vezes.
- **Multi-Fonte Concorrente:** Consulta repositórios nacionais e plataformas remotas globais em paralelo usando *Goroutines*.

---

## 📂 Fontes de Vagas Integradas

1. **Backend-BR (`backend-br/vagas`):** O maior feed de vagas tech no GitHub do Brasil (Junior, Estágio, Pleno, Remoto).
2. **Frontend-BR (`frontendbr/vagas`):** Feed de oportunidades front-end e mobile da comunidade brasileira.
3. **RemoteOK:** Feed internacional de oportunidades 100% remotas globais.
4. **Módulo HTML Genérico (`goquery`):** Estrutura pronta para raspar qualquer portal com seletores CSS.

---

## 📁 Estrutura do Repositório

```text
vagas-scraper/
├── .github/
│   └── workflows/
│       └── scraper.yml            # Workflow para execução automatizada no GitHub
├── internal/
│   ├── config/
│   │   └── config.go              # Leitor de config.json e variáveis de ambiente
│   ├── models/
│   │   └── vaga.go                # Estrutura de dados da Vaga
│   ├── notifier/
│   │   ├── email.go               # Envio via SMTP seguro (STARTTLS)
│   │   └── template.go            # Template HTML responsivo moderno
│   ├── scrapers/
│   │   ├── scraper.go             # Interface padrão e filtros de palavras
│   │   ├── github_vagas.go        # Scraper dos repositórios GitHub BR
│   │   ├── remoteok.go            # Scraper do RemoteOK
│   │   └── html_scraper.go        # Scraper genérico com goquery (CSS selectors)
│   └── storage/
│       └── storage.go             # Gerenciador de histórico e deduplicação
├── docs/
│   ├── CONFIGURACAO_EMAIL.md      # Passo a passo da Senha de App do Gmail
│   └── AGENDAMENTO_CRON.md        # Tabela de horários e frequências do cron
├── config.json                    # Palavras-chave, exclusões e fontes ativas
├── vagas_vistas.json              # Registro de vagas já enviadas
├── .env.example                   # Exemplo de configuração de credenciais
├── .gitignore
├── go.mod
├── go.sum
├── main.go                        # Ponto de entrada da aplicação
└── README.md
```

---

## ⚙️ Configuração dos Filtros (`config.json`)

Você pode personalizar os termos que procura e o que quer ignorar diretamente no arquivo `config.json`:

```json
{
  "filtros": {
    "termos_busca": [
      "estagio", "estágio", "junior", "júnior", "trainee",
      "backend", "golang", "go", "java", "node", "fullstack", "react"
    ],
    "termos_exclusao": [
      "senior", "sênior", "sr.", "pleno", "tech lead", "especialista"
    ]
  },
  "fontes_habilitadas": [
    "backend_br",
    "frontend_br",
    "remoteok"
  ],
  "max_vagas_por_execucao": 25,
  "salvar_preview_html": true
}
```

> **Dica:** Os `termos_exclusao` evitam que vagas que você não quer (como *Sênior* ou *Tech Lead*) cheguem na sua caixa de entrada, mesmo que citem tecnologias como Go ou Java no texto.

---

## 🚀 Como Executar Localmente

### 1. Teste Rápido (Modo Simulação / Sem Enviar E-mails)
Você pode testar a busca e conferir o resultado no terminal sem precisar de credenciais de e-mail:

```powershell
go run main.go --dry-run
```

* Quando você roda em modo simulação, o script gera automaticamente um arquivo **`preview_email.html`** na raiz. Basta dar um duplo clique nele para abrir no navegador e ver como o e-mail fica bonito!

---

### 2. Configurando o Envio de E-mails Reais

1. Copie o arquivo de exemplo de ambiente:
   ```powershell
   Copy-Item .env.example .env
   ```
2. Abra o `.env` e preencha suas informações:
   ```env
   EMAIL_REMETENTE=seu.email@gmail.com
   EMAIL_SENHA_APP=xxxx xxxx xxxx xxxx
   EMAIL_DESTINATARIO=seu.email@gmail.com
   SMTP_HOST=smtp.gmail.com
   SMTP_PORT=587
   DRY_RUN=false
   ```
   *(Caso não saiba como gerar a Senha de App do Gmail, consulte o [docs/CONFIGURACAO_EMAIL.md](docs/CONFIGURACAO_EMAIL.md)).*

3. Execute para valer:
   ```powershell
   go run main.go
   ```

---

### 3. Compilando para um Executável Único (Opcional)

Se quiser gerar um `.exe` que roda sozinho sem precisar do comando `go`:

```powershell
go build -ldflags="-s -w" -o scraper.exe
```

Para rodar:
```powershell
.\scraper.exe --dry-run
```

---

## ☁️ Como Ativar no GitHub Actions (3x ao Dia)

Quando você decidir colocar o repositório no GitHub para rodar na nuvem sem precisar do computador ligado:

1. **Crie um repositório** no seu GitHub.
2. Adicione os **Secrets** em `Settings -> Secrets and variables -> Actions`:
   - `EMAIL_REMETENTE`
   - `EMAIL_SENHA_APP`
   - `EMAIL_DESTINATARIO`
3. Habilite a permissão de escrita em `Settings -> Actions -> General -> Workflow permissions`:
   - Selecione **"Read and write permissions"** e salve.
4. O GitHub Actions já está configurado no arquivo [`.github/workflows/scraper.yml`](.github/workflows/scraper.yml) para rodar automaticamente **3 vezes ao dia** (às 09:00, 14:00 e 19:00 no Horário de Brasília).
5. Para mudar a frequência, consulte o [docs/AGENDAMENTO_CRON.md](docs/AGENDAMENTO_CRON.md).

---

## 🛠️ Comandos Úteis

| Comando | Descrição |
| :--- | :--- |
| `go run main.go --dry-run` | Executa busca sem disparar e-mails e gera o preview HTML. |
| `go run main.go --limpar-historico` | Limpa o histórico de vagas salvas em `vagas_vistas.json`. |
| `go run main.go --config outro.json` | Roda utilizando um arquivo de configuração alternativo. |
| `go run main.go --version` | Exibe a versão instalada do scraper. |
