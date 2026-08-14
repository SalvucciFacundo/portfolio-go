// Package ports defines the outgoing (driven) ports of the hexagonal
// architecture: the repository interfaces the application needs.
package ports

import (
	"context"

	"github.com/SalvucciFacundo/portfolio-go/internal/domain"
)

// ProfileRepo persists the singleton profile row (id=1) and its socials.
// Get returns only the base profile fields; socials and skills are composed
// by the caller from their own repositories.
type ProfileRepo interface {
	Get(ctx context.Context) (domain.Profile, error)
	Update(ctx context.Context, profile domain.Profile) error
	SetResume(ctx context.Context, resumeURL, resumeFilename string) error
	SetAvatar(ctx context.Context, avatarURL string) error
	SetSocials(ctx context.Context, socials []domain.SocialLink) error
}

// SkillRepo persists skills. Lists are always ordered by position.
type SkillRepo interface {
	List(ctx context.Context) ([]domain.Skill, error)
	GetByID(ctx context.Context, id int64) (domain.Skill, error)
	Create(ctx context.Context, skill domain.Skill) (int64, error)
	Update(ctx context.Context, skill domain.Skill) (bool, error)
	Delete(ctx context.Context, id int64) (bool, error)
	ExistsByName(ctx context.Context, name string) (bool, error)
}

// ProjectRepo persists projects and their screenshots (project_images).
// List omits screenshots; GetByID loads them ordered by position.
type ProjectRepo interface {
	List(ctx context.Context) ([]domain.Project, error)
	GetByID(ctx context.Context, id int64) (domain.Project, error)
	Create(ctx context.Context, project domain.Project) (int64, error)
	Update(ctx context.Context, project domain.Project) (bool, error)
	Delete(ctx context.Context, id int64) (bool, error)
	AddImages(ctx context.Context, projectID int64, urls []string) error
	DeleteImage(ctx context.Context, imageID int64) (bool, error)
	ExistsByTitleEn(ctx context.Context, titleEn string) (bool, error)
}

// ExperienceRepo persists work experience entries.
type ExperienceRepo interface {
	List(ctx context.Context) ([]domain.Experience, error)
	GetByID(ctx context.Context, id int64) (domain.Experience, error)
	Create(ctx context.Context, experience domain.Experience) (int64, error)
	Update(ctx context.Context, experience domain.Experience) (bool, error)
	Delete(ctx context.Context, id int64) (bool, error)
}

// EducationRepo persists education entries.
type EducationRepo interface {
	List(ctx context.Context) ([]domain.Education, error)
	GetByID(ctx context.Context, id int64) (domain.Education, error)
	Create(ctx context.Context, education domain.Education) (int64, error)
	Update(ctx context.Context, education domain.Education) (bool, error)
	Delete(ctx context.Context, id int64) (bool, error)
}

// AuthRepo persists the admin account and its sessions.
type AuthRepo interface {
	CreateAdmin(ctx context.Context, username, passwordHash string) error
	GetAdminCount(ctx context.Context) (int, error)
	GetAdminByUsername(ctx context.Context, username string) (domain.AdminUser, error)
	CreateSession(ctx context.Context, session domain.Session) error
	GetSessionByTokenHash(ctx context.Context, tokenHash string) (domain.Session, error)
	DeleteSession(ctx context.Context, tokenHash string) error
	DeleteExpiredSessions(ctx context.Context) error
}
