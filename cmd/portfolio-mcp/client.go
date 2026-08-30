package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/SalvucciFacundo/portfolio-go/internal/domain"
)

const userAgent = "portfolio-mcp/1.0"

// ProfileUpdate representa los campos opcionales del perfil para UpdateProfile.
type ProfileUpdate struct {
	Name       *string `json:"name,omitempty"`
	RoleEs     *string `json:"role_es,omitempty"`
	RoleEn     *string `json:"role_en,omitempty"`
	HeadlineEs *string `json:"headline_es,omitempty"`
	HeadlineEn *string `json:"headline_en,omitempty"`
	SummaryEs  *string `json:"summary_es,omitempty"`
	SummaryEn  *string `json:"summary_en,omitempty"`
	Email      *string `json:"email,omitempty"`
}

// Client gestiona la comunicación HTTP con la API REST del portafolio.
type Client struct {
	baseURL    string
	password   string
	httpClient *http.Client
	csrfToken  string
	mu         sync.Mutex
	loggedIn   bool
}

// NewClient inicializa un cliente con cookie jar para mantener la sesión de admin.
func NewClient(baseURL, password string) (*Client, error) {
	baseURL = strings.TrimRight(baseURL, "/")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("crear cookie jar: %w", err)
	}

	return &Client{
		baseURL:  baseURL,
		password: password,
		httpClient: &http.Client{
			Jar:     jar,
			Timeout: 30 * time.Second,
		},
	}, nil
}

// EnsureLogin autentica contra /api/v1/auth/login si no hay una sesión activa.
func (c *Client) EnsureLogin(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.loginLocked(ctx)
}

func (c *Client) loginLocked(ctx context.Context) error {
	if c.password == "" {
		return errors.New("PORTFOLIO_ADMIN_PASSWORD no está configurada en las variables de entorno")
	}
	loginURL := c.baseURL + "/api/v1/auth/login"
	payload, _ := json.Marshal(map[string]string{"password": c.password})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, loginURL, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("crear request de login: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", userAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("conectar con la API (%s): %w", loginURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp struct {
			Error string `json:"error"`
		}
		_ = json.NewDecoder(resp.Body).Decode(&errResp)
		if errResp.Error != "" {
			return fmt.Errorf("error de autenticación (%d): %s", resp.StatusCode, errResp.Error)
		}
		return fmt.Errorf("error de autenticación con status %d", resp.StatusCode)
	}

	// Extraer el token CSRF de las cookies devueltas
	u, _ := url.Parse(c.baseURL)
	for _, cookie := range c.httpClient.Jar.Cookies(u) {
		if cookie.Name == "csrf_token" {
			c.csrfToken = cookie.Value
			break
		}
	}

	c.loggedIn = true
	return nil
}

func (c *Client) doRequest(ctx context.Context, method, endpoint string, body any, target any, isRetry bool) error {
	fullURL := c.baseURL + endpoint

	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("serializar body: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, fullURL, bodyReader)
	if err != nil {
		return fmt.Errorf("crear request: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("User-Agent", userAgent)

	// Métodos de mutación requieren autenticación y CSRF
	if method == http.MethodPost || method == http.MethodPut || method == http.MethodDelete || method == http.MethodPatch {
		c.mu.Lock()
		if !c.loggedIn {
			if err := c.loginLocked(ctx); err != nil {
				c.mu.Unlock()
				return fmt.Errorf("login requerido: %w", err)
			}
		}
		csrf := c.csrfToken
		c.mu.Unlock()

		if csrf != "" {
			req.Header.Set("X-CSRF-Token", csrf)
		}
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("ejecutar request a %s: %w", fullURL, err)
	}
	defer resp.Body.Close()

	// Si expiró la sesión o token CSRF inválido, reintentar login una vez
	if (resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden) && !isRetry {
		c.mu.Lock()
		c.loggedIn = false
		c.csrfToken = ""
		loginErr := c.loginLocked(ctx)
		c.mu.Unlock()

		if loginErr == nil {
			return c.doRequest(ctx, method, endpoint, body, target, true)
		}
	}

	if resp.StatusCode >= 400 {
		var errResp struct {
			Error string `json:"error"`
		}
		_ = json.NewDecoder(resp.Body).Decode(&errResp)
		if errResp.Error != "" {
			return fmt.Errorf("API error (%d): %s", resp.StatusCode, errResp.Error)
		}
		return fmt.Errorf("API error con status %d", resp.StatusCode)
	}

	if target != nil && resp.StatusCode != http.StatusNoContent {
		if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
			return fmt.Errorf("decodificar respuesta JSON: %w", err)
		}
	}

	return nil
}

func (c *Client) doUpload(ctx context.Context, endpoint, fieldName, filePath string, extraFields map[string]string, target any) error {
	c.mu.Lock()
	if !c.loggedIn {
		if err := c.loginLocked(ctx); err != nil {
			c.mu.Unlock()
			return fmt.Errorf("login requerido: %w", err)
		}
	}
	csrf := c.csrfToken
	c.mu.Unlock()

	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("abrir archivo local %q: %w", filePath, err)
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile(fieldName, filepath.Base(filePath))
	if err != nil {
		return fmt.Errorf("crear form file: %w", err)
	}
	if _, err := io.Copy(part, file); err != nil {
		return fmt.Errorf("copiar contenido del archivo: %w", err)
	}

	for k, v := range extraFields {
		if err := writer.WriteField(k, v); err != nil {
			return fmt.Errorf("escribir campo %s: %w", k, err)
		}
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("cerrar multipart writer: %w", err)
	}

	fullURL := c.baseURL + endpoint
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fullURL, body)
	if err != nil {
		return fmt.Errorf("crear upload request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("User-Agent", userAgent)
	if csrf != "" {
		req.Header.Set("X-CSRF-Token", csrf)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("ejecutar upload a %s: %w", fullURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var errResp struct {
			Error string `json:"error"`
		}
		_ = json.NewDecoder(resp.Body).Decode(&errResp)
		if errResp.Error != "" {
			return fmt.Errorf("upload error (%d): %s", resp.StatusCode, errResp.Error)
		}
		return fmt.Errorf("upload error con status %d", resp.StatusCode)
	}

	if target != nil && resp.StatusCode != http.StatusNoContent {
		if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
			return fmt.Errorf("decodificar respuesta de upload: %w", err)
		}
	}

	return nil
}

// ---- Profile & Socials ----

func (c *Client) GetProfile(ctx context.Context) (*domain.Profile, error) {
	var profile domain.Profile
	err := c.doRequest(ctx, http.MethodGet, "/api/v1/profile", nil, &profile, false)
	return &profile, err
}

func (c *Client) UpdateProfile(ctx context.Context, upd ProfileUpdate) error {
	var resp map[string]bool
	return c.doRequest(ctx, http.MethodPut, "/api/v1/profile", upd, &resp, false)
}

func (c *Client) UpdateSocials(ctx context.Context, socials []domain.SocialLink) error {
	var resp map[string]bool
	return c.doRequest(ctx, http.MethodPut, "/api/v1/socials", socials, &resp, false)
}

func (c *Client) UploadAvatar(ctx context.Context, filePath string) (string, error) {
	var resp struct {
		AvatarURL string `json:"avatar_url"`
	}
	err := c.doUpload(ctx, "/api/v1/profile/avatar", "avatar", filePath, nil, &resp)
	return resp.AvatarURL, err
}

func (c *Client) UploadCV(ctx context.Context, filePath string) (string, error) {
	var resp struct {
		ResumeURL string `json:"resume_url"`
	}
	err := c.doUpload(ctx, "/api/v1/profile/cv", "cv", filePath, nil, &resp)
	return resp.ResumeURL, err
}

// ---- Skills ----

func (c *Client) ListSkills(ctx context.Context) ([]domain.Skill, error) {
	var skills []domain.Skill
	err := c.doRequest(ctx, http.MethodGet, "/api/v1/skills", nil, &skills, false)
	return skills, err
}

func (c *Client) CreateSkill(ctx context.Context, skill domain.Skill) (int64, error) {
	var resp struct {
		ID int64 `json:"id"`
	}
	err := c.doRequest(ctx, http.MethodPost, "/api/v1/skills", skill, &resp, false)
	return resp.ID, err
}

func (c *Client) UpdateSkill(ctx context.Context, id int64, skill domain.Skill) error {
	var resp struct {
		ID int64 `json:"id"`
	}
	return c.doRequest(ctx, http.MethodPut, "/api/v1/skills/"+strconv.FormatInt(id, 10), skill, &resp, false)
}

func (c *Client) DeleteSkill(ctx context.Context, id int64) error {
	return c.doRequest(ctx, http.MethodDelete, "/api/v1/skills/"+strconv.FormatInt(id, 10), nil, nil, false)
}

func (c *Client) UploadSkillIcon(ctx context.Context, id int64, filePath string) (string, error) {
	var resp struct {
		IconURL string `json:"icon_url"`
	}
	err := c.doUpload(ctx, "/api/v1/skills/"+strconv.FormatInt(id, 10)+"/icon", "icon", filePath, nil, &resp)
	return resp.IconURL, err
}

// ---- Projects ----

func (c *Client) ListProjects(ctx context.Context) ([]domain.Project, error) {
	var projects []domain.Project
	err := c.doRequest(ctx, http.MethodGet, "/api/v1/projects", nil, &projects, false)
	return projects, err
}

func (c *Client) GetProject(ctx context.Context, id int64) (*domain.Project, error) {
	var project domain.Project
	err := c.doRequest(ctx, http.MethodGet, "/api/v1/projects/"+strconv.FormatInt(id, 10), nil, &project, false)
	return &project, err
}

func (c *Client) CreateProject(ctx context.Context, project domain.Project) (int64, error) {
	var resp struct {
		ID int64 `json:"id"`
	}
	err := c.doRequest(ctx, http.MethodPost, "/api/v1/projects", project, &resp, false)
	return resp.ID, err
}

func (c *Client) UpdateProject(ctx context.Context, id int64, project domain.Project) error {
	var resp map[string]bool
	return c.doRequest(ctx, http.MethodPut, "/api/v1/projects/"+strconv.FormatInt(id, 10), project, &resp, false)
}

func (c *Client) ReorderProject(ctx context.Context, id int64, newPosition int) error {
	payload := map[string]int{"position": newPosition}
	return c.doRequest(ctx, http.MethodPut, "/api/v1/projects/"+strconv.FormatInt(id, 10)+"/position", payload, nil, false)
}

func (c *Client) DeleteProject(ctx context.Context, id int64) error {
	return c.doRequest(ctx, http.MethodDelete, "/api/v1/projects/"+strconv.FormatInt(id, 10), nil, nil, false)
}

func (c *Client) UploadProjectCover(ctx context.Context, id int64, filePath string) (string, error) {
	var resp struct {
		CoverURL string `json:"cover_url"`
	}
	err := c.doUpload(ctx, "/api/v1/projects/"+strconv.FormatInt(id, 10)+"/cover", "cover", filePath, nil, &resp)
	return resp.CoverURL, err
}

func (c *Client) UploadProjectScreenshots(ctx context.Context, projectID int64, filePaths []string) ([]string, error) {
	c.mu.Lock()
	if !c.loggedIn {
		if err := c.loginLocked(ctx); err != nil {
			c.mu.Unlock()
			return nil, fmt.Errorf("login requerido: %w", err)
		}
	}
	csrf := c.csrfToken
	c.mu.Unlock()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	for _, fp := range filePaths {
		file, err := os.Open(fp)
		if err != nil {
			return nil, fmt.Errorf("abrir screenshot %q: %w", fp, err)
		}
		part, err := writer.CreateFormFile("screenshots", filepath.Base(fp))
		if err != nil {
			file.Close()
			return nil, fmt.Errorf("crear form file para %q: %w", fp, err)
		}
		_, copyErr := io.Copy(part, file)
		file.Close()
		if copyErr != nil {
			return nil, fmt.Errorf("copiar archivo %q: %w", fp, copyErr)
		}
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("cerrar multipart writer: %w", err)
	}

	fullURL := fmt.Sprintf("%s/api/v1/projects/%d/images", c.baseURL, projectID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fullURL, body)
	if err != nil {
		return nil, fmt.Errorf("crear upload screenshots request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("User-Agent", userAgent)
	if csrf != "" {
		req.Header.Set("X-CSRF-Token", csrf)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("upload screenshots: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var errResp struct {
			Error string `json:"error"`
		}
		_ = json.NewDecoder(resp.Body).Decode(&errResp)
		if errResp.Error != "" {
			return nil, fmt.Errorf("upload screenshots error (%d): %s", resp.StatusCode, errResp.Error)
		}
		return nil, fmt.Errorf("upload screenshots error con status %d", resp.StatusCode)
	}

	var result struct {
		Images []string `json:"images"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&result)
	return result.Images, nil
}

func (c *Client) DeleteProjectScreenshot(ctx context.Context, projectID, imageID int64) error {
	endpoint := fmt.Sprintf("/api/v1/projects/%d/images/%d", projectID, imageID)
	return c.doRequest(ctx, http.MethodDelete, endpoint, nil, nil, false)
}

// ---- Experience ----

func (c *Client) ListExperience(ctx context.Context) ([]domain.Experience, error) {
	var items []domain.Experience
	err := c.doRequest(ctx, http.MethodGet, "/api/v1/experience", nil, &items, false)
	return items, err
}

func (c *Client) CreateExperience(ctx context.Context, item domain.Experience) (int64, error) {
	var resp struct {
		ID int64 `json:"id"`
	}
	err := c.doRequest(ctx, http.MethodPost, "/api/v1/experience", item, &resp, false)
	return resp.ID, err
}

func (c *Client) UpdateExperience(ctx context.Context, id int64, item domain.Experience) error {
	var resp struct {
		ID int64 `json:"id"`
	}
	return c.doRequest(ctx, http.MethodPut, "/api/v1/experience/"+strconv.FormatInt(id, 10), item, &resp, false)
}

func (c *Client) DeleteExperience(ctx context.Context, id int64) error {
	return c.doRequest(ctx, http.MethodDelete, "/api/v1/experience/"+strconv.FormatInt(id, 10), nil, nil, false)
}

// ---- Education ----

func (c *Client) ListEducation(ctx context.Context) ([]domain.Education, error) {
	var items []domain.Education
	err := c.doRequest(ctx, http.MethodGet, "/api/v1/education", nil, &items, false)
	return items, err
}

func (c *Client) CreateEducation(ctx context.Context, item domain.Education) (int64, error) {
	var resp struct {
		ID int64 `json:"id"`
	}
	err := c.doRequest(ctx, http.MethodPost, "/api/v1/education", item, &resp, false)
	return resp.ID, err
}

func (c *Client) UpdateEducation(ctx context.Context, id int64, item domain.Education) error {
	var resp struct {
		ID int64 `json:"id"`
	}
	return c.doRequest(ctx, http.MethodPut, "/api/v1/education/"+strconv.FormatInt(id, 10), item, &resp, false)
}

func (c *Client) DeleteEducation(ctx context.Context, id int64) error {
	return c.doRequest(ctx, http.MethodDelete, "/api/v1/education/"+strconv.FormatInt(id, 10), nil, nil, false)
}
