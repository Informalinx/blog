package lib

import (
	"maps"
	"net/http"

	"github.com/gorilla/sessions"
)

type Response struct {
	StatusCode int
	Body       []byte
	Header     http.Header
	Sessions   []*sessions.Session
}

func (response *Response) Redirect(url string, code int) Response {
	response.StatusCode = code
	response.Header.Set("Location", url)

	return *response
}

type HTTPHandler interface {
	Handle(*http.Request) Response
}

type GlobalHandler struct {
	HTTPHandler
}

func (handler *GlobalHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	response := handler.Handle(request)

	for _, session := range response.Sessions {
		if err := session.Save(request, writer); err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	maps.Copy(writer.Header(), response.Header)

	writer.WriteHeader(response.StatusCode)
	writer.Write(response.Body)
}
