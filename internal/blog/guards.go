package blog

import (
	"net/http"
	"slices"

	"github.com/informalinx/blog/internal/lib"
)

func CheckHTTPMethods(allowedMethods []string) lib.Guard {
	return func(request *http.Request) (lib.Response, bool) {
		response := lib.Response{}
		if !slices.Contains(allowedMethods, request.Method) {
			response.StatusCode = http.StatusMethodNotAllowed
			return response, true
		}

		return response, false
	}
}
