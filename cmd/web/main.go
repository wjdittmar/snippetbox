package main

import (
	"crypto/tls"
	"database/sql"
	"flag"
	"github.com/alexedwards/scs/mysqlstore" // New import
	"github.com/alexedwards/scs/v2"
	"github.com/go-playground/form/v4"
	_ "github.com/go-sql-driver/mysql"
	"github.com/wjdittmar/snippetbox/internal/models"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"time"
)

type application struct {
	users          models.UserModelInterface
	logger         *slog.Logger
	snippets       models.SnippetModelInterface // note use of interface so that we can mock it for tests
	templateCache  map[string]*template.Template
	formDecoder    *form.Decoder
	sessionManager *scs.SessionManager
}

func main() {
	addr := flag.String("addr", ":4000", "HTTP network address")
	dsn := flag.String("dsn", "web:pass@/snippetbox?parseTime=true", "MySQL data source name")
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	db, err := openDB(*dsn)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	defer db.Close()

	templateCache, err := newTemplateCache()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	// Initialize a decoder instance...
	formDecoder := form.NewDecoder()

	sessionManager := scs.New()
	sessionManager.Store = mysqlstore.New(db)
	sessionManager.Lifetime = 12 * time.Hour

	app := &application{
		users:          &models.UserModel{DB: db},
		logger:         logger,
		snippets:       &models.SnippetModel{DB: db},
		templateCache:  templateCache,
		formDecoder:    formDecoder,
		sessionManager: sessionManager,
	}

	// improves performance under heavy loads
	tlsConfig := &tls.Config{
		CurvePreferences: []tls.CurveID{tls.X25519, tls.CurveP256},
	}

	logger.Info("starting server", "addr", *addr)

	srv := &http.Server{
		Addr:      *addr,
		Handler:   app.routes(),
		ErrorLog:  slog.NewLogLogger(logger.Handler(), slog.LevelError),
		TLSConfig: tlsConfig,
	}
	err = srv.ListenAndServeTLS("./tls/cert.pem", "./tls/key.pem")

	logger.Error(err.Error())
	os.Exit(1)
}

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}
