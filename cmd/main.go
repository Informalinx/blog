package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
)

type Env struct {
	ServerAddress string
	DatabaseDriver string
	DatabaseDSN string
}

func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Fatal("error while loading .env file : ", err)
	}

	env := Env{}
	lookup := map[string]*string{
		"SERVER_ADDRESS": &env.ServerAddress,
		"DATABASE_DRIVER": &env.DatabaseDriver,
		"DATABASE_DSN": &env.DatabaseDSN,
	}

	for key, val := range lookup {
		found, ok := os.LookupEnv(key)
		if !ok {
			log.Fatalf("undefined environment variable %q", key)
		}

		*val = found
	}

	db, err := sql.Open(env.DatabaseDriver, env.DatabaseDSN)
	if err != nil {
		log.Fatalf("error while connecting to database : %s", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("cannot ping database : %s", err)
	}

	mux := http.NewServeMux()

	mux.Handle("/{$}", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, World !"))
	}))

	fmt.Println("Server listening on :", env.ServerAddress)
	if err := http.ListenAndServe(env.ServerAddress, mux); err != nil {
		log.Fatal(err)
	}
}
