package notifier

import (
	"fmt"
	"strings"
	"time"
	"vagas-scraper/internal/models"
)

// GerarHTMLEmail produz um email com estética dark moderna (fundo preto e superfícies escuras),
// idêntica ao design do painel localhost, com tipografia apurada e legibilidade máxima.
func GerarHTMLEmail(vagas []models.Vaga) string {
	var cardsHTML strings.Builder

	for _, v := range vagas {
		// Tags estruturadas no estilo do localhost (.chip-include)
		var tagsHTML strings.Builder
		for _, tag := range v.Tags {
			t := strings.TrimSpace(tag)
			if t == "" {
				continue
			}
			tagsHTML.WriteString(fmt.Sprintf(`
				<span style="display: inline-block; font-size: 11px; font-family: 'JetBrains Mono', Consolas, Monaco, monospace; font-weight: 500; color: #93c5fd; background-color: rgba(37, 99, 235, 0.12); border: 1px solid rgba(37, 99, 235, 0.3); padding: 2px 7px; border-radius: 4px; margin-right: 5px; margin-bottom: 5px;">
					%s
				</span>`, escapeHTML(t)))
		}

		// Botões de link no padrão do localhost (.job-link-btn)
		var botoesHTML strings.Builder
		if len(v.Links) > 0 {
			for _, lf := range v.Links {
				botoesHTML.WriteString(fmt.Sprintf(`
					<a href="%s" target="_blank" style="display: inline-block; background-color: #18181b; border: 1px solid #3f3f46; color: #fafafa; font-size: 12px; font-weight: 600; text-decoration: none; padding: 7px 14px; border-radius: 6px; margin-right: 8px; margin-bottom: 6px;">
						Ver no %s &rarr;
					</a>`, escapeHTML(lf.URL), escapeHTML(lf.Fonte)))
			}
		} else if v.Link != "" {
			nomeFonte := v.Fonte
			if nomeFonte == "" {
				nomeFonte = "Portal"
			}
			botoesHTML.WriteString(fmt.Sprintf(`
				<a href="%s" target="_blank" style="display: inline-block; background-color: #18181b; border: 1px solid #3f3f46; color: #fafafa; font-size: 12px; font-weight: 600; text-decoration: none; padding: 7px 14px; border-radius: 6px; margin-right: 8px; margin-bottom: 6px;">
					Ver no %s &rarr;
				</a>`, escapeHTML(v.Link), escapeHTML(nomeFonte)))
		}
		botoesHTML.WriteString(botaoMarcarVista(v))

		fonteTexto := v.Fonte
		if len(v.Fontes) > 0 {
			fonteTexto = strings.Join(v.Fontes, " · ")
		}
		if fonteTexto == "" {
			fonteTexto = "Portal de Vagas"
		}

		// Indicador de múltiplas fontes (.multi-tag no estilo do localhost)
		indicadorMulti := ""
		if len(v.Fontes) > 1 {
			indicadorMulti = fmt.Sprintf(`
				<span style="display: inline-block; font-size: 11px; font-family: 'JetBrains Mono', Consolas, Monaco, monospace; font-weight: 600; color: #fbbf24; background-color: rgba(245, 158, 11, 0.12); border: 1px solid rgba(245, 158, 11, 0.3); padding: 2px 8px; border-radius: 4px; text-transform: uppercase; letter-spacing: 0.5px; margin-left: 6px;">
					Presente em %d portais (%s)
				</span>`, len(v.Fontes), escapeHTML(strings.Join(v.Fontes, ", ")))
		}

		// Badge de data / horário
		tempo := v.TempoRelativo()
		tempoHTML := ""
		if tempo != "" {
			tempoHTML = fmt.Sprintf(`<span style="color: #27272a; margin: 0 6px;">·</span><span style="color: #34d399; font-weight: 500;">%s</span>`, escapeHTML(tempo))
		}

		localTexto := v.Localizacao
		if localTexto == "" {
			localTexto = "Remoto / Brasil"
		}

		cardsHTML.WriteString(fmt.Sprintf(`
			<div style="background-color: #121215; border: 1px solid #27272a; border-radius: 8px; padding: 20px; margin-bottom: 14px; text-align: left;">
				<div style="margin-bottom: 8px;">
					<span style="display: inline-block; font-size: 11px; font-family: 'JetBrains Mono', Consolas, Monaco, monospace; font-weight: 600; color: #60a5fa; background-color: rgba(37, 99, 235, 0.12); border: 1px solid rgba(37, 99, 235, 0.3); padding: 2px 8px; border-radius: 4px; letter-spacing: 0.3px;">
						Portal: %s
					</span>
					%s
				</div>

				<h3 style="margin: 0 0 8px 0; font-size: 16px; font-weight: 600; color: #93c5fd; line-height: 1.35;">
					%s
				</h3>
				
				<div style="font-size: 13px; color: #71717a; margin-bottom: 12px; line-height: 1.6;">
					<strong style="color: #fafafa; font-weight: 600;">%s</strong>
					<span style="color: #27272a; margin: 0 6px;">·</span>
					<span style="color: #a1a1aa;">%s</span>
					%s
				</div>

				%s

				<div style="margin-top: 12px;">
					%s
				</div>
			</div>
		`,
			escapeHTML(fonteTexto),
			indicadorMulti,
			escapeHTML(v.Titulo),
			escapeHTML(v.Empresa),
			escapeHTML(localTexto),
			tempoHTML,
			func() string {
				if tagsHTML.Len() > 0 {
					return fmt.Sprintf(`<div style="margin-bottom: 12px;">%s</div>`, tagsHTML.String())
				}
				return ""
			}(),
			botoesHTML.String(),
		))
	}

	// Resumo dos portais consultados
	var portaisUnicos []string
	portaisVistos := make(map[string]bool)
	for _, v := range vagas {
		if v.Fonte != "" && !portaisVistos[v.Fonte] {
			portaisVistos[v.Fonte] = true
			portaisUnicos = append(portaisUnicos, v.Fonte)
		}
		for _, pf := range v.Fontes {
			if pf != "" && !portaisVistos[pf] {
				portaisVistos[pf] = true
				portaisUnicos = append(portaisUnicos, pf)
			}
		}
	}
	resumoPortais := strings.Join(portaisUnicos, " · ")
	if resumoPortais == "" {
		resumoPortais = "LinkedIn · Gupy · GitHub · ProgramaThor · RemoteOK"
	}

	dataExtenso := time.Now().Format("02 de January de 2006 às 15:04")
	dataExtenso = traduzirMes(dataExtenso)

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="pt-BR">
<head>
	<meta charset="utf-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<meta name="color-scheme" content="dark">
	<meta name="supported-color-schemes" content="dark">
	<title>Digest de Vagas • Radar</title>
	<style>
		:root { color-scheme: dark; supported-color-schemes: dark; }
		body { background-color: #09090b !important; color: #fafafa !important; }
		a { color: #93c5fd; }
	</style>
</head>
<body style="margin: 0; padding: 32px 12px; background-color: #09090b; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif; color: #fafafa; -webkit-font-smoothing: antialiased;">
	<table role="presentation" width="100%%" cellspacing="0" cellpadding="0" border="0" style="background-color: #09090b;">
		<tr>
			<td align="center" style="background-color: #09090b;">
				<table role="presentation" width="100%%" style="max-width: 620px; background-color: #09090b;" cellspacing="0" cellpadding="0" border="0">
					<!-- Cabeçalho idêntico ao localhost -->
					<tr>
						<td style="padding: 0 0 20px 0; text-align: left;">
							<div style="background-color: #121215; border: 1px solid #27272a; border-radius: 8px; padding: 22px 24px;">
								<div style="margin-bottom: 10px;">
									<span style="font-family: 'JetBrains Mono', Consolas, Monaco, monospace; font-size: 11px; font-weight: 600; color: #34d399; background-color: rgba(16, 185, 129, 0.1); border: 1px solid rgba(16, 185, 129, 0.25); padding: 3px 8px; border-radius: 4px; letter-spacing: 0.5px;">
										RADAR DE VAGAS • ATIVO
									</span>
								</div>
								<h1 style="margin: 0 0 8px 0; font-size: 22px; font-weight: 700; color: #fafafa; letter-spacing: -0.4px;">
									%d Novas Vagas Selecionadas
								</h1>
								<p style="margin: 0 0 8px 0; font-size: 13px; color: #a1a1aa; line-height: 1.5;">
									Varredura concluída em %s com deduplicação unificada entre plataformas.
								</p>
								<div style="font-size: 12px; color: #71717a;">
									Portais: <strong style="color: #93c5fd; font-weight: 500;">%s</strong>
								</div>
							</div>
						</td>
					</tr>

					<!-- Lista de Vagas -->
					<tr>
						<td>
							%s
						</td>
					</tr>

					<!-- Rodapé Neutro Dark -->
					<tr>
						<td style="padding: 24px 0 16px 0; border-top: 1px solid #27272a; text-align: center;">
							<p style="margin: 0 0 6px 0; font-size: 12px; color: #71717a;">
								Radar de Oportunidades · Notificação automatizada
							</p>
							<p style="margin: 0; font-size: 11px; color: #52525b;">
								Filtros calibrados e deduplicação ativa · <a href="http://localhost:8080" style="color: #60a5fa; text-decoration: none;">Abrir Painel de Controle Local</a>
							</p>
						</td>
					</tr>
				</table>
			</td>
		</tr>
	</table>
</body>
</html>`, len(vagas), dataExtenso, resumoPortais, cardsHTML.String())
}

// GerarHTMLEmailSemVagas gera o e-mail informativo avisando que o robô executou com sucesso
// mas todas as vagas encontradas já haviam sido visualizadas anteriormente.
func GerarHTMLEmailSemVagas(totalEncontradas int, portaisConsultados []string) string {
	now := time.Now()
	dataExtenso := fmt.Sprintf("%02d de %s de %d às %02d:%02d",
		now.Day(),
		traduzirMes(now.Month().String()),
		now.Year(),
		now.Hour(),
		now.Minute(),
	)

	resumoPortais := "LinkedIn, Gupy, GitHub e outros"
	if len(portaisConsultados) > 0 {
		resumoPortais = strings.Join(portaisConsultados, ", ")
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="pt-BR">
<head>
	<meta charset="utf-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<meta name="color-scheme" content="dark">
	<meta name="supported-color-schemes" content="dark">
	<title>Radar de Vagas • Sem Novidades</title>
	<style>
		:root { color-scheme: dark; supported-color-schemes: dark; }
		body { background-color: #09090b !important; color: #fafafa !important; }
		a { color: #93c5fd; }
	</style>
</head>
<body style="margin: 0; padding: 32px 12px; background-color: #09090b; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif; color: #fafafa; -webkit-font-smoothing: antialiased;">
	<table role="presentation" width="100%%" cellspacing="0" cellpadding="0" border="0" style="background-color: #09090b;">
		<tr>
			<td align="center" style="background-color: #09090b;">
				<table role="presentation" width="100%%" style="max-width: 620px; background-color: #09090b;" cellspacing="0" cellpadding="0" border="0">
					<!-- Cabeçalho Dark -->
					<tr>
						<td style="padding: 0 0 20px 0; text-align: left;">
							<div style="background-color: #121215; border: 1px solid #27272a; border-radius: 8px; padding: 22px 24px;">
								<div style="margin-bottom: 10px;">
									<span style="font-family: 'JetBrains Mono', Consolas, Monaco, monospace; font-size: 11px; font-weight: 600; color: #a1a1aa; background-color: rgba(161, 161, 170, 0.1); border: 1px solid rgba(161, 161, 170, 0.25); padding: 3px 8px; border-radius: 4px; letter-spacing: 0.5px;">
										RADAR DE VAGAS • ATUALIZADO
									</span>
								</div>
								<h1 style="margin: 0 0 8px 0; font-size: 22px; font-weight: 700; color: #fafafa; letter-spacing: -0.4px;">
									Nenhuma Nova Vaga Nesta Rodada
								</h1>
								<p style="margin: 0 0 8px 0; font-size: 13px; color: #a1a1aa; line-height: 1.5;">
									O robô consultou os portais com sucesso em %s, mas todas as oportunidades encontradas já constam no seu histórico de enviadas.
								</p>
								<div style="font-size: 12px; color: #71717a;">
									Portais verificados: <strong style="color: #93c5fd; font-weight: 500;">%s</strong>
								</div>
							</div>
						</td>
					</tr>

					<!-- Card Explicativo -->
					<tr>
						<td style="padding: 0 0 20px 0;">
							<div style="background-color: #121215; border: 1px solid #27272a; border-radius: 8px; padding: 20px 24px;">
								<div style="font-size: 14px; font-weight: 600; color: #e4e4e7; margin-bottom: 8px;">
									🔍 Status da Varredura
								</div>
								<p style="margin: 0 0 10px 0; font-size: 13px; color: #a1a1aa; line-height: 1.5;">
									Foram analisadas <strong>%d vagas brutas</strong> nos portais. Como o filtro anti-duplicação está ativo, nenhuma vaga repetida foi encaminhada.
								</p>
								<p style="margin: 0; font-size: 12px; color: #71717a; line-height: 1.5;">
									Você receberá as novas oportunidades assim que forem publicadas pelas empresas na próxima checagem agendada!
								</p>
							</div>
						</td>
					</tr>

					<!-- Rodapé -->
					<tr>
						<td style="padding: 24px 0 16px 0; border-top: 1px solid #27272a; text-align: center;">
							<p style="margin: 0 0 6px 0; font-size: 12px; color: #71717a;">
								Radar de Oportunidades · Notificação automatizada
							</p>
							<p style="margin: 0; font-size: 11px; color: #52525b;">
								Filtros calibrados e deduplicação ativa
							</p>
						</td>
					</tr>
				</table>
			</td>
		</tr>
	</table>
</body>
</html>`, dataExtenso, resumoPortais, totalEncontradas)
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
