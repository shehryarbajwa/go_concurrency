package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"concurrent-downloader/config"
	"concurrent-downloader/database"
	"concurrent-downloader/models"
	"concurrent-downloader/worker"
)

func main() {
	fmt.Println("🚀 Starting concurrent downloader...")

	// Load configuration
	cfg := config.NewConfig()

	// Initialize database
	db, err := database.NewDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Read URLs from file
	urls, err := readURLs("urls.txt")
	if err != nil {
		log.Fatalf("Failed to read URLs: %v", err)
	}
	fmt.Printf("📋 Loaded %d URLs\n", len(urls))

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("\n🛑 Received shutdown signal, waiting for workers to finish...")
		cancel()
	}()

	// Create jobs channel
	jobs := make(chan models.Job, len(urls))

	// Start worker pool
	var wg sync.WaitGroup
	fmt.Printf("👷 Starting %d workers...\n", cfg.NumWorkers)
	worker.StartWorkerPool(ctx, cfg.NumWorkers, jobs, db, &wg)

	// Send all jobs to channel
	for i, url := range urls {
		jobs <- models.Job{
			URL: url,
			ID:  i + 1,
		}
	}
	close(jobs) // No more jobs

	// Wait for all workers to finish
	wg.Wait()

	fmt.Println("\n✅ All downloads complete!")
}

// readURLs reads URLs from a file
func readURLs(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var urls []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			urls = append(urls, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return urls, nil
}
