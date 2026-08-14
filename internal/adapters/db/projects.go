package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/SalvucciFacundo/portfolio-go/internal/domain"
	"github.com/SalvucciFacundo/portfolio-go/internal/ports"
)

// ProjectRepo implements ports.ProjectRepo over the projects table and its
// project_images rows (screenshots).
type ProjectRepo struct {
	pool *pgxpool.Pool
}

var _ ports.ProjectRepo = (*ProjectRepo)(nil)

// NewProjectRepo creates a ProjectRepo bound to pool.
func NewProjectRepo(pool *pgxpool.Pool) *ProjectRepo {
	return &ProjectRepo{pool: pool}
}

// projectColumns excludes screenshots (project_images).
const projectColumns = `id, position, title_es, title_en, description_es, description_en,
	tech_description_es, tech_description_en, category, tags, link, repo_link, cover_url`

func scanProject(row pgx.Row) (domain.Project, error) {
	var p domain.Project
	err := row.Scan(
		&p.ID, &p.Position, &p.TitleEs, &p.TitleEn, &p.DescriptionEs, &p.DescriptionEn,
		&p.TechDescriptionEs, &p.TechDescriptionEn, &p.Category, &p.Tags,
		&p.Link, &p.RepoLink, &p.CoverURL,
	)
	return p, err
}

// List returns every project ordered by position, without screenshots.
func (r *ProjectRepo) List(ctx context.Context) ([]domain.Project, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT `+projectColumns+` FROM projects ORDER BY position ASC, id ASC`)
	if err != nil {
		return nil, fmt.Errorf("list projects: %w", err)
	}
	defer rows.Close()

	projects := make([]domain.Project, 0)
	for rows.Next() {
		p, err := scanProject(rows)
		if err != nil {
			return nil, fmt.Errorf("scan project: %w", err)
		}
		projects = append(projects, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list projects: %w", err)
	}
	return projects, nil
}

// GetByID returns a project with its screenshots ordered by position,
// or pgx.ErrNoRows wrapped when absent.
func (r *ProjectRepo) GetByID(ctx context.Context, id int64) (domain.Project, error) {
	p, err := scanProject(r.pool.QueryRow(ctx,
		`SELECT `+projectColumns+` FROM projects WHERE id = $1`, id))
	if err != nil {
		return domain.Project{}, fmt.Errorf("get project %d: %w", id, err)
	}

	rows, err := r.pool.Query(ctx, `
		SELECT url FROM project_images
		WHERE project_id = $1
		ORDER BY position ASC, id ASC`, id)
	if err != nil {
		return domain.Project{}, fmt.Errorf("list images for project %d: %w", id, err)
	}
	defer rows.Close()

	p.Screenshots = make([]string, 0)
	for rows.Next() {
		var url string
		if err := rows.Scan(&url); err != nil {
			return domain.Project{}, fmt.Errorf("scan image for project %d: %w", id, err)
		}
		p.Screenshots = append(p.Screenshots, url)
	}
	if err := rows.Err(); err != nil {
		return domain.Project{}, fmt.Errorf("list images for project %d: %w", id, err)
	}
	return p, nil
}

// Create inserts a project (and its screenshots when present) and returns
// the generated project id. Project and images share one transaction.
func (r *ProjectRepo) Create(ctx context.Context, p domain.Project) (int64, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("begin create project tx: %w", err)
	}
	defer tx.Rollback(ctx)

	var id int64
	err = tx.QueryRow(ctx, `
		INSERT INTO projects (
			position, title_es, title_en, description_es, description_en,
			tech_description_es, tech_description_en, category, tags,
			link, repo_link, cover_url
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id`,
		p.Position, p.TitleEs, p.TitleEn, p.DescriptionEs, p.DescriptionEn,
		p.TechDescriptionEs, p.TechDescriptionEn, p.Category, p.Tags,
		p.Link, p.RepoLink, p.CoverURL,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("create project %q: %w", p.TitleEn, err)
	}

	if err := insertProjectImages(ctx, tx, id, p.Screenshots); err != nil {
		return 0, err
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("commit create project: %w", err)
	}
	return id, nil
}

// insertProjectImages appends screenshots with the given slice order.
func insertProjectImages(ctx context.Context, tx pgx.Tx, projectID int64, urls []string) error {
	for i, url := range urls {
		if _, err := tx.Exec(ctx, `
			INSERT INTO project_images (project_id, position, url)
			VALUES ($1, $2, $3)`,
			projectID, i, url,
		); err != nil {
			return fmt.Errorf("insert image for project %d: %w", projectID, err)
		}
	}
	return nil
}

// Update modifies a project (screenshots are managed via AddImages and
// DeleteImage) and reports whether a row was affected.
func (r *ProjectRepo) Update(ctx context.Context, p domain.Project) (bool, error) {
	tag, err := r.pool.Exec(ctx, `
		UPDATE projects SET
			position = $1, title_es = $2, title_en = $3,
			description_es = $4, description_en = $5,
			tech_description_es = $6, tech_description_en = $7,
			category = $8, tags = $9, link = $10, repo_link = $11, cover_url = $12,
			updated_at = now()
		WHERE id = $13`,
		p.Position, p.TitleEs, p.TitleEn, p.DescriptionEs, p.DescriptionEn,
		p.TechDescriptionEs, p.TechDescriptionEn, p.Category, p.Tags,
		p.Link, p.RepoLink, p.CoverURL, p.ID,
	)
	if err != nil {
		return false, fmt.Errorf("update project %d: %w", p.ID, err)
	}
	return tag.RowsAffected() > 0, nil
}

// Delete removes a project; its images cascade through the FK constraint.
func (r *ProjectRepo) Delete(ctx context.Context, id int64) (bool, error) {
	tag, err := r.pool.Exec(ctx, `DELETE FROM projects WHERE id = $1`, id)
	if err != nil {
		return false, fmt.Errorf("delete project %d: %w", id, err)
	}
	return tag.RowsAffected() > 0, nil
}

// AddImages appends screenshots to a project in a single transaction. Each
// new image takes the next position after the current maximum.
func (r *ProjectRepo) AddImages(ctx context.Context, projectID int64, urls []string) error {
	if len(urls) == 0 {
		return nil
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin add images tx: %w", err)
	}
	defer tx.Rollback(ctx)

	for _, url := range urls {
		if _, err := tx.Exec(ctx, `
			INSERT INTO project_images (project_id, position, url)
			SELECT $1, COALESCE(MAX(position), -1) + 1, $2
			FROM project_images WHERE project_id = $1`, projectID, url,
		); err != nil {
			return fmt.Errorf("add image %q to project %d: %w", url, projectID, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit add images: %w", err)
	}
	return nil
}

// DeleteImage removes a screenshot and reports whether a row was affected.
func (r *ProjectRepo) DeleteImage(ctx context.Context, imageID int64) (bool, error) {
	tag, err := r.pool.Exec(ctx, `DELETE FROM project_images WHERE id = $1`, imageID)
	if err != nil {
		return false, fmt.Errorf("delete image %d: %w", imageID, err)
	}
	return tag.RowsAffected() > 0, nil
}

// ExistsByTitleEn reports whether a project with the given English title exists.
func (r *ProjectRepo) ExistsByTitleEn(ctx context.Context, titleEn string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM projects WHERE title_en = $1)`, titleEn,
	).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check project title %q: %w", titleEn, err)
	}
	return exists, nil
}
