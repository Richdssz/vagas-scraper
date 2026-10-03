# ⏰ Guia de Agendamento: Expressões Cron & Fusos Horários

O GitHub Actions utiliza expressões **cron** padrão de 5 campos para definir a frequência com que o robô de vagas executa.

Arquivo responsável: `.github/workflows/scraper.yml`

```yaml
on:
  schedule:
    - cron: 'MINUTO HORA DIA_MES MES DIA_SEMANA'
```

---

### Atenção ao Fuso Horário (UTC vs Horário de Brasília)

> [!IMPORTANT]
> Os servidores do GitHub rodam em **UTC** (Tempo Universal Coordenado).
> O Horário de Brasília (BRT) está **3 horas atrás do UTC** (UTC - 3).
> 
> *Para rodar às **09:00** em Brasília, configure para as **12:00 UTC** no cron.*

---

### Tabela de Frequências Prontas para Usar

Copie a linha desejada e cole no seu `.github/workflows/scraper.yml`:

| Frequência Desejada | Horários em Brasília (BRT) | Sintaxe Cron no GitHub (UTC) |
| :--- | :--- | :--- |
| **1x ao dia (Manhã)** | 09:00 | `- cron: '0 12 * * *'` |
| **1x ao dia (Noite)** | 20:00 | `- cron: '0 23 * * *'` |
| **2x ao dia** | 09:00 e 18:00 | `- cron: '0 12,21 * * *'` |
| **3x ao dia (Recomendado)** | 09:00, 14:00 e 19:00 | `- cron: '0 12,17,22 * * *'` |
| **4x ao dia** | 08:00, 12:00, 16:00 e 20:00 | `- cron: '0 11,15,19,23 * * *'` |
| **A cada 3 horas** | A cada 3 horas (dia todo) | `- cron: '0 */3 * * *'` |
| **Apenas Dias Úteis (Seg-Sex)** | 09:00, 14:00 e 19:00 | `- cron: '0 12,17,22 * * 1-5'` |

---

### Como alterar direto no GitHub?

1. No repositório, navegue até `.github/workflows/scraper.yml`.
2. Clique no ícone de lápis ✏️ (**Edit this file**).
3. Modifique a linha do `- cron: '...'`.
4. Clique em **Commit changes...** para salvar. Pronto! A nova frequência entrará em vigor imediatamente.
