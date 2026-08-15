package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/SalvucciFacundo/portfolio-go/internal/domain"
	"github.com/SalvucciFacundo/portfolio-go/internal/ports"
)

// ExperienceRepo implements ports.ExperienceRepo over the experience table.
type ExperienceRepo struct {
	pool *pgxpool.Pool
}

var _ ports.ExperienceRepo = (*ExperienceRepo)(nil)

// NewExperienceRepo creates an ExperienceRepo bound to pool.
func NewExperienceRepo(pool *pgxpool.Pool) *ExperienceRepo {
	return &ExperienceRepo{pool: pool}
}

const experienceColumns = `id, position, company, position_es, position_en,
	period_es, period_en, description_es, description_en`

func scanExperience(row pgx.Row) (domain.Experience, error) {
	var e domain.Experience
	err := row.Scan(
		&e.ID, &e.Position, &e.Company, &e.PositionEs, &e.PositionEn,
		&e.PeriodEs, &e.PeriodEn, &e.DescriptionEs, &e.DescriptionEn,
	)
	return e, err
}

// List returns every experience entry ordered newest-first by the start year
// of its period (e.g. "2026 — Presente" sorts before "2023 — 2025"),
// independent of the manual position field. Within the same start year, active
// roles (periods ending in "Presente"/"Present") come before finished ones.
// Position/id act as final tie-breakers.
func (r *ExperienceRepo) List(ctx context.Context) ([]domain.Experience, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT `+experienceColumns+` FROM experience
		 ORDER BY NULLIF(regexp_replace(period_es, '[^0-9].*$', ''), '')::int DESC NULLS LAST,
		          CASE WHEN period_es ILIKE '%presente%' THEN 0 ELSE 1 END ASC,
		          position ASC, id ASC`)
	if err != nil {
		return nil, fmt.Errorf("list experience: %w", err)
	}
	defer rows.Close()

	items := make([]domain.Experience, 0)
	for rows.Next() {
		e, err := scanExperience(rows)
		if err != nil {
			return nil, fmt.Errorf("scan experience: %w", err)
		}
		items = append(items, e)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list experience: %w", err)
	}
	return items, nil
}

// GetByID returns a single experience entry or pgx.ErrNoRows wrapped when absent.
func (r *ExperienceRepo) GetByID(ctx context.Context, id int64) (domain.Experience, error) {
	e, err := scanExperience(r.pool.QueryRow(ctx,
		`SELECT `+experienceColumns+` FROM experience WHERE id = $1`, id))
	if err != nil {
		return domain.Experience{}, fmt.Errorf("get experience %d: %w", id, err)
	}
	return e, nil
}

// Create inserts an experience entry and returns its generated id.
func (r *ExperienceRepo) Create(ctx context.Context, e domain.Experience) (int64, error) {
	var id int64
	err := r.pool.QueryRow(ctx, `
		INSERT INTO experience (
			position, company, position_es, position_en,
			period_es, period_en, description_es, description_en
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`,
		e.Position, e.Company, e.PositionEs, e.PositionEn,
		e.PeriodEs, e.PeriodEn, e.DescriptionEs, e.DescriptionEn,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("create experience at %q: %w", e.Company, err)
	}
	return id, nil
}

// Update modifies an experience entry and reports whether a row was affected.
func (r *ExperienceRepo) Update(ctx context.Context, e domain.Experience) (bool, error) {
	tag, err := r.pool.Exec(ctx, `
		UPDATE experience SET
			position = $1, company = $2, position_es = $3, position_en = $4,
			period_es = $5, period_en = $6, description_es = $7, description_en = $8
		WHERE id = $9`,
		e.Position, e.Company, e.PositionEs, e.PositionEn,
		e.PeriodEs, e.PeriodEn, e.DescriptionEs, e.DescriptionEn, e.ID,
	)
	if err != nil {
		return false, fmt.Errorf("update experience %d: %w", e.ID, err)
	}
	return tag.RowsAffected() > 0, nil
}

// Delete removes an experience entry and reports whether a row was affected.
func (r *ExperienceRepo) Delete(ctx context.Context, id int64) (bool, error) {
	tag, err := r.pool.Exec(ctx, `DELETE FROM experience WHERE id = $1`, id)
	if err != nil {
		return false, fmt.Errorf("delete experience %d: %w", id, err)
	}
	return tag.RowsAffected() > 0, nil
}
