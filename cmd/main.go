package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/informalinx/blog/internal/env"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	conf, err := env.Load(".env")
	if err != nil {
		log.Fatal(err)
	}

	_ = sessions.NewCookieStore([]byte(conf.SessionKey))

	db, err := sql.Open(conf.DatabaseDriver, conf.DatabaseDSN)
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

	fmt.Println("Server listening on :", conf.ServerAddress)
	if err := http.ListenAndServe(conf.ServerAddress, mux); err != nil {
		log.Fatal(err)
	}
}
