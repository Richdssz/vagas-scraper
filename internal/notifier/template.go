package notifier

import (
	"fmt"
	"strings"
	"time"
	"vagas-scraper/internal/models"
)

// GerarHTMLEmail produz um e-mail com visual moderno e suporte a múltiplos botões de candidatura
func GerarHTMLEmail(vagas []models.Vaga) string {
	var cardsHTML strings.Builder

	for _, v := range vagas {
		// Monta tags HTML
		var tagsHTML strings.Builder
		for _, tag := range v.Tags {
			if strings.TrimSpace(tag) == "" {
				continue
			}
			tagsHTML.WriteString(fmt.Sprintf(`
				<span style="display: inline-block; background-color: #e0f2fe; color: #0369a1; font-size: 11px; font-weight: 600; padding: 2px 8px; border-radius: 9999px; margin-right: 4px; margin-bottom: 4px;">
					%s
				</span>`, escapeHTML(tag)))
		}

		// Monta botões de links (deduplicação cruzada: se a mesma vaga estiver no LinkedIn e na Gupy, exibe ambos!)
		var botoesHTML strings.Builder
		if len(v.Links) > 0 {
			for _, lf := range v.Links {
				botoesHTML.WriteString(fmt.Sprintf(`
					<a href="%s" target="_blank" style="display: inline-block; background: #2563eb; color: #ffffff; text-decoration: none; font-size: 13px; font-weight: 600; padding: 8px 14px; border-radius: 6px; margin-right: 8px; margin-bottom: 6px;">
						🔗 Ver no %s &rarr;
					</a>`, escapeHTML(lf.URL), escapeHTML(lf.Fonte)))
			}
		} else if v.Link != "" {
			botoesHTML.WriteString(fmt.Sprintf(`
				<a href="%s" target="_blank" style="display: inline-block; background: #2563eb; color: #ffffff; text-decoration: none; font-size: 13px; font-weight: 600; padding: 8px 14px; border-radius: 6px;">
					Acessar Vaga &rarr;
				</a>`, escapeHTML(v.Link)))
		}

		// Badge de multi-fontes se encontrada em mais de um portal
		badgeMulti := ""
		if len(v.Fontes) > 1 {
			badgeMulti = fmt.Sprintf(`
				<span style="display: inline-block; background-color: #fef08a; color: #854d0e; font-size: 11px; font-weight: 700; padding: 2px 8px; border-radius: 4px; margin-bottom: 8px;">
					⭐ Disponível em %d plataformas (%s)
				</span>`, len(v.Fontes), escapeHTML(strings.Join(v.Fontes, ", ")))
		}

		fontePrincipal := v.Fonte
		if len(v.Fontes) > 0 {
			fontePrincipal = strings.Join(v.Fontes, ", ")
		}

		cardsHTML.WriteString(fmt.Sprintf(`
			<div style="background-color: #ffffff; border: 1px solid #e2e8f0; border-radius: 10px; padding: 20px; margin-bottom: 16px; box-shadow: 0 2px 4px rgba(0,0,0,0.03);">
				%s
				<h3 style="margin: 0; font-size: 17px; font-weight: 700; color: #0f172a; line-height: 1.3;">
					%s
				</h3>
				
				<div style="margin: 8px 0 12px 0; font-size: 13px; color: #475569; display: flex; flex-wrap: wrap; gap: 12px;">
					<span style="display: inline-block; margin-right: 12px;">🏢 <strong>%s</strong></span>
					<span style="display: inline-block; margin-right: 12px;">📍 %s</span>
					<span style="display: inline-block; background: #f1f5f9; color: #475569; padding: 1px 6px; border-radius: 4px; font-size: 11px;">Fontes: %s</span>
				</div>

				<div style="margin-bottom: 12px;">
					%s
				</div>

				<div style="margin-top: 14px;">
					%s
				</div>
			</div>
		`,
			badgeMulti,
			escapeHTML(v.Titulo),
			escapeHTML(v.Empresa),
			escapeHTML(v.Localizacao),
			escapeHTML(fontePrincipal),
			tagsHTML.String(),
			botoesHTML.String(),
		))
	}

	agoraFmt := time.Now().Format("02/01/2006 às 15:04")

	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
	<meta charset="utf-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>Novas Oportunidades Encontradas</title>
</head>
<body style="margin: 0; padding: 20px 10px; background-color: #f8fafc; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif; color: #334155;">
	<table role="presentation" width="100%%" cellspacing="0" cellpadding="0" border="0">
		<tr>
			<td align="center">
				<table role="presentation" width="100%%" style="max-width: 640px;" cellspacing="0" cellpadding="0" border="0">
					<!-- Cabeçalho -->
					<tr>
						<td style="padding: 24px 20px; background: linear-gradient(135deg, #1e3a8a 0%%, #2563eb 100%%); border-radius: 12px 12px 0 0; text-align: left; color: #ffffff;">
							<div style="display: inline-block; background: rgba(255,255,255,0.2); padding: 4px 10px; border-radius: 9999px; font-size: 12px; font-weight: 600; margin-bottom: 8px;">
								⚡ Vagas Scraper Multicanal
							</div>
							<h1 style="margin: 0 0 6px 0; font-size: 22px; font-weight: 800; color: #ffffff;">
								🎯 %d Novas Oportunidades Encontradas!
							</h1>
							<p style="margin: 0; font-size: 13px; color: #bfdbfe;">
								Varredura concluída em %s com deduplicação cruzada
							</p>
						</td>
					</tr>

					<!-- Conteúdo Principal -->
					<tr>
						<td style="background-color: #f1f5f9; padding: 20px;">
							%s
						</td>
					</tr>

					<!-- Rodapé -->
					<tr>
						<td style="background-color: #ffffff; padding: 18px 20px; border-radius: 0 0 12px 12px; text-align: center; border-top: 1px solid #e2e8f0;">
							<p style="margin: 0 0 4px 0; font-size: 12px; color: #64748b;">
								Robô de Vagas Inteligente • <strong>Go</strong> • Sem interpretadores
							</p>
							<p style="margin: 0; font-size: 11px; color: #94a3b8;">
								Configurado para rodar automaticamente via GitHub Actions ou Agendador do Windows
							</p>
						</td>
					</tr>
				</table>
			</td>
		</tr>
	</table>
</body>
</html>
`, len(vagas), agoraFmt, cardsHTML.String())
}

func escapeHTML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	return s
}
