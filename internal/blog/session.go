package blog

import (
	"net/http"

	"github.com/gorilla/sessions"
)

func GetSession(store *sessions.CookieStore, request *http.Request) (*sessions.Session, error) {
	session, err := store.Get(request, "id")
	if err != nil {
		return &sessions.Session{}, err
	}

	return session, nil
}
