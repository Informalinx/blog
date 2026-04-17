package lib

import "net/http"

// Returns whether the middleware modified the response or not
type Guard func(*http.Request) (Response, bool)

// Returns whether a middleware modified the response or not
func ApplyGuards(guards []Guard, request *http.Request) (Response, bool) {
	for _, guard := range guards {
		if response, ok := guard(request); ok {
			return response, true
		}
	}

	return Response{}, false
}
