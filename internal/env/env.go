package env

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Env struct {
	ServerAddress  string
	DatabaseDriver string
	DatabaseDSN    string
	SessionKey     string
	SMTP           SMTPConfig
}

type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
}

type Loader func(*Env, string) error

func Load(filenames ...string) (Env, error) {
	env := Env{}
	if err := godotenv.Load(filenames...); err != nil {
		return env, fmt.Errorf("error while loading .env file : %s", err)
	}

	loaders := map[string]Loader{
		"SERVER_ADDRESS":  StringLoader(&env.ServerAddress),
		"DATABASE_DRIVER": StringLoader(&env.DatabaseDriver),
		"DATABASE_DSN":    StringLoader(&env.DatabaseDSN),
		"SESSION_KEY":     SecretLoader(&env.SessionKey),
		"SMTP_HOST":       StringLoader(&env.SMTP.Host),
		"SMTP_PORT":       IntLoader(&env.SMTP.Port),
		"SMTP_USERNAME":   StringLoader(&env.SMTP.Username),
		"SMTP_PASSWORD":   StringLoader(&env.SMTP.Password),
	}

	for key, load := range loaders {
		found, ok := os.LookupEnv(key)
		if !ok {
			log.Fatalf("undefined environment variable %q", key)
		}

		if err := load(&env, found); err != nil {
			return env, err
		}
	}

	return env, nil
}

func StringLoader(field *string) Loader {
	return func(env *Env, value string) error {
		*field = value
		return nil
	}
}

func SliceLoader(separator string, field *[]string) Loader {
	return func(env *Env, value string) error {
		elements := strings.Split(value, separator)
		*field = elements
		return nil
	}
}

func IntLoader(field *int) Loader {
	return func(env *Env, value string) error {
		intVal, err := strconv.Atoi(value)
		if err != nil {
			return err
		}

		*field = intVal
		return nil
	}
}

func SecretLoader(field *string) Loader {
	const MinSecretBytes = 64
	return func(env *Env, value string) error {
		if len(value) < MinSecretBytes {
			return fmt.Errorf("secret is not secure enougth : %d bytes minimum required for best security", MinSecretBytes)
		}

		*field = value
		return nil
	}
}
