// cmd/seed/main.go — creates or updates the first superadmin user.
//
// Usage:
//
//	go run ./cmd/seed --email admin@example.com --password secret123 [--name "Admin User"]
//
// Environment variables (fallbacks):
//
//	DATABASE_URL — PostgreSQL DSN
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/bivex/paywall-iap/internal/infrastructure/persistence/sqlc/generated"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	var (
		dbURL    string
		email    string
		password string
		name     string
	)

	flag.StringVar(&dbURL, "database", os.Getenv("DATABASE_URL"), "PostgreSQL connection string")
	flag.StringVar(&email, "email", "", "Admin email address (required)")
	flag.StringVar(&password, "password", "", "Admin password (required, min 8 chars)")
	flag.StringVar(&name, "name", "Admin", "Display name (stored as platform_user_id)")
	flag.Parse()

	if dbURL == "" {
		log.Fatal("DATABASE_URL is required (flag --database or env var)")
	}
	if email == "" || password == "" {
		log.Fatal("--email and --password are required")
	}
	if len(password) < 8 {
		log.Fatal("--password must be at least 8 characters")
	}

	ctx := context.Background()

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("Cannot connect to database: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("Database ping failed: %v", err)
	}

	q := generated.New(pool)

	// Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}

	// Upsert user (insert or get existing by email)
	platformUserID := "admin_web_" + strings.ReplaceAll(email, "@", "_at_")

	user, err := q.GetUserByEmail(ctx, email)
	if err != nil {
		// User doesn't exist — create new superadmin
		user, err = q.CreateUser(ctx, generated.CreateUserParams{
			PlatformUserID: platformUserID,
			DeviceID:       nil,
			Platform:       "web",
			AppVersion:     "1.0.0",
			Email:          email,
			Role:           "superadmin",
		})
		if err != nil {
			log.Fatalf("Failed to create user: %v", err)
		}
		fmt.Printf("✅ Created new superadmin user: %s (id: %s)\n", email, user.ID)
	} else {
		// Existing user — ensure they have superadmin role
		if user.Role != "admin" && user.Role != "superadmin" {
			_, err = q.UpdateUserRole(ctx, generated.UpdateUserRoleParams{
				ID:   user.ID,
				Role: "superadmin",
			})
			if err != nil {
				log.Fatalf("Failed to update user role: %v", err)
			}
		}
		fmt.Printf("✅ Found existing user: %s (id: %s, role: %s)\n", email, user.ID, user.Role)
	}

	// Upsert password hash
	_, err = q.UpsertAdminCredential(ctx, user.ID, string(hash))
	if err != nil {
		log.Fatalf("Failed to store admin credential: %v", err)
	}

	fmt.Printf("✅ Password set for %s\n", email)
	fmt.Printf("\n🔑 Admin credentials ready:\n   Email:    %s\n   Password: %s\n   User ID:  %s\n",
		email, maskPassword(password), uuid.UUID(user.ID).String())
	fmt.Println("\n👉 Login at: http://localhost:3000/auth/v1/login")
}

func maskPassword(p string) string {
	if len(p) <= 2 {
		return "***"
	}
	return p[:2] + strings.Repeat("*", len(p)-2)
}
