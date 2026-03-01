package auth

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// RepoUser is the user record as stored in the database. The auth package
// works with this type so it can access the password hash for verification.
type RepoUser struct {
	ID           string
	Email        *string
	DisplayName  string
	PasswordHash *string
}

// UserRepository defines the data-access methods required by AuthService.
type UserRepository interface {
	CreateUser(ctx context.Context, displayName, email, passwordHash string) (*RepoUser, error)
	GetUserByEmail(ctx context.Context, email string) (*RepoUser, error)
	GetUserByID(ctx context.Context, id string) (*RepoUser, error)
}

// User is the public user representation returned by auth endpoints.
type User struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	DisplayName string `json:"displayName"`
}

// AuthService handles user signup, login and JWT token management.
type AuthService struct {
	repo      UserRepository
	jwtSecret []byte
}

// NewAuthService creates an AuthService. The JWT signing key is read from
// the JWT_SECRET environment variable, falling back to a development default.
func NewAuthService(repo UserRepository) *AuthService {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "dev-secret-change-me"
	}
	return &AuthService{
		repo:      repo,
		jwtSecret: []byte(secret),
	}
}

const (
	bcryptCost     = 12
	tokenExpiryDur = 7 * 24 * time.Hour // 7 days
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrEmailTaken         = errors.New("a user with this email already exists")
)

// Signup creates a new user account and returns the user with a JWT token.
func (s *AuthService) Signup(ctx context.Context, email, password, displayName string) (*User, string, error) {
	// Check if email is already taken
	existing, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, "", fmt.Errorf("check existing user: %w", err)
	}
	if existing != nil {
		return nil, "", ErrEmailTaken
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return nil, "", fmt.Errorf("hash password: %w", err)
	}

	dbUser, err := s.repo.CreateUser(ctx, displayName, email, string(hash))
	if err != nil {
		return nil, "", fmt.Errorf("create user: %w", err)
	}

	user := repoToUser(dbUser)

	token, err := s.issueToken(user.ID)
	if err != nil {
		return nil, "", err
	}

	return user, token, nil
}

// Login authenticates a user by email/password and returns the user with a JWT.
func (s *AuthService) Login(ctx context.Context, email, password string) (*User, string, error) {
	dbUser, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, "", fmt.Errorf("lookup user: %w", err)
	}
	if dbUser == nil {
		return nil, "", ErrInvalidCredentials
	}
	if dbUser.PasswordHash == nil || *dbUser.PasswordHash == "" {
		return nil, "", ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(*dbUser.PasswordHash), []byte(password)); err != nil {
		return nil, "", ErrInvalidCredentials
	}

	user := repoToUser(dbUser)

	token, err := s.issueToken(user.ID)
	if err != nil {
		return nil, "", err
	}

	return user, token, nil
}

// ValidateToken parses and validates a JWT token string, returning the user ID
// from the "sub" claim.
func (s *AuthService) ValidateToken(tokenStr string) (string, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return s.jwtSecret, nil
	})
	if err != nil {
		return "", fmt.Errorf("invalid token: %w", err)
	}

	sub, err := token.Claims.GetSubject()
	if err != nil || sub == "" {
		return "", errors.New("invalid token: missing subject")
	}

	return sub, nil
}

// GetUserByID retrieves user information by ID (used by the /auth/me endpoint).
func (s *AuthService) GetUserByID(ctx context.Context, id string) (*User, error) {
	dbUser, err := s.repo.GetUserByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	if dbUser == nil {
		return nil, errors.New("user not found")
	}

	return repoToUser(dbUser), nil
}

func (s *AuthService) issueToken(userID string) (string, error) {
	now := time.Now()
	claims := jwt.RegisteredClaims{
		Subject:   userID,
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(tokenExpiryDur)),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}
	return signed, nil
}

func repoToUser(ru *RepoUser) *User {
	email := ""
	if ru.Email != nil {
		email = *ru.Email
	}
	return &User{
		ID:          ru.ID,
		Email:       email,
		DisplayName: ru.DisplayName,
	}
}
