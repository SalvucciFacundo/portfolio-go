package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/SalvucciFacundo/portfolio-go/internal/router"
)

func main() {
	mux := http.NewServeMux()

	router.Register(mux)

	fmt.Println("Listening on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
