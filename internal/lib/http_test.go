package lib

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

const ServerOrigin string = "http://example.com"

var DefaultCORSConfig = CORSConfig{
	AccessControlAllowOrigin:      []string{"http://allowed.com"},
	AccessControlExposeHeaders:    []string{},
	AccessControlMaxAge:           24 * 60 * 60,
	AccessControlAllowMethods:     []string{http.MethodGet},
	AccessControlAllowHeaders:     []string{},
	AccessControlAllowCredentials: false,
}

func TestCORSMiddleware(t *testing.T) {
	t.Run("same origin request", func(t *testing.T) {
		writer := httptest.NewRecorder()

		request := httptest.NewRequest(http.MethodGet, "/", nil)
		request.Header.Set("Origin", ServerOrigin)

		handler := CORSMiddleware(DefaultCORSConfig, ServerOrigin, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

		handler.ServeHTTP(writer, request)

		response := writer.Result()

		corsHeaders := []string{
			"Access-Control-Allow-Origin",
			"Access-Control-Expose-Headers",
			"Access-Control-Max-Age",
			"Access-Control-Allow-Methods",
			"Access-Control-Allow-Headers",
			"Access-Control-Allow-Credentials",
		}

		if err := ExpectMissingHeaders(response, corsHeaders); err != nil {
			t.Error(err)
		}
	})

	t.Run("simple request", func(t *testing.T) {
		writer := httptest.NewRecorder()

		request := httptest.NewRequest(http.MethodGet, "/", nil)
		request.Header.Set("Origin", "http://allowed.com")

		handler := CORSMiddleware(DefaultCORSConfig, ServerOrigin, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

		handler.ServeHTTP(writer, request)

		response := writer.Result()

		if err := ExpectDefaultCORSHeaders(response, DefaultCORSConfig.AccessControlAllowOrigin[0], DefaultCORSConfig.AccessControlAllowCredentials); err != nil {
			t.Error(err)
		}
	})

	t.Run("preflight request no Request-Headers", func(t *testing.T) {
		writer := httptest.NewRecorder()

		request := httptest.NewRequest(http.MethodOptions, "/", nil)
		request.Header.Set("Origin", "http://allowed.com")
		request.Header.Set("Access-Control-Request-Method", http.MethodPost)

		handler := CORSMiddleware(DefaultCORSConfig, ServerOrigin, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

		handler.ServeHTTP(writer, request)

		response := writer.Result()

		if err := ExpectDefaultCORSHeaders(response, DefaultCORSConfig.AccessControlAllowOrigin[0], DefaultCORSConfig.AccessControlAllowCredentials); err != nil {
			t.Log(response.Header)
			t.Error(err)
		}

		if err := ExpectDefaultCORSPreflightHeaders(response, DefaultCORSConfig.AccessControlAllowMethods, DefaultCORSConfig.AccessControlMaxAge); err != nil {
			t.Log(response.Header)
			t.Error(err)
		}

		headers := []string{
			"Access-Control-Allow-Headers",
		}

		if err := ExpectMissingHeaders(response, headers); err != nil {
			t.Log(response.Header)
			t.Error(err)
		}
	})

	t.Run("preflight request with Request-Headers", func(t *testing.T) {
		writer := httptest.NewRecorder()

		request := httptest.NewRequest(http.MethodOptions, "/", nil)
		request.Header.Set("Origin", "http://allowed.com")
		request.Header.Set("Access-Control-Request-Method", http.MethodPost)
		request.Header.Set("Access-Control-Request-Headers", "X-Test-Header")

		handler := CORSMiddleware(DefaultCORSConfig, ServerOrigin, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

		handler.ServeHTTP(writer, request)

		response := writer.Result()

		if err := ExpectDefaultCORSHeaders(response, DefaultCORSConfig.AccessControlAllowOrigin[0], DefaultCORSConfig.AccessControlAllowCredentials); err != nil {
			t.Error(err)
		}

		if err := ExpectDefaultCORSPreflightHeaders(response, DefaultCORSConfig.AccessControlAllowMethods, DefaultCORSConfig.AccessControlMaxAge); err != nil {
			t.Error(err)
		}

		if err := ExpectAllowHeaders(response, DefaultCORSConfig.AccessControlAllowHeaders); err != nil {
			t.Error(err)
		}
	})
}

func ExpectDefaultCORSHeaders(response *http.Response, allowedOrigin string, allowCredentials bool) error {
	if err := ExpectAllowOrigin(response, allowedOrigin); err != nil {
		return err
	}

	if err := ExpectAllowCredentials(response, allowCredentials); err != nil {
		return err
	}

	return nil
}

func ExpectDefaultCORSPreflightHeaders(response *http.Response, allowedMethods []string, maxAge int) error {
	if err := ExpectAllowMethods(response, allowedMethods); err != nil {
		return err
	}

	if err := ExpectMaxAge(response, maxAge); err != nil {
		return err
	}

	return nil
}

func ExpectMissingHeaders(response *http.Response, headers []string) error {
	for _, header := range headers {
		if response.Header.Get(header) != "" {
			return fmt.Errorf("expected no CORS headers in response but got %s = %s", header, response.Header.Get(header))
		}
	}

	return nil
}

func ExpectAllowOrigin(response *http.Response, expected string) error {
	allowOrigin := response.Header.Get("Access-Control-Allow-Origin")
	if allowOrigin != expected {
		return fmt.Errorf("expected %s in Access-Control-Allow-Origin response header but got %s", expected, allowOrigin)
	}

	return nil
}

func ExpectExposeHeaders(response *http.Response, headers []string) error {
	exposeHeaders := response.Header.Get("Access-Control-Expose-Headers")
	expected := strings.Join(headers, ", ")
	if exposeHeaders != expected {
		return fmt.Errorf("expected %s in Access-Control-Expose-Headers response header but got %s", expected, exposeHeaders)
	}

	return nil
}

func ExpectAllowCredentials(response *http.Response, allowed bool) error {
	allowCredentials := response.Header.Get("Access-Control-Allow-Credentials")
	expected := ""
	if allowed {
		expected = "true"
	}

	if allowCredentials != expected {
		return fmt.Errorf("expected %s in Access-Control-Expose-Headers response header but got %s", expected, allowCredentials)
	}

	return nil
}

func ExpectAllowMethods(response *http.Response, methods []string) error {
	allowMethods := response.Header.Get("Access-Control-Allow-Methods")
	expected := strings.Join(methods, ", ")
	if allowMethods != expected {
		return fmt.Errorf("expected %s in Access-Control-Allow-Methods response header but got %s", expected, allowMethods)
	}

	return nil
}

func ExpectAllowHeaders(response *http.Response, headers []string) error {
	allowHeaders := response.Header.Get("Access-Control-Allow-Headers")
	expected := strings.Join(headers, ", ")
	if allowHeaders != expected {
		return fmt.Errorf("expected %s in Access-Control-Allow-Headers response header but got %s", expected, allowHeaders)
	}

	return nil
}

func ExpectMaxAge(response *http.Response, maxAge int) error {
	got := response.Header.Get("Access-Control-Max-Age")
	expected := strconv.Itoa(maxAge)
	if got != expected {
		return fmt.Errorf("expected %s in Access-Control-Max-Age response header but go %s", expected, got)
	}

	return nil
}

func BenchmarkCSPDirectivesConcatenation(b *testing.B) {
	directives := StrictCSPDirectives()
	b.Run("base", func(b *testing.B) {
		for b.Loop() {
			_ = CSPConcatBase(directives, true, false)
		}
	})

	b.Run("builder", func(b *testing.B) {
		for b.Loop() {
			_ = CSPConcatBuilder(directives, true, false)
		}
	})
	b.Run("preallocated", func(b *testing.B) {
		for b.Loop() {
			_ = CSPConcatPreallocated(directives, true, false)
		}
	})
}

func CSPConcatBase(directives map[CSPDirective]string, useScriptNonce bool, useStyleNonce bool) string {
	directivesSlice := []string{}
	for name, value := range directives {
		directivesSlice = append(directivesSlice, FormatCSPDirective(name, value))
	}

	useNonce := useScriptNonce || useStyleNonce
	if useNonce {
		nonce := "example"
		directives = map[CSPDirective]string{}
		if useScriptNonce {
			directives[ScriptSrc] = "nonce-" + nonce
		}

		if useStyleNonce {
			directives[StyleSrc] = "nonce-" + nonce
		}

		for name, value := range directives {
			directivesSlice = append(directivesSlice, FormatCSPDirective(name, value))
		}
	}

	directivesStr := strings.Join(directivesSlice, ", ")

	return directivesStr
}

func CSPConcatBuilder(directives map[CSPDirective]string, useScriptNonce bool, useStyleNonce bool) string {
	builder := &strings.Builder{}
	separator := ", "
	for name, value := range directives {
		builder.WriteString(string(name))
		builder.WriteByte(' ')
		builder.WriteString(value)
		builder.WriteString(separator)
	}

	useNonce := useScriptNonce || useStyleNonce
	if useNonce {
		nonce := "example"
		directives = map[CSPDirective]string{}
		if useScriptNonce {
			directives[ScriptSrc] = "nonce-" + nonce
		}

		if useStyleNonce {
			directives[StyleSrc] = "nonce-" + nonce
		}

		for name, value := range directives {
			builder.WriteString(string(name))
			builder.WriteByte(' ')
			builder.WriteString(value)
			builder.WriteString(separator)
		}
	}

	directivesStr := builder.String()
	directivesStr, _ = strings.CutSuffix(directivesStr, separator)

	return directivesStr
}

func CSPConcatPreallocated(directives map[CSPDirective]string, useScriptNonce bool, useStyleNonce bool) string {
	directivesCount := len(directives)
	if useScriptNonce {
		directivesCount++
	}
	if useScriptNonce {
		directivesCount++
	}

	directivesSlice := make([]string, 0, directivesCount)
	for name, value := range directives {
		directivesSlice = append(directivesSlice, FormatCSPDirective(name, value))
	}

	useNonce := useScriptNonce || useStyleNonce
	if useNonce {
		nonce := "example"
		directives = map[CSPDirective]string{}
		if useScriptNonce {
			directives[ScriptSrc] = "nonce-" + nonce
		}

		if useStyleNonce {
			directives[StyleSrc] = "nonce-" + nonce
		}

		for name, value := range directives {
			directivesSlice = append(directivesSlice, FormatCSPDirective(name, value))
		}
	}

	directivesStr := strings.Join(directivesSlice, ", ")

	return directivesStr
}
