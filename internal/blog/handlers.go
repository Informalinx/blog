package blog

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"html/template"
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/informalinx/blog/internal/blog/repository"
	"github.com/informalinx/blog/internal/lib"
	"golang.org/x/crypto/bcrypt"
)

type HomeHandler struct {
	Template *template.Template
}

func (handler *HomeHandler) Handle(request *http.Request) (lib.Response, error) {
	response := lib.Response{}

	data := struct{ Count int }{Count: 10}
	buffer := bytes.Buffer{}
	if err := handler.Template.Execute(&buffer, data); err != nil {
		response.StatusCode = http.StatusInternalServerError
		return response, err
	}

	response.Body = buffer.Bytes()
	response.StatusCode = http.StatusOK

	return response, nil
}

type LoginHandler struct {
	Queries  *repository.Queries
	Template *template.Template
	Store    *sessions.CookieStore
}

func (handler *LoginHandler) Handle(request *http.Request) (lib.Response, error) {
	var Guards = []lib.Guard{
		CheckHTTPMethods([]string{http.MethodGet, http.MethodPost}),
	}
	if response, ok := lib.ApplyGuards(Guards, request); ok {
		return response, nil
	}

	response := lib.Response{}
	if request.Method == http.MethodPost {
		username := request.PostFormValue("login_username")
		password := request.PostFormValue("login_password")

		if err := ValidateUsername(username); err != nil {
			response.StatusCode = http.StatusUnprocessableEntity
			return response, nil
		}

		if err := ValidatePassword(password); err != nil {
			response.StatusCode = http.StatusUnprocessableEntity
			return response, nil
		}

		userExists := true
		user, err := handler.Queries.FindByUsername(context.Background(), username)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			response.StatusCode = http.StatusInternalServerError
			return response, err
		}

		// Dummy hash used for mitigating timing attacks. See below
		hashed := user.Password
		if errors.Is(err, sql.ErrNoRows) {
			userExists = false
			hashed = "$2a$12$1BDi.pNVTLYDZyv4o1Q62ubt39vMrypgPuOVVbO6HfiOkHWrWu7XC"
		}

		// Hash password before checking if the user exists to avoid timing attacks
		err = bcrypt.CompareHashAndPassword([]byte(hashed), []byte(password))
		if err != nil && !errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			response.StatusCode = http.StatusInternalServerError
			return response, err
		}

		if !userExists || errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			response.StatusCode = http.StatusUnauthorized
			return response, nil
		}

		session, err := handler.Store.Get(request, "id")
		if err != nil {
			response.StatusCode = http.StatusInternalServerError
			return response, err
		}

		session.Values["user_id"] = user.ID

		response.Sessions = append(response.Sessions, session)
	}

	buffer := bytes.Buffer{}
	data := map[string]string{}
	if err := handler.Template.Execute(&buffer, data); err != nil {
		response.StatusCode = http.StatusInternalServerError
		return response, err
	}

	response.StatusCode = http.StatusOK
	response.Body = buffer.Bytes()

	return response, nil
}

type RegisterHandler struct {
	Queries  *repository.Queries
	Template *template.Template
}

func (handler *RegisterHandler) Handle(request *http.Request) (lib.Response, error) {
	var Guards = []lib.Guard{
		CheckHTTPMethods([]string{http.MethodPost, http.MethodGet}),
	}
	response := lib.Response{}
	if response, ok := lib.ApplyGuards(Guards, request); ok {
		return response, nil
	}

	if request.Method == http.MethodPost {
		username := request.PostFormValue("register_username")
		password := request.PostFormValue("register_password")

		if err := ValidateUsername(username); err != nil {
			response.StatusCode = http.StatusUnprocessableEntity
			return response, nil
		}

		if err := ValidatePassword(password); err != nil {
			response.StatusCode = http.StatusUnprocessableEntity
			return response, nil
		}

		hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			response.StatusCode = http.StatusInternalServerError
			return response, err
		}

		_, err = handler.Queries.FindByUsername(context.Background(), username)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			response.StatusCode = http.StatusInternalServerError
			return response, err
		}

		if err == nil {
			response.StatusCode = http.StatusUnauthorized
			return response, nil
		}

		createUserParams := repository.CreateUserParams{
			Username: username,
			Password: string(hashed),
		}

		if err := handler.Queries.CreateUser(context.Background(), createUserParams); err != nil {
			response.StatusCode = http.StatusInternalServerError
			return response, nil
		}

		return response.Redirect("/", http.StatusSeeOther), nil
	}

	data := struct{}{}
	buffer := bytes.Buffer{}
	if err := handler.Template.Execute(&buffer, data); err != nil {
		response.StatusCode = http.StatusInternalServerError
		return response, err
	}

	response.Body = buffer.Bytes()
	response.StatusCode = http.StatusOK

	return response, nil
}
