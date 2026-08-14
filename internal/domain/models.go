package domain

// Profile is the root domain entity for the portfolio. Editable text fields
// carry both languages (Es/En) and the view picks one based on the active lang.
type Profile struct {
	ID             int64        `json:"id"`
	Name           string       `json:"name"`
	RoleEs         string       `json:"role_es"`
	RoleEn         string       `json:"role_en"`
	HeadlineEs     string       `json:"headline_es"`
	HeadlineEn     string       `json:"headline_en"`
	SummaryEs      string       `json:"summary_es"`
	SummaryEn      string       `json:"summary_en"`
	Email          string       `json:"email"`
	AvatarURL      string       `json:"avatar_url"`
	ResumeURL      string       `json:"resume_url"`
	ResumeFilename string       `json:"resume_filename"`
	Socials        []SocialLink `json:"socials"`
	Skills         []Skill      `json:"skills"`
	Projects       []Project    `json:"projects"`
	Experience     []Experience `json:"experience"`
	Education      []Education  `json:"education"`
}

type SocialLink struct {
	ID       int64  `json:"id"`
	Position int    `json:"position"`
	Name     string `json:"name"`
	URL      string `json:"url"`
	IconKey  string `json:"icon_key"`
}

type Skill struct {
	ID       int64  `json:"id"`
	Position int    `json:"position"`
	Name     string `json:"name"`
	IconURL  string `json:"icon_url"`
	IsTool   bool   `json:"is_tool"`
}

type Project struct {
	ID                int64    `json:"id"`
	Position          int      `json:"position"`
	TitleEs           string   `json:"title_es"`
	TitleEn           string   `json:"title_en"`
	DescriptionEs     string   `json:"description_es"`
	DescriptionEn     string   `json:"description_en"`
	TechDescriptionEs string   `json:"tech_description_es"`
	TechDescriptionEn string   `json:"tech_description_en"`
	Category          string   `json:"category"`
	Tags              []string `json:"tags"`
	Link              string   `json:"link"`
	RepoLink          string   `json:"repo_link"`
	CoverURL          string   `json:"cover_url"`
	Screenshots       []string `json:"screenshots"`
}

type Experience struct {
	ID            int64  `json:"id"`
	Position      int    `json:"position"`
	PeriodEs      string `json:"period_es"`
	PeriodEn      string `json:"period_en"`
	PositionEs    string `json:"position_es"`
	PositionEn    string `json:"position_en"`
	Company       string `json:"company"`
	DescriptionEs string `json:"description_es"`
	DescriptionEn string `json:"description_en"`
}

type Education struct {
	ID            int64  `json:"id"`
	Position      int    `json:"position"`
	TitleEs       string `json:"title_es"`
	TitleEn       string `json:"title_en"`
	School        string `json:"school"`
	Date          string `json:"date"`
	IsCourse      bool   `json:"is_course"`
	DescriptionEs string `json:"description_es"`
	DescriptionEn string `json:"description_en"`
}

// AdminUser es la cuenta del dueño (única). PasswordHash nunca se serializa.
type AdminUser struct {
	ID           int64  `json:"-"`
	Username     string `json:"username"`
	PasswordHash string `json:"-"`
}

// Session es una sesión de admin. TokenHash (SHA-256) nunca se serializa.
type Session struct {
	ID            int64  `json:"-"`
	TokenHash     string `json:"-"`
	AdminUserID   int64  `json:"-"`
	UserAgentHash string `json:"-"`
	ExpiresAt     int64  `json:"-"` // unix seconds
}
