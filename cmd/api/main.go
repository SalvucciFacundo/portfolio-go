package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/term"

	"github.com/SalvucciFacundo/portfolio-go/internal/adapters/db"
	"github.com/SalvucciFacundo/portfolio-go/internal/adapters/dbmigrate"
	"github.com/SalvucciFacundo/portfolio-go/internal/auth"
	"github.com/SalvucciFacundo/portfolio-go/internal/handler"
	"github.com/SalvucciFacundo/portfolio-go/internal/router"
)

// Uso:
//
//	server [serve]                      → arranca el HTTP server (default)
//	server create-admin [--username X]  → crea admin (pide password interactivo)
func main() {
	ctx := context.Background()

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL no configurada — revisar .env")
	}

	pool, err := dbmigrate.NewPool(ctx, databaseURL)
	if err != nil {
		log.Fatalf("conectar a DB: %v", err)
	}
	defer pool.Close()

	if len(os.Args) > 1 && os.Args[1] == "create-admin" {
		runCreateAdmin(ctx, pool)
		return
	}
	runServe(ctx, pool)
}

// runCreateAdmin crea el usuario admin. El password se pide interactivamente
// (sin eco) y se confirma dos veces; el flag --password existe SOLO para
// testing/CI (documentado) y jamás debe usarse en producción.
func runCreateAdmin(ctx context.Context, pool *pgxpool.Pool) {
	fs := flag.NewFlagSet("create-admin", flag.ExitOnError)
	username := fs.String("username", "admin", "nombre de usuario admin")
	passwordFlag := fs.String("password", "", "password (SOLO testing/CI; en uso normal se pide interactivamente)")
	_ = fs.Parse(os.Args[2:])

	if err := dbmigrate.Migrate(ctx, pool); err != nil {
		log.Fatalf("migraciones: %v", err)
	}

	password := *passwordFlag
	if password == "" {
		var err error
		password, err = readPasswordTwice()
		if err != nil {
			log.Fatal(err)
		}
	}

	svc := auth.NewService(db.NewAuthRepo(pool))
	if err := svc.CreateAdmin(ctx, *username, password); err != nil {
		if errors.Is(err, auth.ErrUsernameTaken) {
			log.Fatalf("error: %v", err)
		}
		log.Fatalf("crear admin: %v", err)
	}
	fmt.Printf("Admin %q creado correctamente\n", *username)
}

// runServe conecta, migra, inyecta el Service de auth en el bridge HTMX y
// arranca el HTTP server.
func runServe(ctx context.Context, pool *pgxpool.Pool) {
	if err := dbmigrate.Migrate(ctx, pool); err != nil {
		log.Fatalf("migraciones: %v", err)
	}
	fmt.Println("Migraciones aplicadas OK")

	svc := auth.NewService(db.NewAuthRepo(pool))
	handler.SetupAuth(svc)

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

// readPasswordTwice pide el password sin eco, lo valida (>= 8 chars) y exige
// confirmación idéntica.
func readPasswordTwice() (string, error) {
	first, err := readPassword("Password: ")
	if err != nil {
		return "", err
	}
	if len(first) < 8 {
		return "", fmt.Errorf("password debe tener al menos 8 caracteres")
	}
	confirm, err := readPassword("Confirm password: ")
	if err != nil {
		return "", err
	}
	if first != confirm {
		return "", fmt.Errorf("las passwords no coinciden")
	}
	return first, nil
}

func readPassword(prompt string) (string, error) {
	fmt.Print(prompt)
	b, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println()
	if err != nil {
		return "", fmt.Errorf("leer password: %w", err)
	}
	return string(b), nil
}
