package blog

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/informalinx/blog/internal/blog/repository"
	"github.com/informalinx/blog/internal/env"
	"github.com/informalinx/blog/internal/lib"
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

	_, err = FindUserByEmail(user.Email, conf.EmailHashKey, queries)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		response.StatusCode = http.StatusInternalServerError
		return response, err
	}

	emailHash, err := HashEmail(user.Email, conf.EmailHashKey)
	if err != nil {
		response.StatusCode = http.StatusInternalServerError
		return response, err
	}

	encryptedEmail, err := EncryptEmail(user.Email, conf.EmailEncryptionKey)
	if err != nil {
		response.StatusCode = http.StatusInternalServerError
		return response, err
	}

	user.Email = encryptedEmail
	user.EmailHash = emailHash
	user.Password = string(hashed)

	if err := queries.CreateUser(context.Background(), user); err != nil {
		response.StatusCode = http.StatusInternalServerError
		return response, nil
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
	user, err := FindUserByEmail(email, conf.EmailHashKey, queries)
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

	response.Sessions = append(response.Sessions, session)

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

func EncryptEmail(email string, key []byte) (string, error) {
	aesBlock, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(aesBlock)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	encrypted := gcm.Seal(nonce, nonce, []byte(email), nil)

	return base64.URLEncoding.EncodeToString(encrypted), nil
}

func DecryptEmail(email string, key []byte) (string, error) {
	ciphertext, err := base64.URLEncoding.DecodeString(email)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	decrypted, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(decrypted), nil
}

func HashEmail(email string, secret string) (string, error) {
	mac := hmac.New(sha256.New, []byte(secret))
	if _, err := mac.Write([]byte(email)); err != nil {
		return "", err
	}

	return hex.EncodeToString(mac.Sum(nil)), nil
}

func VerifyEmail(email string, hash string, secret string) (bool, error) {
	sig, err := hex.DecodeString(hash)
	if err != nil {
		return false, err
	}

	mac := hmac.New(sha256.New, []byte(secret))
	if _, err := mac.Write([]byte(email)); err != nil {
		return false, err
	}

	return hmac.Equal(sig, mac.Sum(nil)), nil
}

func FindUserByEmail(email, secret string, queries *repository.Queries) (repository.FindByEmailRow, error) {
	hashed, err := HashEmail(email, secret)
	if err != nil {
		return repository.FindByEmailRow{}, err
	}

	return queries.FindByEmail(context.Background(), hashed)
}
