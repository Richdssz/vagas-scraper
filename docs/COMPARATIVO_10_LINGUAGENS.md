# 📊 Comparativo de 10 Linguagens para o Web Scraper
## Avaliação sob a ótica de uma Pessoa Leiga (Facilidade, Praticidade e Confiabilidade)

---

## 🎯 Dá para fazer em Dart?

**Sim, perfeitamente!**  
O Dart (a linguagem por trás do Flutter) não serve só para apps de celular. Ele possui um compilador nativo excelente chamado `dart compile exe`. 

* Com um único comando (`dart compile exe bin/main.dart -o scraper.exe`), o Dart gera um executável `.exe` independente.
* Ele possui as bibliotecas `package:http`, `package:html` (similar ao BeautifulSoup) e `package:mailer`.
* Para quem já programa em Flutter ou Java/C#, a sintaxe do Dart é extremamente agradável.

---

## 🧭 Tabela Comparativa: 10 Linguagens

Abaixo, a comparação de como seria criar e manter este robô de vagas em 10 tecnologias diferentes, focando na **experiência de uso de uma pessoa que não é programadora**:

| # | Linguagem | Precisa instalar algo no PC pra rodar? | Como o leigo executa? | Tamanho do Arquivo | Memória RAM | Risco de quebrar sozinho | Nota pro Leigo (1-10) |
|:-:|:---|:---:|:---:|:---:|:---:|:---:|:---:|
| 1 | **Go (Golang)** | **NÃO** (Compilação estática) | 2 cliques no `.exe` | ~5 a 8 MB | ~8 MB | Praticamente Zero | **9.8** |
| 2 | **Dart** | **NÃO** (`dart compile exe`) | 2 cliques no `.exe` | ~15 a 20 MB | ~20 MB | Baixo | **9.0** |
| 3 | **Bun (JS/TS)** | **NÃO** (`bun build --compile`) | 2 cliques no `.exe` | ~50 MB | ~25 MB | Baixo | **8.8** |
| 4 | **C# (.NET)** | **NÃO** (`SingleFile Publish`) | 2 cliques no `.exe` | ~60 MB | ~40 MB | Baixo | **8.5** |
| 5 | **Rust** | **NÃO** (`cargo build --release`) | 2 cliques no `.exe` | ~4 MB | ~4 MB | Zero | **7.5** *(difícil de programar)* |
| 6 | **Python** | **SIM** (Python runtime + pip) | Terminal com `python script.py` | Leve, mas disperso | ~60 MB | Alto (conflito de pacotes) | **5.5** |
| 7 | **Node.js** | **SIM** (Node runtime + node_modules) | Terminal com `npm start` | Pesado (`node_modules`) | ~100 MB | Médio/Alto | **5.0** |
| 8 | **Java** | **SIM** (JVM / Java instalado) | Terminal com `java -jar` | Médio (`.jar` de ~15MB) | ~120 MB | Baixo | **4.5** |
| 9 | **PHP** | **SIM** (PHP CLI + extensões) | Terminal com `php script.php` | Leve | ~35 MB | Médio | **4.0** |
| 10| **PowerShell** | **NÃO** (Já vem no Windows) | Botão direito -> Executar | Apenas 1 script `.ps1` | ~30 MB | Médio (HTML complexo) | **6.0** |

---

## 🔍 Análise Detalhada Linguagem por Linguagem

---

### 1. Go (Golang) — *O Campeão de Eficiência*
* **Como seria:** Foi a arquitetura que criamos no repositório. Uso de goroutines e compilador estático.
* **O que o leigo ganha:** Você gera o `scraper.exe` de apenas 7 MB e manda para qualquer pessoa. Ela clica duas vezes e roda. Não precisa instalar nada, não precisa configurar "variáveis de ambiente do Windows", não precisa de tela preta se não quiser.
* **O que perde:** A escrita do código é mais direta e sem "mágica", exigindo tratar erros manualmente linha por linha durante o desenvolvimento.

---

### 2. Dart — *O Irmão Moderno do Flutter*
* **Como seria:**
  ```dart
  import 'package:http/http.dart' as http;
  import 'package:html/parser.dart' show parse;

  void main() async {
    final res = await http.get(Uri.parse('https://exemplo.com/vagas'));
    final doc = parse(res.body);
    // busca cards e dispara e-mail com package:mailer
  }
  ```
* **O que o leigo ganha:** Uma experiência excelente. O Dart compila nativamente para `.exe` sem precisar da máquina virtual. O código é legível, moderno e com suporte nativo a `async/await`.
* **O que perde:** O `.exe` do Dart fica um pouco maior (~18MB) que o de Go (~7MB), e o ecossistema de raspagem HTML em Dart é bem menor que em Python ou Go.

---

### 3. Bun (JavaScript / TypeScript Moderno)
* **Como seria:** É a versão moderna do Node.js. Permite compilar código JS ou TS direto para um binário autônomo com `bun build --compile scraper.ts --outfile scraper.exe`.
* **O que o leigo ganha:** Não precisa da pasta gigante `node_modules` nem de instalar o Node.js. O usuário leigo só recebe o `.exe`.
* **O que perde:** O arquivo `.exe` resultante fica em torno de 50MB a 60MB porque embute a engine do JavaScript dentro dele.

---

### 4. C# (.NET) — *O Favorito do Ambiente Windows*
* **Como seria:** Usando `HttpClient`, `HtmlAgilityPack` e `MailKit`. Com o comando `dotnet publish -r win-x64 -p:PublishSingleFile=true`, gera um único executável.
* **O que o leigo ganha:** Integração perfeita com o Windows e tarefas agendadas nativas. Facilidade total para rodar com dois cliques.
* **O que perde:** O executável inicial do .NET ultrapassa 60 MB facilmente para carregar a biblioteca base do sistema.

---

### 5. Rust — *A Máxima Performance Possível*
* **Como seria:** Usando `reqwest`, `scraper` e `lettre`. Compilação com `cargo build --release`.
* **O que o leigo ganha:** O menor consumo de memória de todos (cerca de 3MB de RAM) e um executável pequeno de ~4MB à prova de qualquer falha de memória.
* **O que perde:** É a linguagem mais complexa para um desenvolvedor manter ou para uma pessoa leiga fazer qualquer pequena alteração nas regras de busca.

---

### 6. Python — *O Mais Famoso, mas Frágil para Leigos*
* **Como seria:** O tradicional `requests` + `BeautifulSoup` + `smtplib`.
* **O que o leigo ganha:** O código é fácil de ler como se fosse inglês. Se alguém quiser mudar uma palavra-chave direto no código, é intuitivo.
* **O que o leigo sofre (Pesadelo):**
  * Precisa instalar o Python e marcar "Add Python to PATH" (muitos esquecem e nada funciona).
  * Precisa abrir o terminal, criar `venv`, rodar `pip install -r requirements.txt`.
  * Se o Python do Windows atualizar ou a versão de uma biblioteca mudar, o script para de funcionar do nada com erros vermelhos no terminal.

---

### 7. Node.js (JavaScript Tradicional)
* **Como seria:** Usando `axios`, `cheerio` e `nodemailer`.
* **O que o leigo ganha:** Se a pessoa já entende um mínimo de desenvolvimento web (HTML/JS), os seletores são idênticos aos do navegador.
* **O que o leigo sofre:** Precisa instalar o Node.js. A pasta do projeto fica cheia com a pasta `node_modules` (centenas de megabytes e milhares de pequenos arquivos), assustando quem não é da área.

---

### 8. Java — *Robusto, porém Burocrático*
* **Como seria:** Usando `HttpClient` (Java 11+), `JSoup` para raspar e `Jakarta Mail` para enviar.
* **O que o leigo ganha:** Código fortemente estruturado em classes.
* **O que o leigo sofre:** Precisa ter o Java (JRE/JDK) instalado na máquina. Clicar duas vezes num arquivo `.jar` muitas vezes não abre nada no Windows sem configuração prévia.

---

### 9. PHP — *O Clássico da Web*
* **Como seria:** Usando `file_get_contents` ou `cURL`, com `DOMDocument` e a função nativa `mail()`.
* **O que o leigo ganha:** Se for rodar numa hospedagem web barata (cPanel de R$ 10/mês), roda com zero configuração.
* **O que o leigo sofre:** Rodar PHP localmente no Windows requer baixar zips, configurar o `php.ini`, habilitar extensão OpenSSL/cURL manualmente no arquivo de texto, o que é muito confuso para iniciantes.

---

### 10. PowerShell (Nativo do Windows)
* **Como seria:** Um script simples `.ps1` usando `Invoke-RestMethod` e `Send-MailMessage`.
* **O que o leigo ganha:** **Não precisa instalar absolutamente nada.** O Windows já vem com PowerShell instalado de fábrica.
* **O que perde:** Por padrão, o Windows bloqueia a execução de scripts `.ps1` por segurança (erro de *Execution Policy*), e manipular páginas HTML complexas nele é bem trabalhoso.

---

## 🏆 Conclusão & Veredito para Uso Pessoal

Se o objetivo é **entregar para alguém usar sem dor de cabeça** ou rodar silenciosamente:

1. **Top 1:** **Go (Golang)** — Vence pelo equilíbrio absoluto: arquivo único de ~7MB, não pede instalação de nada, gasta quase zero de memória e roda em 0.5 segundo.
2. **Top 2:** **Dart** — Excelente alternativa se você gosta do ecossistema Flutter. Entrega a mesma facilidade de duplo clique que o Go.
3. **Top 3:** **Bun** — Melhor caminho se você quiser programar em TypeScript/JavaScript sem carregar o fardo do `node_modules`.
