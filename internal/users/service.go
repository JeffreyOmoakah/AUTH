package users

import (
	"context"
	"errors"
	"fmt"

	repo "github.com/JeffreyOmoakah/AUTH.git/internal/adapters/postgresql/sqlc"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrCredentialsRequired = errors.New("credentials required")
)

type svc struct {
	repo *repo.Queries
	db   *pgx.Conn
}

type Service interface {
	Signup(ctx context.Context, tempSignupReq createSignupReq ) (repo.User, error)
}

func NewService(repo *repo.Queries, db *pgx.Conn) Service {
	return &svc{
		repo: repo,
		db:   db,
	}
}

func (s *svc) Signup(ctx context.Context, tempSignupReq createSignupReq) (repo.User, error) {
	// VALIDATE THE PAYLOAD 
	if tempSignupReq.Email == "" || tempSignupReq.Password == "" {
        return repo.User{}, ErrCredentialsRequired
    }
	// Hash Password 
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(tempSignupReq.Password), bcrypt.DefaultCost)
    if err != nil {
        return repo.User{}, fmt.Errorf("failed to hash password: %w", err)
    }
	
	tx, err := s.db.Begin(ctx)
	if err != nil { 
		return repo.User{}, err
	}
	defer tx.Rollback(ctx)

	qtx := s.repo.WithTx(tx)
	
	
	// LOOK IF USER ALREADY EXISTS 
	row, err := qtx.CreateUser(ctx, repo.CreateUserParams{
        Email:    tempSignupReq.Email,
        Password: string(hashedPassword),
    })
	
	if err != nil {
        return repo.User{}, err 
    }
    
    newUser := repo.User{
            ID:        row.ID,
            Email:     row.Email,
            Password:  string(hashedPassword),
            CreatedAt: row.CreatedAt,
        }
    // Commit
    if err := tx.Commit(ctx); err != nil {
            return repo.User{}, err
        }
    
    return newUser, nil
}

