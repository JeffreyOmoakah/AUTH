package users

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	repo "github.com/JeffreyOmoakah/AUTH.git/internal/adapters/postgresql/sqlc"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrCredentialsRequired = errors.New("credentials required")
)

type svc struct {
	repo *repo.Queries
	db   *pgx.Conn
	logger *slog.Logger
}

type Service interface {
	Signup(ctx context.Context, tempSignupReq createSignupReq ) (repo.User, error)
	Login(ctx context.Context, req loginReq) (repo.User, error)
	GenerateToken(user repo.User) (string, error)
}

func NewService(repo *repo.Queries, db *pgx.Conn, logger *slog.Logger) Service {
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

func (s *svc) Login(ctx context.Context,req loginReq)(repo.User,error){
	// 1. Get the user row from SQLC
		row, err := s.repo.GetUserByEmail(ctx, req.Email)
		if err != nil {
			return repo.User{}, errors.New("invalid email or password")
		}

		// 2. Compare the password hash
		err = bcrypt.CompareHashAndPassword([]byte(row.Password), []byte(req.Password))
		if err != nil {
			return repo.User{}, errors.New("invalid email or password")
		}

		// 3. MAP the Row to the User struct
		return repo.User{
			ID:       row.ID,
			Email:    row.Email,
			Password: row.Password,
		}, nil
}

func (s *svc) GenerateToken(user repo.User) (string, error) {
    // 1. Create the Claims (the data inside the token)
    claims := jwt.MapClaims{
        "sub":   user.ID,                                // Subject (User ID)
        "exp":   time.Now().Add(time.Hour * 24).Unix(),  // Expiration (24 hours)
        "iat":   time.Now().Unix(),                     // Issued At
        "email": user.Email,                             
    }

    // 2. Create the token using the HMAC SHA256 signing method
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

    // 3. Sign the token with your secret key
    secret := []byte("your-very-secret-key") 
    tokenString, err := token.SignedString(secret)
    if err != nil {
        return "", err
    }

    return tokenString, nil
}