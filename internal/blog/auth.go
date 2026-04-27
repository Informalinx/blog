package blog

import (
	"context"
	"database/sql"
	"errors"
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/informalinx/blog/internal/env"
	"github.com/informalinx/blog/internal/lib"
	"github.com/informalinx/blog/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

func Register(user repository.CreateUserParams, conf env.Env, queries *repository.Queries, onSuccess func(*lib.Response) error) (lib.Response, error) {
	response := lib.Response{}
	if err := ValidateEmail(user.Email); err != nil {
		response.StatusCode = http.StatusUnprocessableEntity
		return response, nil
	}

	if err := ValidateUsername(user.Username); err != nil {
		response.StatusCode = http.StatusUnprocessableEntity
		return response, nil
	}

	if err := ValidatePassword(user.Password); err != nil {
		response.StatusCode = http.StatusUnprocessableEntity
		return response, nil
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		response.StatusCode = http.StatusInternalServerError
		return response, err
	}

	emailHash, err := lib.HashEmail(user.Email, conf.EmailHashKey)
	if err != nil {
		response.StatusCode = http.StatusInternalServerError
		return response, err
	}

	encryptedEmail, err := lib.EncryptEmail(user.Email, conf.EmailEncryptionKey)
	if err != nil {
		response.StatusCode = http.StatusInternalServerError
		return response, err
	}

	user.Email = encryptedEmail
	user.EmailHash = emailHash
	user.Password = string(hashed)

	emailExists, err := queries.EmailExists(context.Background(), emailHash)
	if err != nil {
		response.StatusCode = http.StatusInternalServerError
		return response, err
	}

	if !emailExists {
		if err := queries.CreateUser(context.Background(), user); err != nil {
			response.StatusCode = http.StatusInternalServerError
			return response, err
		}
	}

	if onSuccess == nil {
		return response.Redirect("/", http.StatusSeeOther), nil
	} else {
		err := onSuccess(&response)
		return response, err
	}
}

func Login(email, password string, conf env.Env, queries *repository.Queries, session *sessions.Session, onSuccess func(*lib.Response) error) (lib.Response, error) {
	response := lib.Response{}
	if err := ValidateEmail(email); err != nil {
		response.StatusCode = http.StatusUnprocessableEntity
		return response, nil
	}

	if err := ValidatePassword(password); err != nil {
		response.StatusCode = http.StatusUnprocessableEntity
		return response, nil
	}

	userExists := true
	user, err := queries.FindUserByEmail(email, conf.EmailHashKey)
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

	session.Values["user_id"] = user.ID

	if onSuccess == nil {
		return response.Redirect("/", http.StatusSeeOther), nil
	} else {
		err := onSuccess(&response)
		return response, err
	}
}

func IsLoggedIn(userId int, session *sessions.Session) bool {
	if id, ok := session.Values["user_id"]; ok && id == userId {
		return true
	}

	return false
}
