package blog

import (
	"bytes"
	"context"
	"html/template"
	"log/slog"
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/informalinx/blog/internal/env"
	"github.com/informalinx/blog/internal/lib"
	"github.com/informalinx/blog/internal/repository"
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

type HomeHandler struct {
	Config      Config
	Template    *template.Template
	CookieStore *sessions.CookieStore
	Logger      *slog.Logger
}

func (handler *HomeHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	session, err := GetSession(handler.CookieStore, request)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		OnCriticalError(err, handler.Logger)
		return
	}

	data := struct {
		Session *sessions.Session
		Count   int
	}{Count: 10, Session: session}
	buffer := bytes.Buffer{}
	if err := handler.Template.Execute(&buffer, data); err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		OnCriticalError(err, handler.Logger)
		return
	}

	writer.WriteHeader(http.StatusOK)
	writer.Write(buffer.Bytes())
}

type LoginHandler struct {
	Config      Config
	Queries     *repository.Queries
	Template    *template.Template
	CookieStore *sessions.CookieStore
	Logger      *slog.Logger
}

func (handler *LoginHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	session, err := GetSession(handler.CookieStore, request)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		OnCriticalError(err, handler.Logger)
		return
	}

	if request.Method == http.MethodPost {
		email := request.PostFormValue("login_email")
		password := request.PostFormValue("login_password")

		if err := Login(email, password, handler.Config, handler.Queries, session); err != nil {
			switch err.(type) {
			case *AuthenticationError:
				writer.WriteHeader(http.StatusUnauthorized)
			case *ValidationError:
				writer.WriteHeader(http.StatusUnprocessableEntity)
			default:
				writer.WriteHeader(http.StatusInternalServerError)
			}
		}
		return
	}

	buffer := bytes.Buffer{}
	data := struct{ Session *sessions.Session }{Session: session}
	if err := handler.Template.Execute(&buffer, data); err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		OnCriticalError(err, handler.Logger)
		return
	}

	writer.WriteHeader(http.StatusOK)
	writer.Write(buffer.Bytes())
}

type RegisterHandler struct {
	Config      Config
	Queries     *repository.Queries
	Template    *template.Template
	Localizer   *i18n.Localizer
	CookieStore *sessions.CookieStore
	Logger      *slog.Logger
}

func (handler *RegisterHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	session, err := GetSession(handler.CookieStore, request)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		OnCriticalError(err, handler.Logger)
		return
	}

	if request.Method == http.MethodPost {
		user := repository.CreateUserParams{}

		user.Email = request.PostFormValue("register_email")
		user.Username = request.PostFormValue("register_username")
		user.Password = request.PostFormValue("register_password")

		err := Register(user, handler.Config, session, handler.Localizer, handler.Queries)
		if err != nil {
			switch err.(type) {
			case *AuthenticationError:
				writer.WriteHeader(http.StatusUnauthorized)
			case *ValidationError:
				writer.WriteHeader(http.StatusUnprocessableEntity)
			default:
				writer.WriteHeader(http.StatusInternalServerError)
			}
		}
		return
	}

	data := struct{ Session *sessions.Session }{Session: session}
	buffer := bytes.Buffer{}
	if err := handler.Template.Execute(&buffer, data); err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		OnCriticalError(err, handler.Logger)
		return
	}

	writer.WriteHeader(http.StatusOK)
	writer.Write(buffer.Bytes())
}

type RegisterEmailHandler struct {
	Conf    env.Env
	Queries *repository.Queries
	Logger  *slog.Logger
}

func (handler *RegisterEmailHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	email := "user@test.com"
	emailHash, err := lib.HashEmail(email, handler.Conf.EmailHashKey)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		OnCriticalError(err, handler.Logger)
		return
	}

	user, err := handler.Queries.FindByEmail(context.Background(), emailHash)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		OnCriticalError(err, handler.Logger)
		return
	}

	decryptedEmail, err := lib.DecryptEmail(user.Email, handler.Conf.EmailEncryptionKey)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		OnCriticalError(err, handler.Logger)
		return
	}

	writer.WriteHeader(http.StatusOK)
	writer.Write([]byte(decryptedEmail))
}
