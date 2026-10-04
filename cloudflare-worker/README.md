# Botão "Já vi esta vaga" (Cloudflare Worker)

O e-mail traz um link assinado (HMAC) → Worker valida → grava em `vagas_marcadas.json` no GitHub → o scraper nunca mais envia essa vaga.

## Deploy
1. Gere um **fine-grained token** no GitHub (Settings > Developer settings): acesso só ao repo `vagas-scraper`, permissão **Contents: Read and write**.
2. Escolha uma string aleatória longa para `MARCAR_SECRET`.
3. Na pasta `cloudflare-worker/`:
   ```bash
   npx wrangler login
   npx wrangler secret put GITHUB_TOKEN
   npx wrangler secret put MARCAR_SECRET
   npx wrangler deploy
   ```
   Anote a URL (`https://vagas-marcar.<conta>.workers.dev`).
4. No GitHub: **Settings > Secrets and variables > Actions**, crie:
   * `MARCAR_URL` = URL do Worker
   * `MARCAR_SECRET` = mesma string do passo 2
5. (Opcional, uso local) adicione `MARCAR_URL` e `MARCAR_SECRET` ao `.env`.

Sem `MARCAR_URL`/`MARCAR_SECRET` o botão simplesmente não aparece no e-mail.
