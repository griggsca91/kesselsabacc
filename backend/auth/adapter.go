package auth

import (
	"context"
	"sabacc/db"
)

// DBAdapter wraps a db.Repository to satisfy the auth.UserRepository interface.
type DBAdapter struct {
	repo *db.Repository
}

// NewDBAdapter creates an adapter from a db.Repository.
func NewDBAdapter(repo *db.Repository) *DBAdapter {
	return &DBAdapter{repo: repo}
}

func (a *DBAdapter) CreateUser(ctx context.Context, displayName, email, passwordHash string) (*RepoUser, error) {
	u, err := a.repo.CreateUser(ctx, displayName, email, passwordHash)
	if err != nil {
		return nil, err
	}
	return dbUserToRepoUser(u), nil
}

func (a *DBAdapter) GetUserByEmail(ctx context.Context, email string) (*RepoUser, error) {
	u, err := a.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, nil
	}
	return dbUserToRepoUser(u), nil
}

func (a *DBAdapter) GetUserByID(ctx context.Context, id string) (*RepoUser, error) {
	u, err := a.repo.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, nil
	}
	return dbUserToRepoUser(u), nil
}

func dbUserToRepoUser(u *db.User) *RepoUser {
	return &RepoUser{
		ID:           u.ID,
		Email:        u.Email,
		DisplayName:  u.DisplayName,
		PasswordHash: u.PasswordHash,
	}
}
