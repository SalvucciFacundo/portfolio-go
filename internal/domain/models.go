package domain

// Profile is the root domain entity for the portfolio. Editable text fields
// carry both languages (Es/En) and the view picks one based on the active lang.
type Profile struct {
	Name       string       `json:"name"`
	RoleEs     string       `json:"role_es"`
	RoleEn     string       `json:"role_en"`
	HeadlineEs string       `json:"headline_es"`
	HeadlineEn string       `json:"headline_en"`
	SummaryEs  string       `json:"summary_es"`
	SummaryEn  string       `json:"summary_en"`
	Email      string       `json:"email"`
	AvatarURL  string       `json:"avatar_url"`
	ResumeURL  string       `json:"resume_url"`
	Socials    []SocialLink `json:"socials"`
	Skills     []Skill      `json:"skills"`
	Projects   []Project    `json:"projects"`
	Experience []Experience `json:"experience"`
	Education  []Education  `json:"education"`
}

type SocialLink struct {
	Name    string `json:"name"`
	URL     string `json:"url"`
	IconKey string `json:"icon_key"`
}

type Skill struct {
	Name    string `json:"name"`
	LevelEs string `json:"level_es"`
	LevelEn string `json:"level_en"`
	IconURL string `json:"icon_url"`
}

type Project struct {
	TitleEs       string   `json:"title_es"`
	TitleEn       string   `json:"title_en"`
	DescriptionEs string   `json:"description_es"`
	DescriptionEn string   `json:"description_en"`
	Category      string   `json:"category"`
	StatusLabelEs string   `json:"status_label_es"`
	StatusLabelEn string   `json:"status_label_en"`
	Tags          []string `json:"tags"`
	Link          string   `json:"link"`
	RepoLink      string   `json:"repo_link"`
	CoverURL      string   `json:"cover_url"`
}

type Experience struct {
	PeriodEs      string `json:"period_es"`
	PeriodEn      string `json:"period_en"`
	PositionEs    string `json:"position_es"`
	PositionEn    string `json:"position_en"`
	Company       string `json:"company"`
	DescriptionEs string `json:"description_es"`
	DescriptionEn string `json:"description_en"`
}

type Education struct {
	TitleEs       string `json:"title_es"`
	TitleEn       string `json:"title_en"`
	School        string `json:"school"`
	Date          string `json:"date"`
	DescriptionEs string `json:"description_es"`
	DescriptionEn string `json:"description_en"`
}
