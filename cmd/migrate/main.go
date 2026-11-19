package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"backoffice-service/config"

	"github.com/jackc/pgx/v5"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	fmt.Printf("Connecting to database: %s\n", cfg.DB.MasterDSN)
	conn, err := pgx.Connect(ctx, cfg.DB.MasterDSN)
	if err != nil {
		fmt.Printf("Database connection failed: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close(ctx)

	// Read migration files
	migrationFiles, err := filepath.Glob("migrations/*.up.sql")
	if err != nil {
		fmt.Printf("Failed to find migration files: %v\n", err)
		os.Exit(1)
	}

	sort.Strings(migrationFiles)

	for _, file := range migrationFiles {
		fmt.Printf("Applying migration: %s\n", file)
		content, err := os.ReadFile(file)
		if err != nil {
			fmt.Printf("Failed to read %s: %v\n", file, err)
			os.Exit(1)
		}

		sqlStr := string(content)
		if strings.TrimSpace(sqlStr) == "" {
			continue
		}

		_, err = conn.Exec(ctx, sqlStr)
		if err != nil {
			fmt.Printf("Migration failed in %s: %v\n", file, err)
			os.Exit(1)
		}
		fmt.Printf("Successfully applied %s\n", file)
	}

	fmt.Println("All migrations applied successfully!")
}

