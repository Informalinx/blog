package lib

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
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

// See : https://go.dev/blog/context, especially section #package-userip
type nonceCtxKey int

const NonceCtxKeyName nonceCtxKey = 0

type CSPDirective string

const (
	// Fetch directives
	ChildSrc       = "child-src"
	ConnectSrc     = "connect-scr"
	DefaultSrc     = "default-src"
	FencedFrameSrc = "fenced-frame-src"
	FontSrc        = "font-src"
	FrameSrc       = "frame-src"
	ImgSrc         = "img-src"
	ManifestSrc    = "manifest-src"
	MediaSrc       = "media-src"
	ObjectSrc      = "object-src"
	ScriptSrc      = "script-src"
	ScriptSrcElem  = "script-src-elem"
	ScriptSrcAttr  = "script-src-attr"
	StyleSrc       = "style-src"
	StyleSrcElem   = "style-src-elem"
	StyleSrcAttr   = "style-src-attr"
	WorkerSrc      = "worker-src"

	// Document directives
	BaseURI = "base-uri"
	Sandbox = "sandbox"

	// Navigation directives
	FormAction     = "form-action"
	FrameAncestors = "frame-ancestors"

	// Reporting directives
	ReportTo = "report-to"

	// Other directives
	RequireTrustedTypesFor  = "require-trusted-types-for"
	TrustedTypes            = "trusted-types"
	UpgradeInsecureRequests = "upgrade-insecure-requests"
)

func StrictCSPDirectives() map[CSPDirective]string {
	return map[CSPDirective]string{
		DefaultSrc:             "none",
		BaseURI:                "none",
		Sandbox:                "",
		FormAction:             "none",
		FrameAncestors:         "none",
		RequireTrustedTypesFor: "'script'",
		TrustedTypes:           "none",
	}
}

func FormatCSPDirective(name CSPDirective, value string) string {
	return string(name) + " " + value
}

// Should be used for every requests, not only for documents
// but for static files like images, CSS, scripts ...
func CSPMiddleware(directives map[CSPDirective]string, useScriptNonce bool, useStyleNonce bool, reportingEndpoints map[string]url.URL, next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		builder := &strings.Builder{}
		separator := ", "
		for name, value := range directives {
			builder.WriteString(string(name))
			builder.WriteByte(' ')
			builder.WriteString(value)
			builder.WriteString(separator)
		}

		if val, ok := directives[ReportTo]; ok && val != "" {
			for name, url := range reportingEndpoints {
				writer.Header().Add("Reporting-Endpoints", fmt.Sprintf("%s=%q", name, url.String()))
			}
		}

		useNonce := useScriptNonce || useStyleNonce
		if useNonce {
			buffer := [64]byte{}
			if _, err := rand.Read(buffer[:]); err != nil {
				writer.WriteHeader(http.StatusInternalServerError)
				return
			}

			nonce := base64.URLEncoding.EncodeToString(buffer[:])
			nonceDirectives := map[CSPDirective]string{}
			if useScriptNonce {
				nonceDirectives[ScriptSrc] = "'nonce-" + nonce + "'"
			}

			if useStyleNonce {
				nonceDirectives[StyleSrc] = "'nonce-" + nonce + "'"
			}

			for name, value := range nonceDirectives {
				builder.WriteString(string(name))
				builder.WriteByte(' ')
				builder.WriteString(value)
				builder.WriteString(separator)
			}

			// Pass the nonce value in the request's context so another
			// http handler can get it with "request.Context().Value(NonceCtxKeyName)"
			// This value can then be used by a templating engine such as template/html
			// to add a "nonce" attribute with this value to every "script" / "link rel="stylesheet"" elements
			ctx := context.WithValue(request.Context(), NonceCtxKeyName, nonce)
			request = request.WithContext(ctx)
		}

		directivesStr := builder.String()
		directivesStr, _ = strings.CutSuffix(directivesStr, separator)

		writer.Header().Set("Content-Security-Policy", directivesStr)

		next.ServeHTTP(writer, request)
	})
}

type csrfTokenCtxKey int

const CSRFTokenCtxKey csrfTokenCtxKey = 0

type CSRFConfig struct {
	AllowedOrigins []string
	FormFieldName  string
}

type Store interface {
	Get(*http.Request) (string, error)
	Save(string, *http.Request, http.ResponseWriter) error
}

func CSRFMiddleware(conf CSRFConfig, store Store, next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		token, err := GenerateCSRFToken()
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		if err := store.Save(token, request, writer); err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		safeMethods := []string{http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodTrace}
		if !slices.Contains(safeMethods, request.Method) {
			origin := request.Header.Get("Origin")
			referer := request.Referer()
			if origin == "" && referer == "" {
				writer.WriteHeader(http.StatusForbidden)
				return
			}

			if origin == "" || !slices.Contains(conf.AllowedOrigins, origin) {
				writer.WriteHeader(http.StatusForbidden)
				return
			}

			formToken := request.PostFormValue(conf.FormFieldName)
			userToken, err := store.Get(request)
			if err != nil {
				writer.WriteHeader(http.StatusInternalServerError)
				return
			}

			if formToken != userToken {
				writer.WriteHeader(http.StatusForbidden)
				return
			}
		}

		request = request.WithContext(context.WithValue(request.Context(), CSRFTokenCtxKey, token))

		next.ServeHTTP(writer, request)
	})
}

func GenerateCSRFToken() (string, error) {
	buff := [64]byte{}
	if _, err := rand.Read(buff[:]); err != nil {
		return "", err
	}

	key := base64.URLEncoding.EncodeToString(buff[:])

	return key, nil
}
