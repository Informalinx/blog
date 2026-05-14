package blog

import (
	"errors"
	"net/http"

	"github.com/gorilla/sessions"
)

const CSRFTokenKey = "csrfToken"

type CSRFStore struct {
	CookieStore *sessions.CookieStore
}

func (store *CSRFStore) Get(request *http.Request) (string, error) {
	session, err := GetSession(store.CookieStore, request)
	if err != nil {
		return "", err
	}

	if token, ok := session.Values[CSRFTokenKey].(string); ok {
		return token, nil
	}

	return "", errors.New("no CSRF token found inside request cookie")
}

func (store *CSRFStore) Save(token string, request *http.Request, writer http.ResponseWriter) error {
	session, err := GetSession(store.CookieStore, request)
	if err != nil {
		return err
	}

	session.Save(request, writer)
	return nil
}
