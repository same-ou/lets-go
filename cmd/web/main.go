package main

import (
	"crypto/tls"
	"database/sql"
	"flag"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/alexedwards/scs/mysqlstore" // New import
	"github.com/alexedwards/scs/v2"

	// New import
	"github.com/go-playground/form/v4"
	_ "github.com/go-sql-driver/mysql"
	"github.com/same-ou/lets-go/internal/models"
)

type application struct {
	debug          bool
	errorLog       *log.Logger
	infoLog        *log.Logger
	snippets       models.SnippetModelInterface
	users          models.UserModelInterface
	templateCache  map[string]*template.Template
	formDecoder    *form.Decoder
	sessionManager *scs.SessionManager // New field
}

func main() {
	addr := flag.String("addr", ":4000", "HTTP network address")
	dsn := flag.String("dsn", "user:password@tcp(database:3306)/mydb?parseTime=true", "MySQL data source name")
	debug := flag.Bool("debug", false, "Debug mode")

	flag.Parse()

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	db, err := openDB(*dsn)
	if err != nil {
		errorLog.Fatal(err)
	}

	defer db.Close()

	if err = initDb(db); err != nil {
		errorLog.Fatal(err)
	}

	templateCache, err := newTemplateCache()
	if err != nil {
		errorLog.Fatal(err)
	}

	sessionManager := scs.New()
	sessionManager.Store = mysqlstore.New(db)
	sessionManager.Lifetime = 12 * time.Hour
	defer sessionManager.Store.(*mysqlstore.MySQLStore).Close()

	app := &application{
		debug: 		*debug,
		errorLog:       errorLog,
		infoLog:        infoLog,
		snippets:       &models.SnippetModel{DB: db},
		users:          &models.UserModel{DB: db},
		templateCache:  templateCache,
		formDecoder:    form.NewDecoder(),
		sessionManager: sessionManager,
	}

	tlsConfig := &tls.Config{
		CurvePreferences: []tls.CurveID{tls.X25519, tls.CurveP256},
	}
	srv := &http.Server{
		Addr:      *addr,
		ErrorLog:  errorLog,
		Handler:   app.routes(),
		TLSConfig: tlsConfig,

		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	infoLog.Printf("Starting server on %s", *addr)
	err = srv.ListenAndServeTLS("./tls/cert.pem", "./tls/key.pem")
	errorLog.Fatal(err)
}

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
// ReadSQLFile reads the content of an SQL file
func ReadSQLFile(filename string) (string, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// SplitSQLCommands splits the SQL script into individual statements based on semicolons
func SplitSQLCommands(sqlContent string) []string {
	// Split the SQL content by semicolons (ignoring empty lines and spaces)
	sqlCommands := strings.Split(sqlContent, ";")
	var cleanCommands []string

	for _, cmd := range sqlCommands {
		trimmed := strings.TrimSpace(cmd)
		if trimmed != "" {
			cleanCommands = append(cleanCommands, trimmed)
		}
	}
	return cleanCommands
}

func initDb(db *sql.DB) error {
	// Read the SQL file
	sqlContent, err := ReadSQLFile("./schema.sql")
	if err != nil {
		return err
	}

	// Split the SQL content into individual statements
	sqlCommands := SplitSQLCommands(sqlContent)

	// Execute the SQL statements
	for _, cmd := range sqlCommands {
		_, err := db.Exec(cmd)
		if err != nil {
			return err
		}
	}
	return nil
}