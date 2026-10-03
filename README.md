# ⚡ Vagas Scraper (Go) • Painel Web & Automação Híbrida

> Rastreador de vagas multicanal ultraleve com **Painel Visual Web Local (Porta 8080)**, **Deduplicação Cruzada Inteligente** (unifica a mesma vaga encontrada em portais diferentes) e suporte a execução automática no **Agendador de Tarefas do Windows** ou no **GitHub Actions**.

---

## 🎯 Destaques do Projeto

- **Painel Visual no Navegador:** Basta rodar `go run main.go` ou clicar duas vezes no `scraper.exe` para abrir a interface gráfica em `http://localhost:8080`.
- **Deduplicação Cruzada Inteligente:** Se a mesma vaga for anunciada no **LinkedIn**, na **Gupy** e no **GitHub**, o robô gera **um único card** com botões para todos os portais onde ela foi encontrada!
- **9 Portais de Vagas Concorrentes:**
  1. 💼 **LinkedIn** (API Pública de Visitantes - sem necessidade de login)
  2. 🏢 **Gupy** (Portal Oficial de RH)
  3. 🐙 **Backend-BR** (Mural do GitHub)
  4. 💻 **Frontend-BR** (Mural do GitHub)
  5. ⚛️ **React-Brasil** (Mural do GitHub)
  6. 🧪 **QA-Brasil** (Mural do GitHub)
  7. 🚀 **ProgramaThor** (Startups e Tech BR)
  8. 🌐 **RemoteOK** (Vagas Remotas Globais)
  9. ☕ **WeWorkRemotely** (Feed RSS de Programação)
- **Ultraleve:** O executável completo (`scraper.exe`) tem apenas ~11 MB e roda em menos de 1.5 segundo.
- **Dois Modos de Automação:**
  - **Agendador do Windows (Local):** Registre com 1 clique no painel para rodar 3x ao dia (09h, 14h, 19h) silenciosamente em segundo plano.
  - **GitHub Actions (Nuvem):** Agendado no `.github/workflows/scraper.yml` para rodar na nuvem do GitHub sem precisar do computador ligado.

---

## 🚀 Como Iniciar o Painel Visual

No terminal dentro da pasta `vagas-scraper`:

```powershell
go run main.go
```
*(Ou dê 2 cliques no `scraper.exe`).*

O navegador abrirá automaticamente em:
👉 **`http://localhost:8080`**

Na interface você pode:
1. **Adicionar/remover termos desejados** (ex: `golang`, `estágio`, `junior`, `react`).
2. **Adicionar/remover termos proibidos** (ex: `senior`, `pleno`, `lead`).
3. **Marcar quais portais de vagas consultar**.
4. **Configurar seu e-mail do Gmail** para alertas.
5. **Clicar em "Buscar Vagas Agora"** para ver os resultados ao vivo na tela com os links diretos de cada plataforma.
6. **Clicar em "Registrar no Agendador do Windows"** para ativar o ciclo diário automático.

---

## 🛠️ Comandos Rápidos

| Comando | O que faz |
| :--- | :--- |
| `go run main.go` | Inicia o servidor web na porta 8080 e abre o painel no navegador. |
| `go run main.go --port 3000` | Inicia o servidor web em uma porta personalizada. |
| `go run main.go --scrape` | Executa a varredura silenciosa e encerra imediatamente (modo usado pelo agendador). |
| `go run main.go --scrape --dry-run` | Varre as vagas sem disparar e-mails reais e gera `preview_email.html`. |
| `go run main.go --limpar-historico` | Reseta a lista de vagas já vistas em `vagas_vistas.json`. |

---

## ☁️ Ativação no GitHub Actions

Se preferir rodar 100% na nuvem sem deixar a máquina ligada:
1. Suba este repositório para o seu GitHub quando quiser.
2. Em **Settings -> Secrets and variables -> Actions**, adicione:
   - `EMAIL_REMETENTE`
   - `EMAIL_SENHA_APP`
   - `EMAIL_DESTINATARIO`
3. O workflow [`.github/workflows/scraper.yml`](.github/workflows/scraper.yml) já está pronto e programado para rodar às 09:00, 14:00 e 19:00 (BRT).
