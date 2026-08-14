package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/SalvucciFacundo/portfolio-go/internal/domain"
	"github.com/SalvucciFacundo/portfolio-go/internal/ports"
)

// AuthRepo implements ports.AuthRepo over the admin_users and sessions tables.
type AuthRepo struct {
	pool *pgxpool.Pool
}

var _ ports.AuthRepo = (*AuthRepo)(nil)

// NewAuthRepo creates an AuthRepo bound to pool.
func NewAuthRepo(pool *pgxpool.Pool) *AuthRepo {
	return &AuthRepo{pool: pool}
}

// CreateAdmin inserts a new admin account with an already-hashed password.
func (r *AuthRepo) CreateAdmin(ctx context.Context, username, passwordHash string) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO admin_users (username, password_hash) VALUES ($1, $2)`,
		username, passwordHash,
	)
	if err != nil {
		return fmt.Errorf("create admin %q: %w", username, err)
	}
	return nil
}

// GetAdminCount returns the number of admin accounts.
func (r *AuthRepo) GetAdminCount(ctx context.Context) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM admin_users`).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count admins: %w", err)
	}
	return count, nil
}

// GetAdminByUsername returns the admin account or pgx.ErrNoRows wrapped when absent.
func (r *AuthRepo) GetAdminByUsername(ctx context.Context, username string) (domain.AdminUser, error) {
	var u domain.AdminUser
	err := r.pool.QueryRow(ctx,
		`SELECT id, username, password_hash FROM admin_users WHERE username = $1`, username,
	).Scan(&u.ID, &u.Username, &u.PasswordHash)
	if err != nil {
		return domain.AdminUser{}, fmt.Errorf("get admin %q: %w", username, err)
	}
	return u, nil
}

// GetFirstAdmin returns the first (lowest id) admin account, which is the
// single owner account for this portfolio. Returns pgx.ErrNoRows wrapped when
// no admin exists yet.
func (r *AuthRepo) GetFirstAdmin(ctx context.Context) (domain.AdminUser, error) {
	var u domain.AdminUser
	err := r.pool.QueryRow(ctx,
		`SELECT id, username, password_hash FROM admin_users ORDER BY id ASC LIMIT 1`,
	).Scan(&u.ID, &u.Username, &u.PasswordHash)
	if err != nil {
		return domain.AdminUser{}, fmt.Errorf("get first admin: %w", err)
	}
	return u, nil
}

// CreateSession inserts a session row. ExpiresAt is stored as TIMESTAMPTZ.
func (r *AuthRepo) CreateSession(ctx context.Context, s domain.Session) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO sessions (token_hash, admin_user_id, user_agent_hash, expires_at)
		VALUES ($1, $2, $3, $4)`,
		s.TokenHash, s.AdminUserID, s.UserAgentHash, time.Unix(s.ExpiresAt, 0).UTC(),
	)
	if err != nil {
		return fmt.Errorf("create session: %w", err)
	}
	return nil
}

// GetSessionByTokenHash returns a session joined against admin_users (so the
// account must still exist) or pgx.ErrNoRows wrapped when absent.
func (r *AuthRepo) GetSessionByTokenHash(ctx context.Context, tokenHash string) (domain.Session, error) {
	var (
		s         domain.Session
		expiresAt time.Time
	)
	err := r.pool.QueryRow(ctx, `
		SELECT s.id, s.token_hash, s.admin_user_id, s.user_agent_hash, s.expires_at
		FROM sessions s
		JOIN admin_users a ON a.id = s.admin_user_id
		WHERE s.token_hash = $1`, tokenHash,
	).Scan(&s.ID, &s.TokenHash, &s.AdminUserID, &s.UserAgentHash, &expiresAt)
	if err != nil {
		return domain.Session{}, fmt.Errorf("get session: %w", err)
	}
	s.ExpiresAt = expiresAt.Unix()
	return s, nil
}

// DeleteSession removes a session by its token hash.
func (r *AuthRepo) DeleteSession(ctx context.Context, tokenHash string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM sessions WHERE token_hash = $1`, tokenHash)
	if err != nil {
		return fmt.Errorf("delete session: %w", err)
	}
	return nil
}

// DeleteExpiredSessions removes every session whose expiry is in the past.
func (r *AuthRepo) DeleteExpiredSessions(ctx context.Context) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM sessions WHERE expires_at < now()`)
	if err != nil {
		return fmt.Errorf("delete expired sessions: %w", err)
	}
	return nil
}
