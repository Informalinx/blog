package blog

import (
	"net/http"
	"net/url"

	"github.com/informalinx/blog/internal/env"
	"github.com/informalinx/blog/internal/lib"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Session  SessionConfig
	SMTP     SMTPConfig
	UserData UserDataConfig
	CORS     lib.CORSConfig
}

func NewConfig(env env.Env) Config {
	return Config{
		Server: ServerConfig{
			Origin: env.ServerOrigin,
		},
		Database: DatabaseConfig{
			Driver: env.DatabaseDriver,
			DSN:    env.DatabaseDSN,
		},
		Session: SessionConfig{
			AuthKey:     env.SessionKey,
			HttpOnly:    true,
			MaxAge:      60 * 60 * 20,
			Partitioned: true,
			SameSite:    http.SameSiteStrictMode,
			Path:        "/",
			Secure:      true,
		},
		SMTP: SMTPConfig{
			Host:     env.SMTP.Host,
			Port:     env.SMTP.Port,
			Username: env.SMTP.Username,
			Password: env.SMTP.Password,
		},
		UserData: UserDataConfig{
			EmailHashKey:       env.EmailHashKey,
			EmailEncryptionKey: env.EmailEncryptionKey,
		},
		CORS: lib.CORSConfig{
			AccessControlAllowOrigin:   []string{env.ServerOrigin.String()},
			AccessControlExposeHeaders: []string{},
			AccessControlMaxAge:        60,
			AccessControlAllowMethods:  []string{},
			AccessControlAllowHeaders:  []string{},
		},
	}
}

type ServerConfig struct {
	Origin url.URL
}

type DatabaseConfig struct {
	Driver string
	DSN    string
}

type SessionConfig struct {
	AuthKey     string
	HttpOnly    bool
	MaxAge      int
	Partitioned bool
	SameSite    http.SameSite
	Path        string
	Secure      bool
}

type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
}

type UserDataConfig struct {
	EmailHashKey       string
	EmailEncryptionKey []byte
}
