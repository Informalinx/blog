package blog

import (
	"bytes"
	"html/template"
	"net/http"

	"github.com/informalinx/blog/internal/lib"
)

type HomeHandler struct {
	Template *template.Template
}

func (handler *HomeHandler) Handle(request *http.Request) lib.Response {
	response := lib.Response{}

	data := struct{ Count int }{Count: 10}
	buffer := bytes.Buffer{}
	if err := handler.Template.Execute(&buffer, data); err != nil {
		response.StatusCode = http.StatusInternalServerError
		return response
	}

	response.Body = buffer.Bytes()
	response.StatusCode = http.StatusOK

	return response
}
