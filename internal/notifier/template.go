package notifier

import (
	"fmt"
	"strings"
	"time"
	"vagas-scraper/internal/models"
)

// GerarHTMLEmail produz um email com estética editorial minimalista, tipografia apurada
// e sem ruído visual ou menção a linguagens/tecnologias de backend.
func GerarHTMLEmail(vagas []models.Vaga) string {
	var cardsHTML strings.Builder

	for _, v := range vagas {
		// Tags discretas e estruturadas
		var tagsHTML strings.Builder
		for _, tag := range v.Tags {
			t := strings.TrimSpace(tag)
			if t == "" {
				continue
			}
			tagsHTML.WriteString(fmt.Sprintf(`
				<span style="display: inline-block; font-size: 11px; font-weight: 500; color: #52525b; background-color: #f4f4f5; border: 1px solid #e4e4e7; padding: 2px 7px; border-radius: 4px; margin-right: 4px; margin-bottom: 4px;">
					%s
				</span>`, escapeHTML(t)))
		}

		// Botões de link limpos
		var botoesHTML strings.Builder
		if len(v.Links) > 0 {
			for _, lf := range v.Links {
				botoesHTML.WriteString(fmt.Sprintf(`
					<a href="%s" target="_blank" style="display: inline-block; background-color: #18181b; color: #fafafa; font-size: 12px; font-weight: 600; text-decoration: none; padding: 7px 14px; border-radius: 5px; margin-right: 8px; margin-bottom: 6px;">
						Ver no %s &rarr;
					</a>`, escapeHTML(lf.URL), escapeHTML(lf.Fonte)))
			}
		} else if v.Link != "" {
			botoesHTML.WriteString(fmt.Sprintf(`
				<a href="%s" target="_blank" style="display: inline-block; background-color: #18181b; color: #fafafa; font-size: 12px; font-weight: 600; text-decoration: none; padding: 7px 14px; border-radius: 5px;">
					Acessar vaga &rarr;
				</a>`, escapeHTML(v.Link)))
		}

		// Indicador de múltiplas fontes
		indicadorMulti := ""
		if len(v.Fontes) > 1 {
			indicadorMulti = fmt.Sprintf(`
				<div style="font-size: 11px; font-weight: 600; color: #0284c7; margin-bottom: 6px; text-transform: uppercase; letter-spacing: 0.5px;">
					Disponível em %d plataformas (%s)
				</div>`, len(v.Fontes), escapeHTML(strings.Join(v.Fontes, ", ")))
		}

		fonteTexto := v.Fonte
		if len(v.Fontes) > 0 {
			fonteTexto = strings.Join(v.Fontes, " · ")
		}

		cardsHTML.WriteString(fmt.Sprintf(`
			<div style="background-color: #ffffff; border: 1px solid #e4e4e7; border-radius: 8px; padding: 20px; margin-bottom: 14px;">
				%s
				<h3 style="margin: 0 0 6px 0; font-size: 16px; font-weight: 600; color: #09090b; line-height: 1.35;">
					%s
				</h3>
				
				<div style="font-size: 13px; color: #71717a; margin-bottom: 12px;">
					<strong style="color: #27272a;">%s</strong>
					<span style="color: #d4d4d8; margin: 0 6px;">·</span>
					<span>%s</span>
					<span style="color: #d4d4d8; margin: 0 6px;">·</span>
					<span>%s</span>
				</div>

				<div style="margin-bottom: 14px;">
					%s
				</div>

				<div>
					%s
				</div>
			</div>
		`,
			indicadorMulti,
			escapeHTML(v.Titulo),
			escapeHTML(v.Empresa),
			escapeHTML(v.Localizacao),
			escapeHTML(fonteTexto),
			tagsHTML.String(),
			botoesHTML.String(),
		))
	}

	dataExtenso := time.Now().Format("02 de January de 2006 às 15:04")
	dataExtenso = traduzirMes(dataExtenso)

	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
	<meta charset="utf-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>Digest de Vagas</title>
</head>
<body style="margin: 0; padding: 32px 12px; background-color: #fafafa; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif; color: #18181b;">
	<table role="presentation" width="100%%" cellspacing="0" cellpadding="0" border="0">
		<tr>
			<td align="center">
				<table role="presentation" width="100%%" style="max-width: 600px;" cellspacing="0" cellpadding="0" border="0">
					<!-- Cabeçalho Editorial -->
					<tr>
						<td style="padding: 24px 0 20px 0; border-bottom: 2px solid #18181b; text-align: left;">
							<div style="font-size: 11px; font-weight: 700; letter-spacing: 1px; text-transform: uppercase; color: #71717a; margin-bottom: 6px;">
								Relatório Periódico de Oportunidades
							</div>
							<h1 style="margin: 0 0 8px 0; font-size: 24px; font-weight: 700; color: #09090b; letter-spacing: -0.5px;">
								%d Novas Vagas Selecionadas
							</h1>
							<p style="margin: 0; font-size: 13px; color: #71717a;">
								Varredura concluída em %s com deduplicação cruzada
							</p>
						</td>
					</tr>

					<!-- Lista de Vagas -->
					<tr>
						<td style="padding-top: 20px;">
							%s
						</td>
					</tr>

					<!-- Rodapé Neutro -->
					<tr>
						<td style="padding: 24px 0; border-top: 1px solid #e4e4e7; text-align: center;">
							<p style="margin: 0 0 4px 0; font-size: 12px; color: #71717a;">
								Radar de Oportunidades · Notificação automatizada
							</p>
							<p style="margin: 0; font-size: 11px; color: #a1a1aa;">
								Filtros calibrados e deduplicação unificada entre plataformas
							</p>
						</td>
					</tr>
				</table>
			</td>
		</tr>
	</table>
</body>
</html>`, len(vagas), dataExtenso, cardsHTML.String())
}

func traduzirMes(s string) string {
	m := map[string]string{
		"January": "Janeiro", "February": "Fevereiro", "March": "Março",
		"April": "Abril", "May": "Maio", "June": "Junho",
		"July": "Julho", "August": "Agosto", "September": "Setembro",
		"October": "Outubro", "November": "Novembro", "December": "Dezembro",
	}
	for en, pt := range m {
		s = strings.ReplaceAll(s, en, pt)
	}
	return s
}

func escapeHTML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	return s
}
