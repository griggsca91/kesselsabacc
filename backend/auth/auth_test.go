package auth

import (
	"context"
	"errors"
	"sync"
	"testing"
)

// --- In-memory repository for testing ---

type memRepo struct {
	mu    sync.Mutex
	users map[string]*RepoUser // keyed by ID
	seq   int
}

func newMemRepo() *memRepo {
	return &memRepo{users: make(map[string]*RepoUser)}
}

func (m *memRepo) CreateUser(_ context.Context, displayName, email, passwordHash string) (*RepoUser, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check uniqueness of email
	for _, u := range m.users {
		if u.Email != nil && *u.Email == email {
			return nil, errors.New("duplicate email")
		}
	}

	m.seq++
	id := "user-" + string(rune('0'+m.seq))
	e := email
	h := passwordHash
	u := &RepoUser{
		ID:           id,
		Email:        &e,
		DisplayName:  displayName,
		PasswordHash: &h,
	}
	m.users[id] = u
	return u, nil
}

func (m *memRepo) GetUserByEmail(_ context.Context, email string) (*RepoUser, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, u := range m.users {
		if u.Email != nil && *u.Email == email {
			return u, nil
		}
	}
	return nil, nil
}

func (m *memRepo) GetUserByID(_ context.Context, id string) (*RepoUser, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	u, ok := m.users[id]
	if !ok {
		return nil, nil
	}
	return u, nil
}

// --- Tests ---

func TestSignupAndLogin(t *testing.T) {
	repo := newMemRepo()
	svc := NewAuthService(repo)
	ctx := context.Background()

	// Signup
	user, token, err := svc.Signup(ctx, "test@example.com", "password123", "Lando")
	if err != nil {
		t.Fatalf("Signup failed: %v", err)
	}
	if user.Email != "test@example.com" {
		t.Errorf("expected email test@example.com, got %s", user.Email)
	}
	if user.DisplayName != "Lando" {
		t.Errorf("expected displayName Lando, got %s", user.DisplayName)
	}
	if token == "" {
		t.Error("expected non-empty token")
	}

	// Validate the token
	userID, err := svc.ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken failed: %v", err)
	}
	if userID != user.ID {
		t.Errorf("expected userID %s, got %s", user.ID, userID)
	}

	// Login with correct credentials
	loginUser, loginToken, err := svc.Login(ctx, "test@example.com", "password123")
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}
	if loginUser.ID != user.ID {
		t.Errorf("expected same user ID, got %s", loginUser.ID)
	}
	if loginToken == "" {
		t.Error("expected non-empty login token")
	}

	// Login with wrong password
	_, _, err = svc.Login(ctx, "test@example.com", "wrongpassword")
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}

	// Login with non-existent email
	_, _, err = svc.Login(ctx, "nobody@example.com", "password123")
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestSignupDuplicateEmail(t *testing.T) {
	repo := newMemRepo()
	svc := NewAuthService(repo)
	ctx := context.Background()

	_, _, err := svc.Signup(ctx, "han@example.com", "password123", "Han Solo")
	if err != nil {
		t.Fatalf("First signup failed: %v", err)
	}

	_, _, err = svc.Signup(ctx, "han@example.com", "password456", "Han Solo 2")
	if !errors.Is(err, ErrEmailTaken) {
		t.Errorf("expected ErrEmailTaken, got %v", err)
	}
}

func TestValidateToken_Invalid(t *testing.T) {
	repo := newMemRepo()
	svc := NewAuthService(repo)

	_, err := svc.ValidateToken("not-a-valid-token")
	if err == nil {
		t.Error("expected error for invalid token")
	}
}

func TestValidateToken_WrongSecret(t *testing.T) {
	repo := newMemRepo()
	svc1 := NewAuthService(repo)
	ctx := context.Background()

	_, token, err := svc1.Signup(ctx, "chewie@example.com", "password123", "Chewie")
	if err != nil {
		t.Fatalf("Signup failed: %v", err)
	}

	// Create a second service with a different secret
	svc2 := &AuthService{repo: repo, jwtSecret: []byte("different-secret")}
	_, err = svc2.ValidateToken(token)
	if err == nil {
		t.Error("expected error when validating with wrong secret")
	}
}

func TestGetUserByID(t *testing.T) {
	repo := newMemRepo()
	svc := NewAuthService(repo)
	ctx := context.Background()

	created, _, err := svc.Signup(ctx, "leia@example.com", "password123", "Leia")
	if err != nil {
		t.Fatalf("Signup failed: %v", err)
	}

	user, err := svc.GetUserByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetUserByID failed: %v", err)
	}
	if user.DisplayName != "Leia" {
		t.Errorf("expected displayName Leia, got %s", user.DisplayName)
	}

	// Non-existent user
	_, err = svc.GetUserByID(ctx, "non-existent-id")
	if err == nil {
		t.Error("expected error for non-existent user")
	}
}
