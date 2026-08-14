// Package db implements the PostgreSQL repository adapters backing the
// hexagonal ports in internal/ports. All repositories share one pgx pool.
package db

import (
	"github.com/jackc/pgx/v5/pgxpool"
)

// Store bundles every repository adapter over a single pgx pool.
type Store struct {
	pool *pgxpool.Pool

	Profile    *ProfileRepo
	Skill      *SkillRepo
	Project    *ProjectRepo
	Experience *ExperienceRepo
	Education  *EducationRepo
	Auth       *AuthRepo
}

// New wires all repositories over pool and returns a Store.
func New(pool *pgxpool.Pool) *Store {
	return &Store{
		pool:       pool,
		Profile:    NewProfileRepo(pool),
		Skill:      NewSkillRepo(pool),
		Project:    NewProjectRepo(pool),
		Experience: NewExperienceRepo(pool),
		Education:  NewEducationRepo(pool),
		Auth:       NewAuthRepo(pool),
	}
}
