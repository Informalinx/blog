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
	environment, err := env.Load(".env")
	if err != nil {
		log.Fatal(err)
	}

	conf := blog.NewConfig(environment)

	cookieStore := sessions.NewCookieStore([]byte(conf.Session.AuthKey))
	cookieStore.Options.HttpOnly = conf.Session.HttpOnly
	cookieStore.Options.MaxAge = conf.Session.MaxAge
	cookieStore.Options.Partitioned = conf.Session.Partitioned
	cookieStore.Options.SameSite = conf.Session.SameSite
	cookieStore.Options.Path = conf.Session.Path
	cookieStore.Options.Secure = conf.Session.Secure

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

	db, err := sql.Open(conf.Database.Driver, conf.Database.DSN)
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

	homeHandler := blog.HomeHandler{
		Config:      conf,
		Template:    baseTmpl,
		Logger:      logger,
		CookieStore: cookieStore,
	}

	registerHandler := blog.RegisterHandler{
		Config:      conf,
		Template:    registerTmpl,
		Queries:     queries,
		Localizer:   localizer,
		Logger:      logger,
		CookieStore: cookieStore,
	}

	loginHandler := blog.LoginHandler{
		Config:      conf,
		Queries:     queries,
		Template:    loginTmpl,
		Logger:      logger,
		CookieStore: cookieStore,
	}

	mux.Handle("/{$}", lib.CORSMiddleware(conf.CORS, conf.Server.Origin.String(), &homeHandler))
	mux.Handle("/register", lib.CORSMiddleware(conf.CORS, conf.Server.Origin.String(), &registerHandler))
	mux.Handle("/login", lib.CORSMiddleware(conf.CORS, conf.Server.Origin.String(), &loginHandler))

	fmt.Println("Server listening on :", conf.Server.Origin.Host)
	if err := http.ListenAndServe(conf.Server.Origin.Host, mux); err != nil {
		log.Fatal(err)
	}
}
