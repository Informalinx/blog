package blog

import (
	"bytes"
	"context"
	"html/template"
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/informalinx/blog/internal/blog/repository"
	"github.com/informalinx/blog/internal/env"
	"github.com/informalinx/blog/internal/lib"
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
	Config   env.Env
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
		email := request.PostFormValue("login_email")
		password := request.PostFormValue("login_password")

		session, err := handler.Store.Get(request, "id")
		if err != nil {
			response.StatusCode = http.StatusInternalServerError
			return response, nil
		}

		return Login(email, password, handler.Config, handler.Queries, session, nil)
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
	Config   env.Env
	Queries  *repository.Queries
	Template *template.Template
}

func (handler *RegisterHandler) Handle(request *http.Request) (lib.Response, error) {
	var Guards = []lib.Guard{
		CheckHTTPMethods([]string{http.MethodPost, http.MethodGet}),
	}
	if response, ok := lib.ApplyGuards(Guards, request); ok {
		return response, nil
	}

	if request.Method == http.MethodPost {
		user := repository.CreateUserParams{}

		user.Email = request.PostFormValue("register_email")
		user.Username = request.PostFormValue("register_username")
		user.Password = request.PostFormValue("register_password")

		return Register(user, handler.Config, handler.Queries, nil)
	}

	response := lib.Response{}
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

type RegisterEmailHandler struct {
	Conf    env.Env
	Queries *repository.Queries
}

func (handler *RegisterEmailHandler) Handle(request *http.Request) (lib.Response, error) {
	response := lib.Response{}
	email := "user@test.com"
	emailHash, err := HashEmail(email, handler.Conf.EmailHashKey)
	if err != nil {
		response.StatusCode = http.StatusInternalServerError
		return response, err
	}

	user, err := handler.Queries.FindByEmail(context.Background(), emailHash)
	if err != nil {
		response.StatusCode = http.StatusInternalServerError
		return response, err
	}

	decryptedEmail, err := DecryptEmail(user.Email, handler.Conf.EmailEncryptionKey)
	if err != nil {
		response.StatusCode = http.StatusInternalServerError
		return response, err
	}

	response.Body = []byte(decryptedEmail)
	response.StatusCode = http.StatusOK

	return response, nil
}
