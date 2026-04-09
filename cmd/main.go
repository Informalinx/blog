package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Fatal("error while loading .env file : ", err)
	}

	addr, ok := os.LookupEnv("SERVER_ADDRESS")
	if !ok {
		log.Fatal("undefined environment variable \"SERVER_ADDRESS\"")
	}

	mux := http.NewServeMux()

	mux.Handle("/{$}", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, World !"))
	}))

	fmt.Println("Server listening on :", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}
