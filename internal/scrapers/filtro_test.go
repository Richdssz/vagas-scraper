package scrapers_test

import (
	"testing"
	"vagas-scraper/internal/models"
	"vagas-scraper/internal/scrapers"
)

func TestNormalizarTextoEContemTermo(t *testing.T) {
	testes := []struct {
		texto    string
		termo    string
		esperado bool
	}{
		{"Estágio em Operações Estruturadas", "operacoes estruturadas", true},
		{"ESTÁGIO DE INCLUSÃO (ANOS INICIAIS) - BOA VISTA", "inclusao", true},
		{"ESTÁGIO DE INCLUSÃO (ANOS INICIAIS) - BOA VISTA", "anos iniciais", true},
		{"Desenvolvedor Backend (Sr.)", "sr", true},
		{"Disrupção Digital", "sr", false}, // "sr" não deve casar dentro de "disrupcao"
		{"Estágio em TI", "ti", true},
		{"Assistente Administrativo", "ti", false}, // "ti" não deve casar em "assistente"
		{"Analista de Redes", "redes", true},
		{"Paredes Pintadas", "redes", false}, // "redes" não deve casar em "paredes"
		{"Ciência de Dados", "dados", true},
		{"Cuidados Médicos", "dados", false}, // "dados" não deve casar em "cuidados"
		{"Desenvolvedor Pleno / Sênior", "senior", true},
		{"Desenvolvedor Pleno / Sênior", "pleno", true},
	}

	for _, tc := range testes {
		txtNorm := scrapers.NormalizarTexto(tc.texto)
		trmNorm := scrapers.NormalizarTexto(tc.termo)
		resultado := scrapers.ContemTermo(txtNorm, trmNorm)
		if resultado != tc.esperado {
			t.Errorf("ContemTermo('%s', '%s') = %v, esperado %v (normalizado: '%s' x '%s')",
				tc.texto, tc.termo, resultado, tc.esperado, txtNorm, trmNorm)
		}
	}
}

func TestFiltroAceitaVagaCenariosReais(t *testing.T) {
	termosBusca := []string{
		"estagio", "estágio", "junior", "júnior", "trainee",
		"backend", "java", "node", "fullstack", "react", "spring",
		"flutter", "dart", "sql", "dados", "redes",
	}

	termosExclusao := []string{
		"senior", "sênior", "sr.", "sr", "pleno", "tech lead", "especialista", "coordenador",
		"operacoes estruturadas", "inclusao", "anos iniciais", "pedagogia", "educacao",
		"contabilidade", "contabil", "comercial", "geologia", "enfermagem", "juridico",
		"administrativo",
	}

	// Cenários que DEVEM ser rejeitados (fora do nicho ou excluídos)
	rejeitados := []models.Vaga{
		{
			Titulo: "Estágio em Operações Estruturadas",
			Empresa: "BANCO DAYCOVAL",
			Fonte: "Gupy",
		},
		{
			Titulo: "ESTÁGIO DE INCLUSÃO (ANOS INICIAIS) - BOA VISTA",
			Empresa: "Grupo Positivo",
			Fonte: "Gupy",
		},
		{
			Titulo: "Estágio Contabilidade - Belo Horizonte/MG",
			Empresa: "Skill Consultoria",
			Fonte: "Gupy",
		},
		{
			Titulo: "Estágio Comercial (Área de Operações)",
			Empresa: "Monte Carlo Joias",
			Fonte: "Gupy",
		},
		{
			Titulo: "Nube | Oportunidade de Estágio em Geologia - 381415",
			Empresa: "Nube",
			Fonte: "Gupy",
		},
		{
			Titulo: "Estágio Jurídico Tributário",
			Empresa: "Gertec Brasil",
			Fonte: "Gupy",
		},
		{
			Titulo: "Estágio - Administrativo",
			Empresa: "Serilon",
			Fonte: "Gupy",
		},
		{
			Titulo: "Desenvolvedor Backend Sênior",
			Empresa: "Tech Co",
			Fonte: "Gupy",
		},
		{
			Titulo: "Pessoa Desenvolvedora Pleno",
			Empresa: "Tech Co",
			Fonte: "Gupy",
		},
		{
			Titulo: "Tech Lead Java",
			Empresa: "Tech Co",
			Fonte: "Gupy",
		},
		{
			Titulo: "Desenvolvedor(a) Full Stack SêniorNOVA",
			Empresa: "Galileu Saúde",
			Fonte: "ProgramaThor",
		},
		{
			Titulo: "Desenvolvedor(a) Back-End PlenoNOVA",
			Empresa: "Concept IaaS",
			Fonte: "ProgramaThor",
		},
	}

	for _, v := range rejeitados {
		if scrapers.FiltroAceitaVaga(v, termosBusca, termosExclusao) {
			t.Errorf("FALHA: Vaga deveria ser REJEITADA mas foi aceita: '%s' (%s)", v.Titulo, v.Fonte)
		}
	}

	// Cenários que DEVEM ser aceitos (no nicho e não proibidos)
	aceitos := []models.Vaga{
		{
			Titulo: "ESTÁGIO EM TI",
			Empresa: "Grupo CEV",
			Fonte: "Gupy",
		},
		{
			Titulo: "Visagio Talentos - Estágio: Desenvolvedor(a) de Software",
			Empresa: "Visagio",
			Fonte: "Gupy",
		},
		{
			Titulo: "Estágio TI - (Sistemas)",
			Empresa: "Agronegócio",
			Fonte: "Gupy",
		},
		{
			Titulo: "Estágio em Desenvolvimento Backend",
			Empresa: "Startup Tech",
			Fonte: "Gupy",
		},
		{
			Titulo: "Desenvolvedor(a) Full Stack Júnior — React e TypeScript",
			Empresa: "Viradev",
			Fonte: "ProgramaThor",
		},
		{
			Titulo: "[Remoto] Full-stack Engineer React.js • UX Design • Node.js",
			Empresa: "Strider",
			Fonte: "Frontend-BR",
		},
		{
			Titulo: "[Remoto] Estagiário - Engenharia de Software",
			Empresa: "Empresa X",
			Fonte: "Backend-BR",
		},
	}

	for _, v := range aceitos {
		if !scrapers.FiltroAceitaVaga(v, termosBusca, termosExclusao) {
			t.Errorf("FALHA: Vaga deveria ser ACEITA mas foi rejeitada: '%s' (%s)", v.Titulo, v.Fonte)
		}
	}
}

func TestResolverTermosPortais(t *testing.T) {
	// Se tiver estagio nos termos, deve gerar consultas qualificadas para tech no Gupy/LinkedIn
	termos := []string{"estagio", "junior", "backend", "java"}
	resolvidos := scrapers.ResolverTermosPortais(termos)

	if len(resolvidos) == 0 {
		t.Fatalf("Esperava termos resolvidos, recebeu vazio")
	}

	encontrouEstagioTech := false
	for _, r := range resolvidos {
		if r == "estagio ti" || r == "estagio desenvolvedor" {
			encontrouEstagioTech = true
		}
	}

	if !encontrouEstagioTech {
		t.Errorf("Esperava que o resolver qualificasse estágio para TI/desenvolvedor, recebeu: %v", resolvidos)
	}
}
