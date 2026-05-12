package lib

import (
	"net/http"
	"slices"
	"strconv"
	"strings"
)

type CORSConfig struct {
	AccessControlAllowOrigin      []string
	AccessControlExposeHeaders    []string
	AccessControlMaxAge           int
	AccessControlAllowMethods     []string
	AccessControlAllowHeaders     []string
	AccessControlAllowCredentials bool
}

// See : https://developer.mozilla.org/en-US/docs/Web/HTTP/Guides/CORS
// See : https://fetch.spec.whatwg.org/#http-cors-protocol
func CORSMiddleware(conf CORSConfig, serverOrigin string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		clientOrigin := request.Header.Get("Origin")

		// See : https://fetch.spec.whatwg.org/#http-requests
		isCORSRequest := clientOrigin != ""
		isCrossOrigin := serverOrigin != "" && clientOrigin != serverOrigin
		if !isCORSRequest || !isCrossOrigin {
			next.ServeHTTP(writer, request)
			return
		}

		// CORS REQUESTS
		allowedOrigins := conf.AccessControlAllowOrigin
		if slices.Contains(allowedOrigins, clientOrigin) {
			writer.Header().Set("Access-Control-Allow-Origin", clientOrigin)
			writer.Header().Set("Vary", "Origin")
		}

		if conf.AccessControlAllowCredentials {
			writer.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		// CORS PREFLIGHT REQUESTS
		isPreflight := request.Method == http.MethodOptions && request.Header.Get("Access-Control-Request-Method") != ""
		if isPreflight {
			writer.Header().Set("Access-Control-Allow-Methods", strings.Join(conf.AccessControlAllowMethods, ", "))

			if request.Header.Get("Access-Control-Request-Headers") != "" {
				writer.Header().Set("Access-Control-Allow-Headers", strings.Join(conf.AccessControlAllowHeaders, ", "))
			}

			writer.Header().Set("Access-Control-Max-Age", strconv.Itoa(conf.AccessControlMaxAge))
			writer.WriteHeader(http.StatusNoContent)
			return
		} else {
			writer.Header().Set("Access-Control-Expose-Headers", strings.Join(conf.AccessControlExposeHeaders, ", "))
		}

		next.ServeHTTP(writer, request)
	})
}
