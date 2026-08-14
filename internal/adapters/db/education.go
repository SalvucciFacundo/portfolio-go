package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/SalvucciFacundo/portfolio-go/internal/domain"
	"github.com/SalvucciFacundo/portfolio-go/internal/ports"
)

// EducationRepo implements ports.EducationRepo over the education table.
type EducationRepo struct {
	pool *pgxpool.Pool
}

var _ ports.EducationRepo = (*EducationRepo)(nil)

// NewEducationRepo creates an EducationRepo bound to pool.
func NewEducationRepo(pool *pgxpool.Pool) *EducationRepo {
	return &EducationRepo{pool: pool}
}

const educationColumns = `id, position, title_es, title_en, school, date,
	is_course, description_es, description_en`

func scanEducation(row pgx.Row) (domain.Education, error) {
	var e domain.Education
	err := row.Scan(
		&e.ID, &e.Position, &e.TitleEs, &e.TitleEn, &e.School, &e.Date,
		&e.IsCourse, &e.DescriptionEs, &e.DescriptionEn,
	)
	return e, err
}

// List returns every education entry ordered by position.
func (r *EducationRepo) List(ctx context.Context) ([]domain.Education, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT `+educationColumns+` FROM education ORDER BY position ASC, id ASC`)
	if err != nil {
		return nil, fmt.Errorf("list education: %w", err)
	}
	defer rows.Close()

	items := make([]domain.Education, 0)
	for rows.Next() {
		e, err := scanEducation(rows)
		if err != nil {
			return nil, fmt.Errorf("scan education: %w", err)
		}
		items = append(items, e)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list education: %w", err)
	}
	return items, nil
}

// GetByID returns a single education entry or pgx.ErrNoRows wrapped when absent.
func (r *EducationRepo) GetByID(ctx context.Context, id int64) (domain.Education, error) {
	e, err := scanEducation(r.pool.QueryRow(ctx,
		`SELECT `+educationColumns+` FROM education WHERE id = $1`, id))
	if err != nil {
		return domain.Education{}, fmt.Errorf("get education %d: %w", id, err)
	}
	return e, nil
}

// Create inserts an education entry and returns its generated id.
func (r *EducationRepo) Create(ctx context.Context, e domain.Education) (int64, error) {
	var id int64
	err := r.pool.QueryRow(ctx, `
		INSERT INTO education (
			position, title_es, title_en, school, date,
			is_course, description_es, description_en
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`,
		e.Position, e.TitleEs, e.TitleEn, e.School, e.Date,
		e.IsCourse, e.DescriptionEs, e.DescriptionEn,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("create education %q: %w", e.TitleEn, err)
	}
	return id, nil
}

// Update modifies an education entry and reports whether a row was affected.
func (r *EducationRepo) Update(ctx context.Context, e domain.Education) (bool, error) {
	tag, err := r.pool.Exec(ctx, `
		UPDATE education SET
			position = $1, title_es = $2, title_en = $3, school = $4, date = $5,
			is_course = $6, description_es = $7, description_en = $8
		WHERE id = $9`,
		e.Position, e.TitleEs, e.TitleEn, e.School, e.Date,
		e.IsCourse, e.DescriptionEs, e.DescriptionEn, e.ID,
	)
	if err != nil {
		return false, fmt.Errorf("update education %d: %w", e.ID, err)
	}
	return tag.RowsAffected() > 0, nil
}

// Delete removes an education entry and reports whether a row was affected.
func (r *EducationRepo) Delete(ctx context.Context, id int64) (bool, error) {
	tag, err := r.pool.Exec(ctx, `DELETE FROM education WHERE id = $1`, id)
	if err != nil {
		return false, fmt.Errorf("delete education %d: %w", id, err)
	}
	return tag.RowsAffected() > 0, nil
}
