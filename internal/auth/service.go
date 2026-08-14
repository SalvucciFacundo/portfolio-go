package auth

import (
	"context"
	"crypto/subtle"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/SalvucciFacundo/portfolio-go/internal/domain"
	"github.com/SalvucciFacundo/portfolio-go/internal/ports"
)

var (
	// ErrInvalidCredentials is generic on purpose: it never reveals whether the
	// username or the password was wrong.
	ErrInvalidCredentials = errors.New("invalid credentials")
	// ErrNoAdmin is returned when no admin account exists yet.
	ErrNoAdmin = errors.New("no admin user configured")
	// ErrUsernameTaken is returned when the requested username already exists.
	ErrUsernameTaken = errors.New("username already exists")
)

const (
	defaultAdminUsername = "admin"
	passwordMinLen       = 8
	sessionTTL           = 24 * time.Hour
)

// Service implements the admin authentication use cases over an AuthRepo.
type Service struct {
	repo            ports.AuthRepo
	defaultUsername string
}

// NewService creates a Service that authenticates against repo. The default
// admin username is "admin".
func NewService(repo ports.AuthRepo) *Service {
	return &Service{
		repo:            repo,
		defaultUsername: defaultAdminUsername,
	}
}

// CreateAdmin creates the admin account with the given credentials. It fails
// with ErrUsernameTaken when the username is already in use.
func (s *Service) CreateAdmin(ctx context.Context, username, password string) error {
	if len(password) < passwordMinLen {
		return fmt.Errorf("auth: password must be at least %d characters", passwordMinLen)
	}

	if _, err := s.repo.GetAdminByUsername(ctx, username); err == nil {
		return ErrUsernameTaken
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("auth: checking existing admin: %w", err)
	}

	hash, err := HashPassword(password)
	if err != nil {
		return fmt.Errorf("auth: hashing password: %w", err)
	}

	if err := s.repo.CreateAdmin(ctx, username, hash); err != nil {
		return fmt.Errorf("auth: creating admin: %w", err)
	}
	return nil
}

// Login validates password against the single owner admin account (the first
// created) and returns a new session token (kept hashed in the repo) on
// success. The frontend login form only asks for the password, so the username
// is not part of the flow.
func (s *Service) Login(ctx context.Context, password, userAgent string) (string, error) {
	admin, err := s.repo.GetFirstAdmin(ctx)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrNoAdmin
	}
	if err != nil {
		return "", fmt.Errorf("auth: loading admin: %w", err)
	}

	ok, err := VerifyPassword(admin.PasswordHash, password)
	if err != nil {
		return "", fmt.Errorf("auth: verifying password: %w", err)
	}
	if !ok {
		return "", ErrInvalidCredentials
	}

	token, err := randomHex(32)
	if err != nil {
		return "", fmt.Errorf("auth: generating token: %w", err)
	}

	session := domain.Session{
		TokenHash:     sha256hex(token),
		AdminUserID:   admin.ID,
		UserAgentHash: sha256hex(userAgent),
		ExpiresAt:     time.Now().Add(sessionTTL).Unix(),
	}
	if err := s.repo.CreateSession(ctx, session); err != nil {
		return "", fmt.Errorf("auth: creating session: %w", err)
	}
	return token, nil
}

// Logout invalidates the session associated with token.
func (s *Service) Logout(ctx context.Context, token string) error {
	if err := s.repo.DeleteSession(ctx, sha256hex(token)); err != nil {
		return fmt.Errorf("auth: deleting session: %w", err)
	}
	return nil
}

// Authenticate reports whether token belongs to a valid, unexpired session
// whose user-agent fingerprint matches userAgent.
func (s *Service) Authenticate(ctx context.Context, token, userAgent string) bool {
	if token == "" {
		return false
	}

	session, err := s.repo.GetSessionByTokenHash(ctx, sha256hex(token))
	if err != nil {
		return false
	}

	if time.Now().Unix() > session.ExpiresAt {
		_ = s.repo.DeleteSession(ctx, session.TokenHash)
		return false
	}

	if subtle.ConstantTimeCompare([]byte(session.UserAgentHash), []byte(sha256hex(userAgent))) != 1 {
		return false
	}
	return true
}
