package main

import (
	"flag"
	"fmt"
	"log"
	"maize/internal/driver"
	"maize/internal/models"
	"net/http"
	"os"
	"time"
)

const version = "1.0.0"

// config is the application configuration structure
type config struct {
	port int
	env  string
	db   struct {
		dsn string
	}
	stripe struct {
		secret string
		key    string
	}
	smtp struct {
		host     string
		port     int
		username string
		password string
	}
}

// application is the application structure
type application struct {
	config   config
	infoLog  *log.Logger
	errorLog *log.Logger
	version  string
	DB       models.DBModel
}

// serve is the application entry point
func (app *application) serve() error {
	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", app.config.port),
		Handler:           app.routes(),
		IdleTimeout:       30 * time.Second,
		ReadTimeout:       10 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      5 * time.Second,
	}

	app.infoLog.Printf("Starting Back end server in %s mode on port %d", app.config.env, app.config.port)

	return srv.ListenAndServe()
}

func main() {
	var cfg config

	mailTrapUser := GoDotEnvVariable("MAILTRAP_USER")
	mailTrapPass := GoDotEnvVariable("MAILTRAP_PASS")

	flag.IntVar(&cfg.port, "port", 4001, "Server port to listen on")
	flag.StringVar(&cfg.env, "env", "development", "Application enviornment {development|production|maintenance}")
	flag.StringVar(&cfg.db.dsn, "dsn", "maize:maize@tcp(localhost:3306)/maize?parseTime=true&tls=false", "DSN")
	flag.StringVar(&cfg.smtp.host, "smtphost", "smtp.mailtrap.io", "smtp host")
	flag.StringVar(&cfg.smtp.username, "smtpuser", mailTrapUser, "smtp user")
	flag.StringVar(&cfg.smtp.password, "smtppass", mailTrapPass, "smtp password")
	flag.IntVar(&cfg.smtp.port, "smtpport", 587, "SMTP port")

	log.Println(mailTrapPass, mailTrapUser, "test")

	flag.Parse()

	cfg.stripe.key = os.Getenv("STRIPE_KEY")
	cfg.stripe.secret = os.Getenv("STRIPE_SECRET")

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	conn, err := driver.OpenDb(cfg.db.dsn)
	if err != nil {
		errorLog.Fatal(err)
	}
	defer conn.Close()

	app := &application{
		config:   cfg,
		infoLog:  infoLog,
		errorLog: errorLog,
		version:  version,
		DB:       models.DBModel{DB: conn},
	}

	err = app.serve()
	if err != nil {
		log.Fatal(err)
	}
}
