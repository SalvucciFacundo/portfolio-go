package data

import (
	"sync"

	"github.com/SalvucciFacundo/portfolio-go/internal/domain"
)

const (
	devicon = "https://cdn.jsdelivr.net/gh/devicons/devicon/icons"
	cover   = "https://placehold.co/800x500/E0E0E0/1A1A1A?text="
)

var (
	profile domain.Profile
	mu      sync.RWMutex
)

func init() {
	profile = initMockProfile()
}

// GetProfile returns a copy of the current profile.
func GetProfile() domain.Profile {
	mu.RLock()
	defer mu.RUnlock()
	return profile
}

// UpdateProfile replaces the in-memory profile.
func UpdateProfile(p domain.Profile) {
	mu.Lock()
	profile = p
	mu.Unlock()
}

// MockData returns the profile (kept for backward compatibility).
func MockData() domain.Profile {
	return GetProfile()
}

func initMockProfile() domain.Profile {
	return domain.Profile{
		Name:       "Facundo Salvucci",
		RoleEs:     "Full Stack Developer & DevTools Builder",
		RoleEn:     "Full Stack Developer & DevTools Builder",
		HeadlineEs: "Desarrollo aplicaciones web robustas con Go y TypeScript, creando además herramientas que optimizan el flujo de trabajo de otros desarrolladores.",
		HeadlineEn: "Building robust web applications with Go and TypeScript, while crafting tools that streamline developer workflows.",
		SummaryEs:  "Desarrollador Full Stack de Mendoza, Argentina. En el backend trabajo con Go aplicando arquitectura limpia, concurrencia y análisis estático con AST. En el frontend construyo interfaces ágiles con Angular y React. Mi diferencial está en crear tooling práctico —servidores MCP, linters y CLIs— pensado para resolver problemas reales del ecosistema.",
		SummaryEn:  "Full Stack Developer based in Mendoza, Argentina. On the backend I work with Go focusing on clean architecture, concurrency, and AST static analysis. On the frontend I build responsive interfaces with Angular and React. My edge is creating practical tooling —MCP servers, linters, and CLIs— designed to solve real-world problems in the ecosystem.",
		Email:      "fds1288@gmail.com",
		AvatarURL:  "https://placehold.co/400x400/666666/F4F4F2?text=FS",
		ResumeURL:  "",
		Socials: []domain.SocialLink{
			{Name: "GitHub", URL: "https://github.com/SalvucciFacundo", IconKey: "github"},
			{Name: "LinkedIn", URL: "https://linkedin.com/in/facundo-salvucci", IconKey: "linkedin"},
		},
		Skills: []domain.Skill{
			{Name: "Go", IconURL: devicon + "/go/go-original.svg", IsTool: false},
			{Name: "templ", IconURL: "https://placehold.co/64x64/1A1A1A/F4F4F2?text=t", IsTool: false},
			{Name: "HTMX", IconURL: devicon + "/htmx/htmx-original.svg", IsTool: false},
			{Name: "PostgreSQL", IconURL: devicon + "/postgresql/postgresql-original.svg", IsTool: false},
			{Name: "Docker", IconURL: devicon + "/docker/docker-original.svg", IsTool: true},
			{Name: "TypeScript", IconURL: devicon + "/typescript/typescript-original.svg", IsTool: false},
			{Name: "Angular", IconURL: devicon + "/angular/angular-original.svg", IsTool: false},
			{Name: "Git", IconURL: devicon + "/git/git-original.svg", IsTool: true},
		},
		Projects: []domain.Project{
			{
				TitleEs:           "Portafolio Go",
				TitleEn:           "Go Portfolio",
				DescriptionEs:     "Este mismo portafolio: server-rendered con templ + HTMX, arquitectura hexagonal en Go y CSS puro con tokens de diseño.",
				DescriptionEn:     "This very portfolio: server-rendered with templ + HTMX, hexagonal architecture in Go and plain CSS with design tokens.",
				TechDescriptionEs: "Implementado con Go vanilla, usando templ para tipado seguro en vistas y HTMX para swaps parciales. La arquitectura hexagonal separa dominio y adaptadores de transporte de forma limpia.",
				TechDescriptionEn: "Implemented in vanilla Go, using templ for type-safe rendering and HTMX for partial DOM swaps. Hexagonal architecture separates domain models and transport adapters cleanly.",
				Category:          "Web",
				Tags:              []string{"Go", "templ", "HTMX", "PostgreSQL"},
				Link:              "",
				RepoLink:          "https://github.com/SalvucciFacundo/portfolio-go",
				CoverURL:          cover + "Portafolio+Go",
				Screenshots:       []domain.ProjectImage{{URL: cover + "Portafolio+Go"}, {URL: cover + "Bento+Layout"}, {URL: cover + "Hexagonal+Architecture"}},
			},
			{
				TitleEs:           "GAIA",
				TitleEn:           "GAIA",
				DescriptionEs:     "Agente de IA que conecta modelos de lenguaje con herramientas reales usando Go: orquestación de agentes, ejecución de tareas y respuestas en tiempo real.",
				DescriptionEn:     "AI agent that connects language models with real tools using Go: agent orchestration, task execution and real-time responses.",
				TechDescriptionEs: "Orquestador de agentes implementado sobre modelos LLM de Anthropic y OpenAI. El motor de herramientas en Go resuelve en paralelo llamadas externas y valida esquemas JSON dinámicamente.",
				TechDescriptionEn: "Agent orchestrator built on top of Anthropic and OpenAI LLM models. The Go-based tool executor handles parallel tool invocations and validates JSON schemas dynamically.",
				Category:          "AI",
				Tags:              []string{"Go", "AI", "LLM"},
				Link:              "",
				RepoLink:          "https://github.com/SalvucciFacundo",
				CoverURL:          cover + "GAIA",
				Screenshots:       []domain.ProjectImage{{URL: cover + "GAIA"}, {URL: cover + "AI+Agent+Flow"}, {URL: cover + "JSON+Validation"}},
			},
			{
				TitleEs:           "Mis Canarios",
				TitleEn:           "Mis Canarios",
				DescriptionEs:     "App para el registro y seguimiento de canarios: datos de cada ave, concursos y descendencias. Frontend con Angular y datos en Firebase.",
				DescriptionEn:     "App for registering and tracking canaries: bird data, contests and offspring. Angular frontend with Firebase as the data backend.",
				TechDescriptionEs: "Cliente SPA desarrollado con Angular que consume servicios reactivos de Firestore. Cuenta con sincronización de estado local e indexación gráfica de pedigrí de aves.",
				TechDescriptionEn: "SPA client developed with Angular consuming reactive Firestore services. Features local offline synchronization state and graphical pedigree tree rendering.",
				Category:          "Web",
				Tags:              []string{"Angular", "Firebase", "TypeScript"},
				Link:              "",
				RepoLink:          "https://github.com/SalvucciFacundo",
				CoverURL:          cover + "Mis+Canarios",
				Screenshots:       []domain.ProjectImage{{URL: cover + "Mis+Canarios"}, {URL: cover + "Canary+Genetics"}, {URL: cover + "Firebase+Sync"}},
			},
		},
		Experience: []domain.Experience{
			{
				PeriodEs:      "2023 — Presente",
				PeriodEn:      "2023 — Present",
				PositionEs:    "Desarrollador Full Stack",
				PositionEn:    "Full-Stack Developer",
				Company:       "Freelance / Independent",
				DescriptionEs: "• Arquitectura, desarrollo y despliegue de aplicaciones web de producción de punta a punta, gestionando el ciclo completo de vida de software hasta el despliegue en servidores autohospedados con Linux, Dokploy y Docker.\n• Mis Canarios (SaaS en Producción): Plataforma integral de gestión para criadores de aves con Angular y Firebase. Monetización por suscripciones recurrentes con Mercado Pago Checkout Pro y webhooks automatizados.\n• ScraperHub (Web App en Producción): Plataforma de web scraping bajo demanda y marketplace de datos en Go (Chi), con extracción instantánea de datasets estructurados, vistas previas freemium y cobros con Lemon Squeezy.\n• FDSTech (Plataforma en Producción): Soluciones digitales con flujos automatizados en n8n para gestión de leads, notificaciones y orquestación de tareas en Linux.\n• Developer Tools & Agentes de IA: Creación de herramientas open-source y agentes autónomos en Go: arch-vet (linter AST y guardrail MCP para IA), GAIA (asistente de código multi-agente con SDD), templ-islands (librería de islas para Go SSR) y go-arch.",
				DescriptionEn: "• Architected, developed, and deployed end-to-end production web applications, managing the complete lifecycle from requirements gathering to server deployment on self-hosted Linux, Dokploy, and Docker.\n• Mis Canarios (Production SaaS): Engineered an all-in-one management platform for bird breeders using Angular and Firebase. Integrated recurring subscription monetization with Mercado Pago Checkout Pro and automated webhooks.\n• ScraperHub (Production Web App): Developed an on-demand web scraping platform and data marketplace in Go (Chi), featuring instant structured dataset extraction, freemium previews, and international payment processing via Lemon Squeezy.\n• FDSTech (Production Platform): Deployed a digital solutions platform integrated with n8n automated workflows for lead management, customer notifications, and task orchestration on self-hosted Linux infrastructure.\n• Developer Tools & AI Agents: Built open-source developer tooling and autonomous AI agents in Go: arch-vet (AST architecture linter & native MCP guardrail for AI), GAIA (autonomous multi-agent coding assistant with SDD), templ-islands (islands architecture library for Go SSR), and go-arch.",
			},
			{
				PeriodEs:      "2023 — 2025",
				PeriodEn:      "2023 — 2025",
				PositionEs:    "QA Tester & Especialista en Soporte Técnico",
				PositionEn:    "QA Tester & Technical Support Specialist",
				Company:       "Dubbz",
				DescriptionEs: "• Ejecución de pruebas manuales, exploratorias y de regresión en plataformas web para identificar, documentar y reportar defectos de software con pasos de reproducción precisos.\n• Colaboración estrecha con equipos de ingeniería multidisciplinarios para validar correcciones de bugs y verificar el cumplimiento de especificaciones técnicas.\n• Gestión de soporte técnico Tier-1 y Tier-2 en un entorno ágil de alto ritmo, traduciendo el feedback de los usuarios en reportes de fallos accionables y mejoras continuas de producto.",
				DescriptionEn: "• Conducted manual, exploratory, and regression testing across web platforms to identify, document, and track software defects with clear reproduction steps.\n• Collaborated closely with cross-functional engineering teams to validate bug fixes and verify compliance with technical specifications.\n• Managed Tier-1/Tier-2 technical support inquiries in a fast-paced environment, translating user feedback into actionable bug reports and product improvements.",
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
