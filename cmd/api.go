package main

import (
	"log"
	"net/http"
	"time"

	repo "github.com/JeffreyOmoakah/AUTH.git/internal/adapters/postgresql/sqlc"
	"github.com/JeffreyOmoakah/AUTH.git/internal/users"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/jackc/pgx/v5"
)

func (app *application) mount() http.Handler{
	r := chi.NewRouter()
 // A good base middleware stack
		r.Use(middleware.RequestID) // important for rate limiting
		r.Use(middleware.RealIP)    // import for rate limiting and analytics and tracing
		r.Use(middleware.Logger)
		r.Use(middleware.Recoverer) // recover from crashes
 
		// Set a timeout value on the request context (ctx), that will signal
		// through ctx.Done() that the request has timed out and further
		// processing should be stopped.
		r.Use(middleware.Timeout(60 * time.Second))
		
		// The health endpoint
		r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("all good"))
		})
		
		signupService := users.NewService(repo.New(app.db), app.db)
		signupHandler := users.NewHandler(signupService)
		r.Post("/Signup", signupHandler.Signup)
		return r
}

func (app *application) run(h http.Handler) error{
	srv := &http.Server{
		Addr: app.config.addr,
		Handler: h,
		ReadTimeout: 10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout: time.Minute,
	}
	
	// print a message to the console
	log.Printf("Starting server on %s", app.config.addr)
	
	return srv.ListenAndServe()
}

type application struct {
	config config // Application manager
	db *pgx.Conn
}

type config struct {
	addr string // Address of the API server 
	db dbConfig // Database configuration
}

type dbConfig struct {
	dsn string // Database connection string 
}