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
5. **Copie esse código** de 16 caracteres (exemplo: `abcdefghijklmnop` — sem espaços).

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
Quando você for rodar no GitHub Actions:

1. **Permissão de gravação para o histórico**:
   * No seu repositório no GitHub, vá em **Settings** -> **Actions** -> **General**.
   * Role até a seção **Workflow permissions**.
   * Marque a opção **Read and write permissions** e clique em **Save**. *(Isso permite ao robô salvar o `vagas_vistas.json` para não reenviar vagas repetidas).*

2. **Segredos do Repositório (Secrets)**:
   * Vá em **Settings** -> **Secrets and variables** -> **Actions**.
   * Clique em **New repository secret** e cadastre:
     * **`EMAIL_REMETENTE`**: seu endereço Gmail completo (ex: `seu.email@gmail.com`).
     * **`EMAIL_SENHA_APP`**: a senha de app de 16 letras gerada no Google (o código agora remove espaços automaticamente, mas prefira salvar sem espaços: `abcdefghijklmnop`).
     * **`EMAIL_DESTINATARIO`**: o endereço onde você quer receber o e-mail de alerta.

---

> [!NOTE]
> Essa senha de app pode ser revogada por você a qualquer momento na mesma página do Google caso queira interromper os envios.
