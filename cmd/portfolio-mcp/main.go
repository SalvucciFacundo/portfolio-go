package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/SalvucciFacundo/portfolio-go/internal/domain"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	portfolioURL := os.Getenv("PORTFOLIO_URL")
	if portfolioURL == "" {
		portfolioURL = "http://localhost:8080"
	}
	adminPassword := os.Getenv("PORTFOLIO_ADMIN_PASSWORD")
	if adminPassword == "" {
		fmt.Fprintln(os.Stderr, "ADVERTENCIA: PORTFOLIO_ADMIN_PASSWORD no está configurada. Las operaciones que requieran login fallarán.")
	}

	client, err := NewClient(portfolioURL, adminPassword)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error al inicializar cliente de API: %v\n", err)
		os.Exit(1)
	}

	s := server.NewMCPServer(
		"portfolio-mcp",
		"1.0.0",
		server.WithToolCapabilities(true),
		server.WithDescription("Servidor MCP para gestionar secciones y datos del portafolio personal"),
	)

	registerProfileTools(s, client)
	registerSkillsTools(s, client)
	registerProjectsTools(s, client)
	registerExperienceTools(s, client)
	registerEducationTools(s, client)

	if err := server.ServeStdio(s); err != nil {
		fmt.Fprintf(os.Stderr, "Error en el servidor MCP: %v\n", err)
		os.Exit(1)
	}
}

func jsonResult(v any) (*mcp.CallToolResult, error) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("error al serializar resultado: %v", err)), nil
	}
	return mcp.NewToolResultText(string(data)), nil
}

// ---- PROFILE & SOCIALS ----

func registerProfileTools(s *server.MCPServer, client *Client) {
	// get_profile
	s.AddTool(
		mcp.NewTool(
			"get_profile",
			mcp.WithDescription("Devuelve el perfil completo del portafolio (nombre, bio, redes, resumen en ES/EN, avatar_url, resume_url)"),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			profile, err := client.GetProfile(ctx)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("error al obtener perfil: %v", err)), nil
			}
			return jsonResult(profile)
		},
	)

	// update_profile
	s.AddTool(
		mcp.NewTool(
			"update_profile",
			mcp.WithDescription("Actualiza los campos de texto del perfil. Solo se modifican los campos enviados."),
			mcp.WithString("name", mcp.Description("Nombre completo")),
			mcp.WithString("role_es", mcp.Description("Título profesional en español")),
			mcp.WithString("role_en", mcp.Description("Título profesional en inglés")),
			mcp.WithString("headline_es", mcp.Description("Titular / frase corta destacada en español")),
			mcp.WithString("headline_en", mcp.Description("Titular / frase corta destacada en inglés")),
			mcp.WithString("summary_es", mcp.Description("Biografía / resumen extendido en español")),
			mcp.WithString("summary_en", mcp.Description("Biografía / resumen extendido en inglés")),
			mcp.WithString("email", mcp.Description("Email de contacto")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var upd ProfileUpdate
			if v := req.GetString("name", ""); v != "" {
				upd.Name = &v
			}
			if v := req.GetString("role_es", ""); v != "" {
				upd.RoleEs = &v
			}
			if v := req.GetString("role_en", ""); v != "" {
				upd.RoleEn = &v
			}
			if v := req.GetString("headline_es", ""); v != "" {
				upd.HeadlineEs = &v
			}
			if v := req.GetString("headline_en", ""); v != "" {
				upd.HeadlineEn = &v
			}
			if v := req.GetString("summary_es", ""); v != "" {
				upd.SummaryEs = &v
			}
			if v := req.GetString("summary_en", ""); v != "" {
				upd.SummaryEn = &v
			}
			if v := req.GetString("email", ""); v != "" {
				upd.Email = &v
			}

			if err := client.UpdateProfile(ctx, upd); err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("error al actualizar perfil: %v", err)), nil
			}
			return mcp.NewToolResultText("Perfil actualizado correctamente."), nil
		},
	)

	// update_socials
	s.AddTool(
		mcp.NewTool(
			"update_socials",
			mcp.WithDescription("Reemplaza la lista completa de redes sociales del perfil"),
			mcp.WithArray(
				"socials",
				mcp.Required(),
				mcp.Description("Array de enlaces sociales con campos: name, url, icon_key (github, linkedin, etc.), position"),
			),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var args struct {
				Socials []domain.SocialLink `json:"socials"`
			}
			if err := req.BindArguments(&args); err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("argumentos inválidos: %v", err)), nil
			}

			if err := client.UpdateSocials(ctx, args.Socials); err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("error al actualizar redes sociales: %v", err)), nil
			}
			return mcp.NewToolResultText("Redes sociales actualizadas correctamente."), nil
		},
	)

	// upload_avatar
	s.AddTool(
		mcp.NewTool(
			"upload_avatar",
			mcp.WithDescription("Sube una imagen local como avatar del perfil y la almacena en Cloudinary"),
			mcp.WithString("file_path", mcp.Required(), mcp.Description("Ruta absoluta o relativa del archivo de imagen local")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			filePath, err := req.RequireString("file_path")
			if err != nil {
				return mcp.NewToolResultError("file_path es requerido"), nil
			}
			url, err := client.UploadAvatar(ctx, filePath)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("error al subir avatar: %v", err)), nil
			}
			return mcp.NewToolResultText(fmt.Sprintf("Avatar subido exitosamente: %s", url)), nil
		},
	)

	// upload_cv
	s.AddTool(
		mcp.NewTool(
			"upload_cv",
			mcp.WithDescription("Sube un archivo PDF local como CV/resume del perfil y lo almacena en Cloudinary"),
			mcp.WithString("file_path", mcp.Required(), mcp.Description("Ruta del archivo PDF local")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			filePath, err := req.RequireString("file_path")
			if err != nil {
				return mcp.NewToolResultError("file_path es requerido"), nil
			}
			url, err := client.UploadCV(ctx, filePath)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("error al subir CV: %v", err)), nil
			}
			return mcp.NewToolResultText(fmt.Sprintf("CV subido exitosamente: %s", url)), nil
		},
	)
}

// ---- SKILLS ----

func registerSkillsTools(s *server.MCPServer, client *Client) {
	// list_skills
	s.AddTool(
		mcp.NewTool(
			"list_skills",
			mcp.WithDescription("Devuelve la lista de habilidades / tecnologías ordenadas por posición"),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			skills, err := client.ListSkills(ctx)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("error al listar skills: %v", err)), nil
			}
			return jsonResult(skills)
		},
	)

	// create_skill
	s.AddTool(
		mcp.NewTool(
			"create_skill",
			mcp.WithDescription("Crea una nueva habilidad o tecnología en el portafolio"),
			mcp.WithString("name", mcp.Required(), mcp.Description("Nombre de la tecnología/habilidad (ej. Go, Docker, React)")),
			mcp.WithBoolean("is_tool", mcp.Description("true si es una herramienta (ej. Git, Figma, Docker), false si es lenguaje/framework")),
			mcp.WithNumber("position", mcp.Description("Posición para el orden de visualización")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			name, err := req.RequireString("name")
			if err != nil {
				return mcp.NewToolResultError("name es requerido"), nil
			}
			isTool := req.GetBool("is_tool", false)
			position := req.GetInt("position", 0)

			id, err := client.CreateSkill(ctx, domain.Skill{
				Name:     name,
				IsTool:   isTool,
				Position: position,
			})
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("error al crear skill: %v", err)), nil
			}
			return mcp.NewToolResultText(fmt.Sprintf("Skill creado exitosamente con ID %d", id)), nil
		},
	)

	// update_skill
	s.AddTool(
		mcp.NewTool(
			"update_skill",
			mcp.WithDescription("Actualiza una habilidad existente por su ID"),
			mcp.WithNumber("id", mcp.Required(), mcp.Description("ID numérico del skill")),
			mcp.WithString("name", mcp.Required(), mcp.Description("Nombre de la tecnología")),
			mcp.WithBoolean("is_tool", mcp.Description("Indica si es herramienta")),
			mcp.WithNumber("position", mcp.Description("Posición en la lista")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			id, err := req.RequireInt("id")
			if err != nil {
				return mcp.NewToolResultError("id es requerido"), nil
			}
			name, err := req.RequireString("name")
			if err != nil {
				return mcp.NewToolResultError("name es requerido"), nil
			}
			isTool := req.GetBool("is_tool", false)
			position := req.GetInt("position", 0)

			err = client.UpdateSkill(ctx, int64(id), domain.Skill{
				ID:       int64(id),
				Name:     name,
				IsTool:   isTool,
				Position: position,
			})
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("error al actualizar skill: %v", err)), nil
			}
			return mcp.NewToolResultText(fmt.Sprintf("Skill %d actualizado exitosamente", id)), nil
		},
	)

	// delete_skill
	s.AddTool(
		mcp.NewTool(
			"delete_skill",
			mcp.WithDescription("Elimina una habilidad del portafolio por su ID"),
			mcp.WithNumber("id", mcp.Required(), mcp.Description("ID del skill a eliminar")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			id, err := req.RequireInt("id")
			if err != nil {
				return mcp.NewToolResultError("id es requerido"), nil
			}
			if err := client.DeleteSkill(ctx, int64(id)); err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("error al eliminar skill: %v", err)), nil
			}
			return mcp.NewToolResultText(fmt.Sprintf("Skill %d eliminado exitosamente", id)), nil
		},
	)

	// upload_skill_icon
	s.AddTool(
		mcp.NewTool(
			"upload_skill_icon",
			mcp.WithDescription("Sube un ícono de imagen local para una habilidad específica"),
			mcp.WithNumber("id", mcp.Required(), mcp.Description("ID del skill")),
			mcp.WithString("file_path", mcp.Required(), mcp.Description("Ruta de la imagen local (SVG, PNG, WebP, etc.)")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			id, err := req.RequireInt("id")
			if err != nil {
				return mcp.NewToolResultError("id es requerido"), nil
			}
			filePath, err := req.RequireString("file_path")
			if err != nil {
				return mcp.NewToolResultError("file_path es requerido"), nil
			}

			url, err := client.UploadSkillIcon(ctx, int64(id), filePath)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("error al subir icono de skill: %v", err)), nil
			}
			return mcp.NewToolResultText(fmt.Sprintf("Icono de skill subido exitosamente: %s", url)), nil
		},
	)
}

// ---- PROJECTS ----

func registerProjectsTools(s *server.MCPServer, client *Client) {
	// list_projects
	s.AddTool(
		mcp.NewTool(
			"list_projects",
			mcp.WithDescription("Devuelve todos los proyectos del portafolio"),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			projects, err := client.ListProjects(ctx)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("error al listar proyectos: %v", err)), nil
			}
			return jsonResult(projects)
		},
	)

	// get_project
	s.AddTool(
		mcp.NewTool(
			"get_project",
			mcp.WithDescription("Devuelve el detalle de un proyecto específico con todas sus capturas de pantalla"),
			mcp.WithNumber("id", mcp.Required(), mcp.Description("ID del proyecto")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			id, err := req.RequireInt("id")
			if err != nil {
				return mcp.NewToolResultError("id es requerido"), nil
			}
			project, err := client.GetProject(ctx, int64(id))
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("error al obtener proyecto: %v", err)), nil
			}
			return jsonResult(project)
		},
	)

	// create_project
	s.AddTool(
		mcp.NewTool(
			"create_project",
			mcp.WithDescription("Crea un nuevo proyecto en el portafolio"),
			mcp.WithString("title_es", mcp.Required(), mcp.Description("Título en español")),
			mcp.WithString("title_en", mcp.Required(), mcp.Description("Título en inglés")),
			mcp.WithString("description_es", mcp.Description("Descripción corta en español")),
			mcp.WithString("description_en", mcp.Description("Descripción corta en inglés")),
			mcp.WithString("tech_description_es", mcp.Description("Descripción técnica detallada en español")),
			mcp.WithString("tech_description_en", mcp.Description("Descripción técnica detallada en inglés")),
			mcp.WithString("category", mcp.Description("Categoría (ej. Backend, Frontend, Full Stack, AI, CLI)")),
			mcp.WithString("status", mcp.Description("Estado del proyecto (production, development, demo)")),
			mcp.WithArray("tags", mcp.WithStringItems(), mcp.Description("Lista de tecnologías/tags (ej. [\"Go\", \"Docker\", \"PostgreSQL\"])")),
			mcp.WithString("link", mcp.Description("URL pública del proyecto en vivo")),
			mcp.WithString("repo_link", mcp.Description("URL del repositorio (GitHub, etc.)")),
			mcp.WithNumber("position", mcp.Description("Posición para ordenamiento")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			titleEs, err := req.RequireString("title_es")
			if err != nil {
				return mcp.NewToolResultError("title_es es requerido"), nil
			}
			titleEn, err := req.RequireString("title_en")
			if err != nil {
				return mcp.NewToolResultError("title_en es requerido"), nil
			}

			project := domain.Project{
				TitleEs:           titleEs,
				TitleEn:           titleEn,
				DescriptionEs:     req.GetString("description_es", ""),
				DescriptionEn:     req.GetString("description_en", ""),
				TechDescriptionEs: req.GetString("tech_description_es", ""),
				TechDescriptionEn: req.GetString("tech_description_en", ""),
				Category:          req.GetString("category", ""),
				Status:            req.GetString("status", "production"),
				Tags:              req.GetStringSlice("tags", []string{}),
				Link:              req.GetString("link", ""),
				RepoLink:          req.GetString("repo_link", ""),
				Position:          req.GetInt("position", 0),
			}

			id, err := client.CreateProject(ctx, project)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("error al crear proyecto: %v", err)), nil
			}
			return mcp.NewToolResultText(fmt.Sprintf("Proyecto creado exitosamente con ID %d", id)), nil
		},
	)

	// update_project
	s.AddTool(
		mcp.NewTool(
			"update_project",
			mcp.WithDescription("Actualiza un proyecto existente"),
			mcp.WithNumber("id", mcp.Required(), mcp.Description("ID numérico del proyecto")),
			mcp.WithString("title_es", mcp.Required(), mcp.Description("Título en español")),
			mcp.WithString("title_en", mcp.Required(), mcp.Description("Título en inglés")),
			mcp.WithString("description_es", mcp.Description("Descripción corta en español")),
			mcp.WithString("description_en", mcp.Description("Descripción corta en inglés")),
			mcp.WithString("tech_description_es", mcp.Description("Descripción técnica en español")),
			mcp.WithString("tech_description_en", mcp.Description("Descripción técnica en inglés")),
			mcp.WithString("category", mcp.Description("Categoría")),
			mcp.WithString("status", mcp.Description("Estado (production, development, demo)")),
			mcp.WithArray("tags", mcp.WithStringItems(), mcp.Description("Tags / tecnologías")),
			mcp.WithString("link", mcp.Description("URL del proyecto")),
			mcp.WithString("repo_link", mcp.Description("URL del repositorio")),
			mcp.WithNumber("position", mcp.Description("Posición")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			id, err := req.RequireInt("id")
			if err != nil {
				return mcp.NewToolResultError("id es requerido"), nil
			}
			titleEs, err := req.RequireString("title_es")
			if err != nil {
				return mcp.NewToolResultError("title_es es requerido"), nil
			}
			titleEn, err := req.RequireString("title_en")
			if err != nil {
				return mcp.NewToolResultError("title_en es requerido"), nil
			}

			project := domain.Project{
				ID:                int64(id),
				TitleEs:           titleEs,
				TitleEn:           titleEn,
				DescriptionEs:     req.GetString("description_es", ""),
				DescriptionEn:     req.GetString("description_en", ""),
				TechDescriptionEs: req.GetString("tech_description_es", ""),
				TechDescriptionEn: req.GetString("tech_description_en", ""),
				Category:          req.GetString("category", ""),
				Status:            req.GetString("status", "production"),
				Tags:              req.GetStringSlice("tags", []string{}),
				Link:              req.GetString("link", ""),
				RepoLink:          req.GetString("repo_link", ""),
				Position:          req.GetInt("position", 0),
			}

			if err := client.UpdateProject(ctx, int64(id), project); err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("error al actualizar proyecto: %v", err)), nil
			}
			return mcp.NewToolResultText(fmt.Sprintf("Proyecto %d actualizado exitosamente", id)), nil
		},
	)

	// reorder_project
	s.AddTool(
		mcp.NewTool(
			"reorder_project",
			mcp.WithDescription("Cambia la posición de visualización de un proyecto"),
			mcp.WithNumber("id", mcp.Required(), mcp.Description("ID del proyecto")),
			mcp.WithNumber("position", mcp.Required(), mcp.Description("Nueva posición numérica")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			id, err := req.RequireInt("id")
			if err != nil {
				return mcp.NewToolResultError("id es requerido"), nil
			}
			pos, err := req.RequireInt("position")
			if err != nil {
				return mcp.NewToolResultError("position es requerido"), nil
			}

			if err := client.ReorderProject(ctx, int64(id), pos); err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("error al reordenar proyecto: %v", err)), nil
			}
			return mcp.NewToolResultText(fmt.Sprintf("Proyecto %d movido a la posición %d", id, pos)), nil
		},
	)

	// delete_project
	s.AddTool(
		mcp.NewTool(
			"delete_project",
			mcp.WithDescription("Elimina un proyecto del portafolio por su ID"),
			mcp.WithNumber("id", mcp.Required(), mcp.Description("ID del proyecto")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			id, err := req.RequireInt("id")
			if err != nil {
				return mcp.NewToolResultError("id es requerido"), nil
			}
			if err := client.DeleteProject(ctx, int64(id)); err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("error al eliminar proyecto: %v", err)), nil
			}
			return mcp.NewToolResultText(fmt.Sprintf("Proyecto %d eliminado exitosamente", id)), nil
		},
	)

	// upload_project_cover
	s.AddTool(
		mcp.NewTool(
			"upload_project_cover",
			mcp.WithDescription("Sube una imagen local como portada (cover) de un proyecto"),
			mcp.WithNumber("id", mcp.Required(), mcp.Description("ID del proyecto")),
			mcp.WithString("file_path", mcp.Required(), mcp.Description("Ruta de la imagen local")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			id, err := req.RequireInt("id")
			if err != nil {
				return mcp.NewToolResultError("id es requerido"), nil
			}
			filePath, err := req.RequireString("file_path")
			if err != nil {
				return mcp.NewToolResultError("file_path es requerido"), nil
			}

			url, err := client.UploadProjectCover(ctx, int64(id), filePath)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("error al subir portada: %v", err)), nil
			}
			return mcp.NewToolResultText(fmt.Sprintf("Portada subida exitosamente: %s", url)), nil
		},
	)

	// upload_project_screenshots
	s.AddTool(
		mcp.NewTool(
			"upload_project_screenshots",
			mcp.WithDescription("Sube una o varias imágenes locales como capturas de pantalla de un proyecto"),
			mcp.WithNumber("id", mcp.Required(), mcp.Description("ID del proyecto")),
			mcp.WithArray("file_paths", mcp.Required(), mcp.WithStringItems(), mcp.Description("Lista de rutas locales de imágenes")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			id, err := req.RequireInt("id")
			if err != nil {
				return mcp.NewToolResultError("id es requerido"), nil
			}
			filePaths := req.GetStringSlice("file_paths", []string{})
			if len(filePaths) == 0 {
				return mcp.NewToolResultError("file_paths no puede estar vacío"), nil
			}

			urls, err := client.UploadProjectScreenshots(ctx, int64(id), filePaths)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("error al subir capturas: %v", err)), nil
			}
			return jsonResult(map[string]any{
				"message": "Capturas subidas exitosamente",
				"urls":    urls,
			})
		},
	)

	// delete_project_screenshot
	s.AddTool(
		mcp.NewTool(
			"delete_project_screenshot",
			mcp.WithDescription("Elimina una captura de pantalla específica de un proyecto"),
			mcp.WithNumber("project_id", mcp.Required(), mcp.Description("ID del proyecto")),
			mcp.WithNumber("image_id", mcp.Required(), mcp.Description("ID numérico de la imagen a eliminar")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			projectID, err := req.RequireInt("project_id")
			if err != nil {
				return mcp.NewToolResultError("project_id es requerido"), nil
			}
			imageID, err := req.RequireInt("image_id")
			if err != nil {
				return mcp.NewToolResultError("image_id es requerido"), nil
			}

			if err := client.DeleteProjectScreenshot(ctx, int64(projectID), int64(imageID)); err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("error al eliminar screenshot: %v", err)), nil
			}
			return mcp.NewToolResultText(fmt.Sprintf("Screenshot %d eliminada exitosamente del proyecto %d", imageID, projectID)), nil
		},
	)
}

// ---- EXPERIENCE ----

func registerExperienceTools(s *server.MCPServer, client *Client) {
	// list_experience
	s.AddTool(
		mcp.NewTool(
			"list_experience",
			mcp.WithDescription("Devuelve la lista de experiencias laborales"),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			items, err := client.ListExperience(ctx)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("error al listar experiencias: %v", err)), nil
			}
			return jsonResult(items)
		},
	)

	// create_experience
	s.AddTool(
		mcp.NewTool(
			"create_experience",
			mcp.WithDescription("Crea una nueva experiencia laboral en el portafolio"),
			mcp.WithString("company", mcp.Required(), mcp.Description("Nombre de la empresa u organización")),
			mcp.WithString("position_es", mcp.Required(), mcp.Description("Puesto / rol en español")),
			mcp.WithString("position_en", mcp.Required(), mcp.Description("Puesto / rol en inglés")),
			mcp.WithString("period_es", mcp.Description("Período en español (ej. '2023 - Presente')")),
			mcp.WithString("period_en", mcp.Description("Período en inglés (ej. '2023 - Present')")),
			mcp.WithString("description_es", mcp.Description("Descripción de tareas y logros en español")),
			mcp.WithString("description_en", mcp.Description("Descripción de tareas y logros en inglés")),
			mcp.WithNumber("position", mcp.Description("Posición para ordenamiento")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			company, err := req.RequireString("company")
			if err != nil {
				return mcp.NewToolResultError("company es requerido"), nil
			}
			posEs, err := req.RequireString("position_es")
			if err != nil {
				return mcp.NewToolResultError("position_es es requerido"), nil
			}
			posEn, err := req.RequireString("position_en")
			if err != nil {
				return mcp.NewToolResultError("position_en es requerido"), nil
			}

			id, err := client.CreateExperience(ctx, domain.Experience{
				Company:       company,
				PositionEs:    posEs,
				PositionEn:    posEn,
				PeriodEs:      req.GetString("period_es", ""),
				PeriodEn:      req.GetString("period_en", ""),
				DescriptionEs: req.GetString("description_es", ""),
				DescriptionEn: req.GetString("description_en", ""),
				Position:      req.GetInt("position", 0),
			})
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("error al crear experiencia: %v", err)), nil
			}
			return mcp.NewToolResultText(fmt.Sprintf("Experiencia creada exitosamente con ID %d", id)), nil
		},
	)

	// update_experience
	s.AddTool(
		mcp.NewTool(
			"update_experience",
			mcp.WithDescription("Actualiza una experiencia laboral existente"),
			mcp.WithNumber("id", mcp.Required(), mcp.Description("ID de la experiencia")),
			mcp.WithString("company", mcp.Required(), mcp.Description("Empresa")),
			mcp.WithString("position_es", mcp.Required(), mcp.Description("Puesto en español")),
			mcp.WithString("position_en", mcp.Required(), mcp.Description("Puesto en inglés")),
			mcp.WithString("period_es", mcp.Description("Período en español")),
			mcp.WithString("period_en", mcp.Description("Período en inglés")),
			mcp.WithString("description_es", mcp.Description("Descripción en español")),
			mcp.WithString("description_en", mcp.Description("Descripción en inglés")),
			mcp.WithNumber("position", mcp.Description("Posición")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			id, err := req.RequireInt("id")
			if err != nil {
				return mcp.NewToolResultError("id es requerido"), nil
			}
			company, err := req.RequireString("company")
			if err != nil {
				return mcp.NewToolResultError("company es requerido"), nil
			}
			posEs, err := req.RequireString("position_es")
			if err != nil {
				return mcp.NewToolResultError("position_es es requerido"), nil
			}
			posEn, err := req.RequireString("position_en")
			if err != nil {
				return mcp.NewToolResultError("position_en es requerido"), nil
			}

			err = client.UpdateExperience(ctx, int64(id), domain.Experience{
				ID:            int64(id),
				Company:       company,
				PositionEs:    posEs,
				PositionEn:    posEn,
				PeriodEs:      req.GetString("period_es", ""),
				PeriodEn:      req.GetString("period_en", ""),
				DescriptionEs: req.GetString("description_es", ""),
				DescriptionEn: req.GetString("description_en", ""),
				Position:      req.GetInt("position", 0),
			})
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("error al actualizar experiencia: %v", err)), nil
			}
			return mcp.NewToolResultText(fmt.Sprintf("Experiencia %d actualizada exitosamente", id)), nil
		},
	)

	// delete_experience
	s.AddTool(
		mcp.NewTool(
			"delete_experience",
			mcp.WithDescription("Elimina una experiencia laboral por su ID"),
			mcp.WithNumber("id", mcp.Required(), mcp.Description("ID de la experiencia")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			id, err := req.RequireInt("id")
			if err != nil {
				return mcp.NewToolResultError("id es requerido"), nil
			}
			if err := client.DeleteExperience(ctx, int64(id)); err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("error al eliminar experiencia: %v", err)), nil
			}
			return mcp.NewToolResultText(fmt.Sprintf("Experiencia %d eliminada exitosamente", id)), nil
		},
	)
}

// ---- EDUCATION ----

func registerEducationTools(s *server.MCPServer, client *Client) {
	// list_education
	s.AddTool(
		mcp.NewTool(
			"list_education",
			mcp.WithDescription("Devuelve la lista de educación y certificaciones"),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			items, err := client.ListEducation(ctx)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("error al listar educación: %v", err)), nil
			}
			return jsonResult(items)
		},
	)

	// create_education
	s.AddTool(
		mcp.NewTool(
			"create_education",
			mcp.WithDescription("Crea un nuevo estudio, título universitario o curso"),
			mcp.WithString("title_es", mcp.Required(), mcp.Description("Título del curso o carrera en español")),
			mcp.WithString("title_en", mcp.Required(), mcp.Description("Título del curso o carrera en inglés")),
			mcp.WithString("school", mcp.Description("Institución / Universidad / Plataforma")),
			mcp.WithString("date", mcp.Description("Fecha o período (ej. '2020 - 2024')")),
			mcp.WithBoolean("is_course", mcp.Description("true si es un curso/certificación corta, false si es carrera de grado")),
			mcp.WithString("description_es", mcp.Description("Descripción en español")),
			mcp.WithString("description_en", mcp.Description("Descripción en inglés")),
			mcp.WithNumber("position", mcp.Description("Posición para ordenamiento")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			titleEs, err := req.RequireString("title_es")
			if err != nil {
				return mcp.NewToolResultError("title_es es requerido"), nil
			}
			titleEn, err := req.RequireString("title_en")
			if err != nil {
				return mcp.NewToolResultError("title_en es requerido"), nil
			}

			id, err := client.CreateEducation(ctx, domain.Education{
				TitleEs:       titleEs,
				TitleEn:       titleEn,
				School:        req.GetString("school", ""),
				Date:          req.GetString("date", ""),
				IsCourse:      req.GetBool("is_course", false),
				DescriptionEs: req.GetString("description_es", ""),
				DescriptionEn: req.GetString("description_en", ""),
				Position:      req.GetInt("position", 0),
			})
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("error al crear educación: %v", err)), nil
			}
			return mcp.NewToolResultText(fmt.Sprintf("Educación creada exitosamente con ID %d", id)), nil
		},
	)

	// update_education
	s.AddTool(
		mcp.NewTool(
			"update_education",
			mcp.WithDescription("Actualiza un estudio o curso existente"),
			mcp.WithNumber("id", mcp.Required(), mcp.Description("ID de la educación")),
			mcp.WithString("title_es", mcp.Required(), mcp.Description("Título en español")),
			mcp.WithString("title_en", mcp.Required(), mcp.Description("Título en inglés")),
			mcp.WithString("school", mcp.Description("Institución")),
			mcp.WithString("date", mcp.Description("Fecha o período")),
			mcp.WithBoolean("is_course", mcp.Description("Indica si es curso")),
			mcp.WithString("description_es", mcp.Description("Descripción en español")),
			mcp.WithString("description_en", mcp.Description("Descripción en inglés")),
			mcp.WithNumber("position", mcp.Description("Posición")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			id, err := req.RequireInt("id")
			if err != nil {
				return mcp.NewToolResultError("id es requerido"), nil
			}
			titleEs, err := req.RequireString("title_es")
			if err != nil {
				return mcp.NewToolResultError("title_es es requerido"), nil
			}
			titleEn, err := req.RequireString("title_en")
			if err != nil {
				return mcp.NewToolResultError("title_en es requerido"), nil
			}

			err = client.UpdateEducation(ctx, int64(id), domain.Education{
				ID:            int64(id),
				TitleEs:       titleEs,
				TitleEn:       titleEn,
				School:        req.GetString("school", ""),
				Date:          req.GetString("date", ""),
				IsCourse:      req.GetBool("is_course", false),
				DescriptionEs: req.GetString("description_es", ""),
				DescriptionEn: req.GetString("description_en", ""),
				Position:      req.GetInt("position", 0),
			})
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("error al actualizar educación: %v", err)), nil
			}
			return mcp.NewToolResultText(fmt.Sprintf("Educación %d actualizada exitosamente", id)), nil
		},
	)

	// delete_education
	s.AddTool(
		mcp.NewTool(
			"delete_education",
			mcp.WithDescription("Elimina un estudio o curso por su ID"),
			mcp.WithNumber("id", mcp.Required(), mcp.Description("ID de la educación")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			id, err := req.RequireInt("id")
			if err != nil {
				return mcp.NewToolResultError("id es requerido"), nil
			}
			if err := client.DeleteEducation(ctx, int64(id)); err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("error al eliminar educación: %v", err)), nil
			}
			return mcp.NewToolResultText(fmt.Sprintf("Educación %d eliminada exitosamente", id)), nil
		},
	)
}
