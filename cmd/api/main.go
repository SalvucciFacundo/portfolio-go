package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/SalvucciFacundo/portfolio-go/internal/router"
)

func main() {
	mux := http.NewServeMux()

	router.Register(mux)

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080" // default local
	}

	fmt.Printf("Listening on :%s\n", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
