package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/SalvucciFacundo/portfolio-go/internal/domain"
	"github.com/SalvucciFacundo/portfolio-go/internal/ports"
)

// ProfileRepo implements ports.ProfileRepo over the singleton profile table.
type ProfileRepo struct {
	pool *pgxpool.Pool
}

var _ ports.ProfileRepo = (*ProfileRepo)(nil)

// NewProfileRepo creates a ProfileRepo bound to pool.
func NewProfileRepo(pool *pgxpool.Pool) *ProfileRepo {
	return &ProfileRepo{pool: pool}
}

const profileColumns = `id, name, role_es, role_en, headline_es, headline_en,
	summary_es, summary_en, email, avatar_url, resume_url, resume_filename`

func scanProfile(row pgx.Row) (domain.Profile, error) {
	var p domain.Profile
	err := row.Scan(
		&p.ID, &p.Name, &p.RoleEs, &p.RoleEn, &p.HeadlineEs, &p.HeadlineEn,
		&p.SummaryEs, &p.SummaryEn, &p.Email, &p.AvatarURL, &p.ResumeURL, &p.ResumeFilename,
	)
	return p, err
}

// Get returns the base profile row (id=1) without socials or skills.
func (r *ProfileRepo) Get(ctx context.Context) (domain.Profile, error) {
	p, err := scanProfile(r.pool.QueryRow(ctx,
		`SELECT `+profileColumns+` FROM profile WHERE id = 1`))
	if err != nil {
		return domain.Profile{}, fmt.Errorf("get profile: %w", err)
	}
	return p, nil
}

// Update replaces every editable field of the profile row (id=1).
func (r *ProfileRepo) Update(ctx context.Context, p domain.Profile) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE profile SET
			name = $1, role_es = $2, role_en = $3,
			headline_es = $4, headline_en = $5,
			summary_es = $6, summary_en = $7,
			email = $8, avatar_url = $9,
			resume_url = $10, resume_filename = $11,
			updated_at = now()
		WHERE id = 1`,
		p.Name, p.RoleEs, p.RoleEn, p.HeadlineEs, p.HeadlineEn,
		p.SummaryEs, p.SummaryEn, p.Email, p.AvatarURL, p.ResumeURL, p.ResumeFilename,
	)
	if err != nil {
		return fmt.Errorf("update profile: %w", err)
	}
	return nil
}

// SetResume updates only the CV fields of the profile row.
func (r *ProfileRepo) SetResume(ctx context.Context, resumeURL, resumeFilename string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE profile SET resume_url = $1, resume_filename = $2, updated_at = now()
		WHERE id = 1`,
		resumeURL, resumeFilename,
	)
	if err != nil {
		return fmt.Errorf("set resume: %w", err)
	}
	return nil
}

// SetAvatar updates only the avatar of the profile row.
func (r *ProfileRepo) SetAvatar(ctx context.Context, avatarURL string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE profile SET avatar_url = $1, updated_at = now()
		WHERE id = 1`,
		avatarURL,
	)
	if err != nil {
		return fmt.Errorf("set avatar: %w", err)
	}
	return nil
}

// SetSocials replaces every social link in a single transaction.
func (r *ProfileRepo) SetSocials(ctx context.Context, socials []domain.SocialLink) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin socials tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `DELETE FROM social_links`); err != nil {
		return fmt.Errorf("clear socials: %w", err)
	}
	for _, s := range socials {
		if _, err := tx.Exec(ctx, `
			INSERT INTO social_links (position, name, url, icon_key)
			VALUES ($1, $2, $3, $4)`,
			s.Position, s.Name, s.URL, s.IconKey,
		); err != nil {
			return fmt.Errorf("insert social %q: %w", s.Name, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit socials: %w", err)
	}
	return nil
}
