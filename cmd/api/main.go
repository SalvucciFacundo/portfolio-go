package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/SalvucciFacundo/portfolio-go/internal/adapters/dbmigrate"
	"github.com/SalvucciFacundo/portfolio-go/internal/router"
)

func main() {
	ctx := context.Background()

	// Conexión a PostgreSQL (requerida)
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL no configurada — revisar .env")
	}
	pool, err := dbmigrate.NewPool(ctx, databaseURL)
	if err != nil {
		log.Fatalf("conectar a DB: %v", err)
	}
	defer pool.Close()

	// Aplicar migraciones
	if err := dbmigrate.Migrate(ctx, pool); err != nil {
		log.Fatalf("migraciones: %v", err)
	}
	fmt.Println("Migraciones aplicadas OK")

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
