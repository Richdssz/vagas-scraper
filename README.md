# ⚡ Vagas Scraper (Go) • Binários Separados & Painel Web

> Rastreador de vagas multicanal ultraleve dividido em **dois executáveis dedicados**: um para configuração visual (`painel.exe`) e outro para execução autônoma em segundo plano (`scraper.exe`).

---

## 📦 Os Dois Executáveis Separados

| Executável | Função Principal | Como roda? |
| :--- | :--- | :--- |
| 🎛️ **`painel.exe`** | **Painel de Controle Visual:** Abre o navegador em `http://localhost:8080` para você adicionar/remover termos, ligar/desligar portais, testar e configurar o agendador. | Dois cliques quando você quiser alterar configurações ou ver vagas na tela. |
| ⚡ **`scraper.exe`** | **Motor de Varredura Silencioso:** Varre 9 portais, deduplica, dispara o e-mail e desliga imediatamente em menos de 1.5s. | Chamado 3x ao dia pelo **Agendador do Windows**, **Termux** ou **GitHub Actions**. |

---

## 🎯 9 Portais de Vagas Concorrentes

1. 💼 **LinkedIn** (API Pública de Visitantes - sem login)
2. 🏢 **Gupy** (Portal Oficial de RH)
3. 🐙 **Backend-BR** (Mural do GitHub)
4. 💻 **Frontend-BR** (Mural do GitHub)
5. ⚛️ **React-Brasil** (Mural do GitHub)
6. 🧪 **QA-Brasil** (Mural do GitHub)
7. 🚀 **ProgramaThor** (Startups e Tech BR)
8. 🌐 **RemoteOK** (Vagas Remotas Globais)
9. ☕ **WeWorkRemotely** (Feed RSS de Programação)

---

## 🚀 Como Usar

### 1. Para Abrir o Painel de Configurações
Dê 2 cliques em **`painel.exe`** (ou execute no terminal):
```powershell
.\painel.exe
```
O navegador abrirá automaticamente em **`http://localhost:8080`**.

### 2. Para Executar a Varredura Manualmente em Segundo Plano
Dê 2 cliques em **`scraper.exe`** (ou execute no terminal):
```powershell
.\scraper.exe
```
*(Ou teste em modo simulação com `.\scraper.exe --dry-run`).*

---

## 🔨 Como Recompilar Ambos (Se Alterar Código)

```powershell
# Compila o painel visual
go build -ldflags="-s -w" -o painel.exe ./cmd/painel

# Compila o motor silencioso
go build -ldflags="-s -w" -o scraper.exe ./cmd/scraper

# Compila para o Galaxy S10+ (Termux / Linux ARM64)
$env:GOOS="linux"; $env:GOARCH="arm64"; go build -ldflags="-s -w" -o scraper-android ./cmd/scraper
```

---

## ⏰ Automação (3x ao Dia)

- **No Windows:** Abra o `painel.exe` e clique no botão verde **"⚡ Registrar no Agendador do Windows (1 Clique)"**. Ele registrará o `scraper.exe` para rodar às 09:00, 14:00 e 19:00.
- **No S10+ (Termux):** Configure no `crontab -e`:
  ```cron
  0 9,14,19 * * * /data/data/com.termux/files/home/scraper-android
  ```
- **No GitHub Actions:** O workflow [`.github/workflows/scraper.yml`](.github/workflows/scraper.yml) já está pronto para rodar o `./cmd/scraper`.
