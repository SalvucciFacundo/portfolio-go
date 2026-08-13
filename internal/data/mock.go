package data

import "github.com/SalvucciFacundo/portfolio-go/internal/domain"

const (
	devicon = "https://cdn.jsdelivr.net/gh/devicons/devicon/icons"
	cover   = "https://placehold.co/800x500/E0E0E0/1A1A1A?text="
)

// MockData returns the portfolio content. Bilingual editable fields carry both
// languages; the view picks one based on the active lang.
func MockData() domain.Profile {
	return domain.Profile{
		Name:       "Facundo Salvucci",
		RoleEs:     "Desarrollador Full Stack",
		RoleEn:     "Full Stack Developer",
		HeadlineEs: "Construyo aplicaciones web con Go, templ y HTMX.",
		HeadlineEn: "I build web applications with Go, templ and HTMX.",
		SummaryEs:  "Desarrollador Full Stack de Mendoza, Argentina. Me enfoco en backends sólidos con Go y en interfaces rápidas renderizadas en el servidor con templ + HTMX. Sistemas simples, rendimiento real y código que se entiende.",
		SummaryEn:  "Full Stack Developer from Mendoza, Argentina. I focus on solid backends with Go and fast server-rendered interfaces with templ + HTMX. Simple systems, real performance and code that reads well.",
		Email:      "fds1288@gmail.com",
		AvatarURL:  "https://placehold.co/400x400/666666/F4F4F2?text=FS",
		ResumeURL:  "",
		Socials: []domain.SocialLink{
			{Name: "GitHub", URL: "https://github.com/SalvucciFacundo", IconKey: "github"},
			{Name: "LinkedIn", URL: "https://linkedin.com/in/facundo-salvucci", IconKey: "linkedin"},
		},
		Skills: []domain.Skill{
			{Name: "Go", LevelEs: "Avanzado", LevelEn: "Advanced", IconURL: devicon + "/go/go-original.svg", IsTool: false},
			{Name: "templ", LevelEs: "Intermedio", LevelEn: "Intermediate", IconURL: "https://placehold.co/64x64/1A1A1A/F4F4F2?text=t", IsTool: false},
			{Name: "HTMX", LevelEs: "Intermedio", LevelEn: "Intermediate", IconURL: devicon + "/htmx/htmx-original.svg", IsTool: false},
			{Name: "PostgreSQL", LevelEs: "Intermedio", LevelEn: "Intermediate", IconURL: devicon + "/postgresql/postgresql-original.svg", IsTool: false},
			{Name: "Docker", LevelEs: "Intermedio", LevelEn: "Intermediate", IconURL: devicon + "/docker/docker-original.svg", IsTool: true},
			{Name: "TypeScript", LevelEs: "Avanzado", LevelEn: "Advanced", IconURL: devicon + "/typescript/typescript-original.svg", IsTool: false},
			{Name: "Angular", LevelEs: "Avanzado", LevelEn: "Advanced", IconURL: devicon + "/angular/angular-original.svg", IsTool: false},
			{Name: "Git", LevelEs: "Avanzado", LevelEn: "Advanced", IconURL: devicon + "/git/git-original.svg", IsTool: true},
		},
		Projects: []domain.Project{
			{
				TitleEs:       "Portafolio Go",
				TitleEn:       "Go Portfolio",
				DescriptionEs: "Este mismo portafolio: server-rendered con templ + HTMX, arquitectura hexagonal en Go y CSS puro con tokens de diseño.",
				DescriptionEn: "This very portfolio: server-rendered with templ + HTMX, hexagonal architecture in Go and plain CSS with design tokens.",
				Category:      "Web",
				StatusLabelEs: "Producción",
				StatusLabelEn: "Production",
				Tags:          []string{"Go", "templ", "HTMX", "PostgreSQL"},
				Link:          "",
				RepoLink:      "https://github.com/SalvucciFacundo/portfolio-go",
				CoverURL:      cover + "Portafolio+Go",
			},
			{
				TitleEs:       "GAIA",
				TitleEn:       "GAIA",
				DescriptionEs: "Agente de IA que conecta modelos de lenguaje con herramientas reales usando Go: orquestación de agentes, ejecución de tareas y respuestas en tiempo real.",
				DescriptionEn: "AI agent that connects language models with real tools using Go: agent orchestration, task execution and real-time responses.",
				Category:      "AI",
				StatusLabelEs: "Producción",
				StatusLabelEn: "Production",
				Tags:          []string{"Go", "AI", "LLM"},
				Link:          "",
				RepoLink:      "https://github.com/SalvucciFacundo",
				CoverURL:      cover + "GAIA",
			},
			{
				TitleEs:       "Mis Canarios",
				TitleEn:       "Mis Canarios",
				DescriptionEs: "App para el registro y seguimiento de canarios: datos de cada ave, concursos y descendencias. Frontend con Angular y datos en Firebase.",
				DescriptionEn: "App for registering and tracking canaries: bird data, contests and offspring. Angular frontend with Firebase as the data backend.",
				Category:      "Web",
				StatusLabelEs: "Producción",
				StatusLabelEn: "Production",
				Tags:          []string{"Angular", "Firebase", "TypeScript"},
				Link:          "",
				RepoLink:      "https://github.com/SalvucciFacundo",
				CoverURL:      cover + "Mis+Canarios",
			},
		},
		Experience: []domain.Experience{
			{
				PeriodEs:      "2023 — Presente",
				PeriodEn:      "2023 — Present",
				PositionEs:    "Desarrollador Full Stack",
				PositionEn:    "Full Stack Developer",
				Company:       "Freelance",
				DescriptionEs: "Desarrollo de aplicaciones web de extremo a extremo: API en Go, frontend server-rendered y despliegue.",
				DescriptionEn: "End-to-end web application development: Go APIs, server-rendered frontends and deployment.",
			},
			{
				PeriodEs:      "2023 — 2025",
				PeriodEn:      "2023 — 2025",
				PositionEs:    "QA Tester",
				PositionEn:    "QA Tester",
				Company:       "Dubbz",
				DescriptionEs: "Testing manual y automatizado de una plataforma de streaming: reporte de bugs, cobertura de regresiones y verificación de releases.",
				DescriptionEn: "Manual and automated testing of a streaming platform: bug reporting, regression coverage and release verification.",
			},
			{
				PeriodEs:      "2023 — 2025",
				PeriodEn:      "2023 — 2025",
				PositionEs:    "Customer Support Specialist",
				PositionEn:    "Customer Support Specialist",
				Company:       "Dubbz",
				DescriptionEs: "Soporte técnico de la plataforma: resolución de incidencias de usuarios y coordinación con los equipos de producto.",
				DescriptionEn: "Platform technical support: resolving user issues and coordinating with product teams.",
			},
		},
		Education: []domain.Education{
			{
				TitleEs:       "Tecnicatura Universitaria en Programación",
				TitleEn:       "Associate Degree in Computer Programming",
				School:        "Universidad Tecnológica Nacional (UTN)",
				Date:          "2019 — 2021",
				DescriptionEs: "Formación en programación, bases de datos y desarrollo de software.",
				DescriptionEn: "Training in programming, databases and software development.",
				IsCourse:      false,
			},
			{
				TitleEs:       "Técnico Mecánico Industrial",
				TitleEn:       "Industrial Mechanical Technician",
				School:        "Escuela Técnico Emilio Civit",
				Date:          "2012 — 2018",
				DescriptionEs: "Título técnico con orientación en mecánica industrial y sistemas.",
				DescriptionEn: "Technical degree focused on industrial mechanics and systems.",
				IsCourse:      false,
			},
			{
				TitleEs:       "Desarrollo Backend en Go",
				TitleEn:       "Go Backend Development",
				School:        "EducacionIT",
				Date:          "2023",
				DescriptionEs: "Curso avanzado de desarrollo web, concurrencia y bases de datos relacionales con Go.",
				DescriptionEn: "Advanced course on web development, concurrency, and relational databases with Go.",
				IsCourse:      true,
			},
			{
				TitleEs:       "Arquitectura Hexagonal y DDD",
				TitleEn:       "Hexagonal Architecture & DDD",
				School:        "Udemy",
				Date:          "2024",
				DescriptionEs: "Curso sobre desacoplamiento de código, patrones de diseño de software y diseño guiado por el dominio.",
				DescriptionEn: "Course on code decoupling, software design patterns, and Domain-Driven Design.",
				IsCourse:      true,
			},
		},
	}
}
