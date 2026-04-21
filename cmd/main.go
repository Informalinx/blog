package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/sessions"
	"github.com/informalinx/blog/internal/blog"
	"github.com/informalinx/blog/internal/env"
	"github.com/informalinx/blog/internal/lib"
	"github.com/informalinx/blog/internal/repository"
	_ "github.com/mattn/go-sqlite3"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v3"
)

// Protect with sync/mutex if modified inside mutliples http handlers or goroutines
var userLocale = language.English.String()

func main() {
	conf, err := env.Load(".env")
	if err != nil {
		log.Fatal(err)
	}

	cookieStore := sessions.NewCookieStore([]byte(conf.SessionKey))

	_ = validator.New()

	bundle := i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("yaml", yaml.Unmarshal)
	msgFile, err := bundle.LoadMessageFile("./translations/messages.en.yaml")
	if err != nil {
		log.Fatal(err)
	}

	if err := bundle.AddMessages(language.English, msgFile.Messages...); err != nil {
		log.Fatal(err)
	}

	db, err := sql.Open(conf.DatabaseDriver, conf.DatabaseDSN)
	if err != nil {
		log.Fatalf("error while connecting to database : %s", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("cannot ping database : %s", err)
	}

	queries := repository.New(db)

	localizer := i18n.NewLocalizer(bundle, userLocale)

	funcMap := template.FuncMap{
		"trans": func(params any, message string) string {
			return localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID:    message,
				TemplateData: params,
			})
		},
		"dict": func(values ...any) map[any]any {
			if len(values)%2 != 0 {
				panic("map : expected an even number of arguments (key/value pairs)")
			}

			result := make(map[any]any, len(values)/2)
			for i, val := range values {
				if i%2 != 0 {
					result[values[i-1]] = val
				}
			}

			return result
		},
	}

	file, err := os.OpenFile("./logs/error.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal(err)
	}

	defer file.Close()

	logger := slog.New(slog.NewJSONHandler(file, &slog.HandlerOptions{}))

	baseTmpl := template.Must(template.New("base.html").Funcs(funcMap).ParseFiles("./website/templates/base.html"))
	registerTmpl := template.Must(template.Must(baseTmpl.Clone()).ParseFiles("./website/templates/register/index.html"))
	loginTmpl := template.Must(template.Must(baseTmpl.Clone()).ParseFiles("./website/templates/login/index.html"))

	mux := http.NewServeMux()

	homeHandler := lib.GlobalHandler{
		Logger: logger,
		HTTPHandler: &blog.HomeHandler{
			Template: baseTmpl,
		},
	}

	registerHandler := lib.GlobalHandler{
		Logger: logger,
		HTTPHandler: &blog.RegisterHandler{
			Config:   conf,
			Template: registerTmpl,
			Queries:  queries,
		},
	}

	loginHandler := lib.GlobalHandler{
		Logger: logger,
		HTTPHandler: &blog.LoginHandler{
			Queries:  queries,
			Template: loginTmpl,
			Store:    cookieStore,
		},
	}

	mux.Handle("/{$}", &homeHandler)
	mux.Handle("/register", &registerHandler)
	mux.Handle("/login", &loginHandler)

	fmt.Println("Server listening on :", conf.ServerAddress)
	if err := http.ListenAndServe(conf.ServerAddress, mux); err != nil {
		log.Fatal(err)
	}
}
