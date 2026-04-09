package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	addr, ok := os.LookupEnv("SERVER_ADDRESS")
	if !ok {
		addr = "127.0.0.1:8080"
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
