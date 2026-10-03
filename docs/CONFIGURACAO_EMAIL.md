# 🔐 Guia: Configuração de Senha de App do Gmail

Para que o robô em Go consiga enviar e-mails automatizados através da sua conta Google/Gmail de forma segura, você **não** usa a sua senha normal de login. Em vez disso, o Google fornece uma **Senha de App** (*App Password*), que é um token único de 16 letras exclusivo para scripts e automações.

---

### Passo a Passo

#### 1. Verifique se a Verificação em Duas Etapas está ativa
1. Acesse sua conta Google em: [myaccount.google.com](https://myaccount.google.com/).
2. No menu lateral, clique em **Segurança** (*Security*).
3. Na seção *"Como você faz login no Google"*, certifique-se de que a **Verificação em duas etapas** está **Ativada**.
   *(O Google só permite criar senhas de app se a verificação em duas etapas estiver ligada).*

#### 2. Gerar a Senha de App
1. Na barra de pesquisa do topo da sua Conta Google, digite: **Senhas de app** (ou acesse diretamente: [myaccount.google.com/apppasswords](https://myaccount.google.com/apppasswords)).
2. Dê um nome para identificar a aplicação (exemplo: `VagasScraper`).
3. Clique em **Criar** (*Create*).
4. O Google exibirá um código amarelo de 16 letras (exemplo: `abcd efgh ijkl mnop`).
5. **Copie esse código** (você pode usá-lo com ou sem os espaços).

---

### Onde colocar a Senha de App?

#### A) Para Testes no seu Computador (Local):
Crie um arquivo `.env` na raiz do projeto `vagas-scraper/` (copiando do `.env.example`):

```env
EMAIL_REMETENTE=seu.email@gmail.com
EMAIL_SENHA_APP=abcdefghijklmnop
EMAIL_DESTINATARIO=seu.email@gmail.com
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
DRY_RUN=false
```

#### B) No GitHub Actions (Nuvem):
Quando você for subir para o GitHub:
1. No seu repositório, vá em **Settings** -> **Secrets and variables** -> **Actions**.
2. Clique em **New repository secret**.
3. Adicione:
   * **`EMAIL_REMETENTE`**: seu endereço Gmail.
   * **`EMAIL_SENHA_APP`**: o código de 16 caracteres gerado pelo Google.
   * **`EMAIL_DESTINATARIO`**: o endereço onde você quer receber o e-mail de alerta.

---

> [!NOTE]
> Essa senha de app pode ser revogada por você a qualquer momento na mesma página do Google caso queira interromper os envios.
