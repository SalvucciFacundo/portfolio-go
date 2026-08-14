package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/SalvucciFacundo/portfolio-go/internal/domain"
	"github.com/SalvucciFacundo/portfolio-go/internal/ports"
)

// SkillRepo implements ports.SkillRepo over the skills table.
type SkillRepo struct {
	pool *pgxpool.Pool
}

var _ ports.SkillRepo = (*SkillRepo)(nil)

// NewSkillRepo creates a SkillRepo bound to pool.
func NewSkillRepo(pool *pgxpool.Pool) *SkillRepo {
	return &SkillRepo{pool: pool}
}

const skillColumns = `id, position, name, icon_url, is_tool`

func scanSkill(row pgx.Row) (domain.Skill, error) {
	var s domain.Skill
	err := row.Scan(&s.ID, &s.Position, &s.Name, &s.IconURL, &s.IsTool)
	return s, err
}

// List returns every skill ordered by position.
func (r *SkillRepo) List(ctx context.Context) ([]domain.Skill, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT `+skillColumns+` FROM skills ORDER BY position ASC, id ASC`)
	if err != nil {
		return nil, fmt.Errorf("list skills: %w", err)
	}
	defer rows.Close()

	skills := make([]domain.Skill, 0)
	for rows.Next() {
		s, err := scanSkill(rows)
		if err != nil {
			return nil, fmt.Errorf("scan skill: %w", err)
		}
		skills = append(skills, s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list skills: %w", err)
	}
	return skills, nil
}

// GetByID returns a single skill or pgx.ErrNoRows wrapped when absent.
func (r *SkillRepo) GetByID(ctx context.Context, id int64) (domain.Skill, error) {
	s, err := scanSkill(r.pool.QueryRow(ctx,
		`SELECT `+skillColumns+` FROM skills WHERE id = $1`, id))
	if err != nil {
		return domain.Skill{}, fmt.Errorf("get skill %d: %w", id, err)
	}
	return s, nil
}

// Create inserts a skill and returns its generated id.
func (r *SkillRepo) Create(ctx context.Context, s domain.Skill) (int64, error) {
	var id int64
	err := r.pool.QueryRow(ctx, `
		INSERT INTO skills (position, name, icon_url, is_tool)
		VALUES ($1, $2, $3, $4) RETURNING id`,
		s.Position, s.Name, s.IconURL, s.IsTool,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("create skill %q: %w", s.Name, err)
	}
	return id, nil
}

// Update modifies a skill and reports whether a row was affected.
func (r *SkillRepo) Update(ctx context.Context, s domain.Skill) (bool, error) {
	tag, err := r.pool.Exec(ctx, `
		UPDATE skills SET position = $1, name = $2, icon_url = $3, is_tool = $4
		WHERE id = $5`,
		s.Position, s.Name, s.IconURL, s.IsTool, s.ID,
	)
	if err != nil {
		return false, fmt.Errorf("update skill %d: %w", s.ID, err)
	}
	return tag.RowsAffected() > 0, nil
}

// Delete removes a skill and reports whether a row was affected.
func (r *SkillRepo) Delete(ctx context.Context, id int64) (bool, error) {
	tag, err := r.pool.Exec(ctx, `DELETE FROM skills WHERE id = $1`, id)
	if err != nil {
		return false, fmt.Errorf("delete skill %d: %w", id, err)
	}
	return tag.RowsAffected() > 0, nil
}

// ExistsByName reports whether a skill with the given name exists.
func (r *SkillRepo) ExistsByName(ctx context.Context, name string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM skills WHERE name = $1)`, name,
	).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check skill name %q: %w", name, err)
	}
	return exists, nil
}
