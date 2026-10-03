# ⚡ Vagas Scraper (Go) • Manual Completo de Configuração

> Rastreador de vagas multicanal ultraleve com **Deduplicação Cruzada** (unifica vagas iguais entre LinkedIn, Gupy, GitHub e outros portais), **Painel Visual Web** e motor de execução autônomo em segundo plano.

---

## 📦 Como o Sistema é Dividido

O projeto conta com **dois executáveis independentes**:

```text
vagas-scraper/
├── 🎛️ painel.exe     # Abre a interface visual em http://localhost:8080 para configurar e testar
├── ⚡ scraper.exe    # Motor silencioso que roda e encerra em 1 segundo (usado nos agendadores)
├── ⚙️ config.json    # Suas palavras-chave, exclusões e portais ativos
├── 🔑 .env           # Suas credenciais de envio de e-mail (protegido pelo .gitignore)
└── 📦 vagas_vistas.json # Memória anti-duplicatas do robô
```

---

## 🛠️ Passo a Passo: Como Configurar o Sistema

Você pode configurar o sistema de **duas formas**: pelo **Painel Visual (Recomendado)** ou **manualmente por arquivos**.

---

### Opção A: Pelo Painel Visual (O Jeito Mais Fácil)

1. Dê dois cliques em **`painel.exe`** (ou rode `.\painel.exe` no terminal).
2. O seu navegador abrirá automaticamente em:
   👉 **`http://localhost:8080`**

3. **Configure seus Critérios de Pesquisa:**
   * **O que pesquisar:** Digite suas linguagens e níveis desejados (ex: `golang`, `estágio`, `junior`, `react`, `backend`) e tecle `Enter` para adicionar as tags azuis.
   * **O que NÃO pesquisar (Exclusões):** Digite o que você quer descartar (ex: `senior`, `sênior`, `pleno`, `lead`, `coordenador`) e tecle `Enter` para adicionar as tags vermelhas. Isso garante que vagas avançadas sejam eliminadas mesmo que citem tecnologias que você conhece.

4. **Selecione os Portais de Vagas:**
   * Marque ou desmarque as caixinhas dos portais desejados:
     * 💼 **LinkedIn:** API pública de visitantes (sem pedir login).
     * 🏢 **Gupy:** Vagas das maiores empresas do Brasil (Nubank, Stone, Ambev, etc.).
     * 🐙 **Backend-BR, Frontend-BR, React-Brasil, QA-Brasil:** Murais do GitHub da comunidade BR.
     * 🚀 **ProgramaThor:** Vagas em startups tech do Brasil.
     * 🌐 **RemoteOK & WeWorkRemotely:** Vagas remotas internacionais em dólar/euro.

5. **Configure o E-mail de Alerta:**
   * **Seu E-mail (Remetente):** Seu endereço Gmail (ex: `seu.email@gmail.com`).
   * **Senha de App do Gmail:** Código de 16 letras gerado no Google ([veja como gerar abaixo](#-como-gerar-a-senha-de-app-do-gmail)).
   * **E-mail de Destino:** Onde você quer receber os alertas com os botões de candidatura.

6. **Salve as Preferências:**
   * Clique no botão **"💾 Salvar Preferências"** no topo da página. O painel salvará tudo automaticamente no `config.json` e no `.env`.

7. **Faça um Teste ao Vivo:**
   * Clique no botão azul **"🔍 Buscar Vagas Agora"**.
   * O robô fará a varredura e exibirá as vagas na sua tela na hora, mostrando os botões para cada portal onde a vaga foi encontrada!

---

### Opção B: Configuração Manual por Arquivos

Se você preferir editar diretamente pelo VS Code ou Bloco de Notas:

#### 1. Editando o `config.json`:
Abra o arquivo [`config.json`](config.json) e personalize suas preferências:

```json
{
  "filtros": {
    "termos_busca": [
      "estagio", "estágio", "junior", "júnior", "trainee",
      "backend", "golang", "go", "java", "node", "fullstack", "react"
    ],
    "termos_exclusao": [
      "senior", "sênior", "sr.", "pleno", "tech lead", "especialista", "coordenador"
    ]
  },
  "fontes_habilitadas": [
    "linkedin",
    "gupy",
    "backend_br",
    "frontend_br",
    "programathor",
    "remoteok",
    "weworkremotely"
  ],
  "max_vagas_por_execucao": 30,
  "salvar_preview_html": true
}
```

#### 2. Criando o arquivo `.env`:
Copie o arquivo `.env.example` para `.env`:
```powershell
Copy-Item .env.example .env
```
Preencha com suas informações:
```env
EMAIL_REMETENTE=seu.email@gmail.com
EMAIL_SENHA_APP=abcd efgh ijkl mnop
EMAIL_DESTINATARIO=seu.email@gmail.com
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
DRY_RUN=false
```

---

## 🔑 Como Gerar a Senha de App do Gmail

Para sua segurança, o Google não aceita sua senha comum de login em scripts automatizados. Você precisa gerar uma senha de app exclusiva:

1. Acesse sua Conta Google em: [myaccount.google.com/security](https://myaccount.google.com/security).
2. Certifique-se de que a **Verificação em duas etapas** está **Ativada**.
3. Na barra de busca no topo da página, digite **"Senhas de app"** (ou acesse [myaccount.google.com/apppasswords](https://myaccount.google.com/apppasswords)).
4. Dê um nome para a aplicação (ex: `VagasScraper`) e clique em **Criar**.
5. O Google gerará um código de **16 letras** (exemplo: `abcd efgh ijkl mnop`).
6. Cole esse código no campo **Senha de App** do painel ou no arquivo `.env`.

*(Para instruções detalhadas com prints, consulte o guia dedicado em [docs/CONFIGURACAO_EMAIL.md](docs/CONFIGURACAO_EMAIL.md)).*

---

## ⏰ Como Ativar a Execução Automática (3x ao Dia)

Escolha onde você prefere que o robô rode de forma 100% autônoma:

### Método 1: No seu Windows (Agendador de Tarefas)
O Windows acorda o executável `scraper.exe` 3 vezes ao dia (às **09:00, 14:00 e 19:00**), varre todas as fontes em 1 segundo, envia o e-mail com as novidades e **desliga imediatamente** (0% de CPU no resto do dia).

* **Pelo Painel:** Abra o `painel.exe` e clique no botão verde **"⚡ Registrar no Agendador do Windows (1 Clique)"**.
* **Pelo PowerShell (Manual):**
  ```powershell
  $exe = (Get-Item .\scraper.exe).FullName
  $action = New-ScheduledTaskAction -Execute $exe
  $t1 = New-ScheduledTaskTrigger -Daily -At 9am
  $t2 = New-ScheduledTaskTrigger -Daily -At 2pm
  $t3 = New-ScheduledTaskTrigger -Daily -At 7pm
  Register-ScheduledTask -TaskName "VagasScraperDiario" -Action $action -Trigger $t1,$t2,$t3 -Description "Varredura automática de vagas" -Force
  ```

---

### Método 2: No seu Galaxy S10+ com Termux (Mini Servidor 24/7)
Transforme o S10+ em um servidor doméstico silencioso que nunca desliga:

1. Compile o executável para a arquitetura do celular (Linux ARM64):
   ```powershell
   $env:GOOS="linux"; $env:GOARCH="arm64"; go build -ldflags="-s -w" -o scraper-android ./cmd/scraper
   ```
2. Passe o arquivo `scraper-android`, o `config.json` e o `.env` para o celular.
3. No Termux, agende via cron (`crontab -e`):
   ```cron
   0 9,14,19 * * * /data/data/com.termux/files/home/scraper-android
   ```
*(Consulte o tutorial completo passo a passo em [docs-gerais/GUIA_S10_MINI_SERVER_TERMUX.md](../docs-gerais/GUIA_S10_MINI_SERVER_TERMUX.md)).*

---

### Método 3: No GitHub Actions (Nuvem Gratuita)
Rode nos servidores do GitHub sem precisar de nenhum computador ligado em casa:

1. Crie um repositório no seu GitHub e suba o projeto.
2. No repositório, vá em **Settings -> Secrets and variables -> Actions**.
3. Adicione os 3 segredos:
   * `EMAIL_REMETENTE`
   * `EMAIL_SENHA_APP`
   * `EMAIL_DESTINATARIO`
4. Em **Settings -> Actions -> General -> Workflow permissions**, marque **"Read and write permissions"** e salve.
5. O robô rodará automaticamente via [`.github/workflows/scraper.yml`](.github/workflows/scraper.yml) nos horários definidos no cron.

---

## 🧪 Comandos Úteis para o Dia a Dia

| Comando | O que faz |
| :--- | :--- |
| `.\painel.exe` | Abre o painel visual no navegador (`http://localhost:8080`). |
| `.\scraper.exe` | Executa uma varredura silenciosa e envia e-mails se houver vagas novas. |
| `.\scraper.exe --dry-run` | Executa a varredura em modo teste (gera `preview_email.html` e não manda e-mails reais). |
| `.\scraper.exe --limpar-historico` | Apaga a memória de vagas já vistas (`vagas_vistas.json`) para testar como se fosse a primeira vez. |
| `.\scraper.exe --version` | Exibe a versão instalada do motor de vagas. |

---

## 🧬 Como Funciona a Deduplicação Cruzada

Se a mesma vaga da empresa **Nubank** for publicada no **LinkedIn**, na **Gupy** e no **GitHub**, o robô:
1. Normaliza os títulos e empresas para uma chave canônica única (`nubank:devbackendjr`).
2. Identifica que se trata da mesmíssima oportunidade.
3. Junta tudo em um único cartão de e-mail com múltiplos botões:
   > 💼 **Desenvolvedor Backend Júnior (Nubank)**  
   > ⭐ *Disponível em 2 plataformas (LinkedIn, Gupy)*  
   > 🔗 `[ Ver no LinkedIn ]` &nbsp;&nbsp; 🔗 `[ Ver na Gupy ]`
