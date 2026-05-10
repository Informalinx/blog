package blog

import (
	"context"
	"database/sql"
	"errors"

	"github.com/gorilla/sessions"
	"github.com/informalinx/blog/internal/lib"
	"github.com/informalinx/blog/internal/repository"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/crypto/bcrypt"
)

func Register(user repository.CreateUserParams, conf Config, session *sessions.Session, localizer *i18n.Localizer, queries *repository.Queries) error {
	if err := ValidateEmail(user.Email); err != nil {
		return &ValidationError{Message: err.Error()}
	}

	if err := ValidateUsername(user.Username); err != nil {
		return &ValidationError{Message: err.Error()}
	}

	if err := ValidatePassword(user.Password); err != nil {
		return &ValidationError{Message: err.Error()}
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return &CriticalError{Message: err.Error()}
	}

	emailHash, err := lib.HashEmail(user.Email, conf.UserData.EmailHashKey)
	if err != nil {
		return &CriticalError{Message: err.Error()}
	}

	encryptedEmail, err := lib.EncryptEmail(user.Email, conf.UserData.EmailEncryptionKey)
	if err != nil {
		return &CriticalError{Message: err.Error()}
	}

	user.Email = encryptedEmail
	user.EmailHash = emailHash
	user.Password = string(hashed)

	emailExists, err := queries.EmailExists(context.Background(), emailHash)
	if err != nil {
		return &CriticalError{Message: err.Error()}
	}

	if !emailExists {
		if err := queries.CreateUser(context.Background(), user); err != nil {
			return &CriticalError{Message: err.Error()}
		}
	}

	message, err := localizer.Localize(&i18n.LocalizeConfig{
		MessageID: "register_success",
	})

	if err != nil {
		return &CriticalError{Message: err.Error()}
	}

	session.AddFlash(message, "flash_success")

	return nil
}

func Login(email, password string, conf Config, queries *repository.Queries, session *sessions.Session) error {
	if err := ValidateEmail(email); err != nil {
		return &ValidationError{Message: err.Error()}
	}

	if err := ValidatePassword(password); err != nil {
		return &ValidationError{Message: err.Error()}
	}

	userExists := true
	user, err := queries.FindUserByEmail(email, conf.UserData.EmailHashKey)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return &CriticalError{Message: err.Error()}
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
		return &CriticalError{Message: err.Error()}
	}

	if !userExists || errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		return &AuthenticationError{Message: err.Error()}
	}

	session.Values["user_id"] = user.ID

	return nil
}

func IsLoggedIn(userId int, session *sessions.Session) bool {
	if id, ok := session.Values["user_id"]; ok && id == userId {
		return true
	}

	return false
}
