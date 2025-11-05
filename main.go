package main

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// Job represents a download task
type Job struct {
	URL string
	ID  int
}

func main() {
	// Step 1: Read URLs from file
	urls, err := readURLs("urls.txt")
	if err != nil {
		fmt.Println("Error reading URLs:", err)
		return
	}

	// Step 2: Create downloads directory
	os.MkdirAll("downloads", 0755)

	// Create job channel
	jobs := make(chan Job, len(urls))

	var wg sync.WaitGroup

	// Start 5 workers
	numWorkers := 5
	for w := 1; w <= numWorkers; w++ {
		wg.Add(1)
		go worker(w, jobs, &wg)
	}

	// Send all jobs to channel
	for i, url := range urls {
		jobs <- Job{URL: url, ID: i}
	}
	close(jobs) // No more jobs

	wg.Wait()
	fmt.Println("All downloads complete!")
}

func worker(id int, jobs <-chan Job, wg *sync.WaitGroup) {
	defer wg.Done()

	// Keep taking jobs from channel until it's closed
	for job := range jobs {
		fmt.Printf("Worker %d: starting download %d\n", id, job.ID)
		downloadFile(job.URL, job.ID)
	}

	fmt.Printf("Worker %d: finished\n", id)
}

func downloadFile(url string, id int) {
	fmt.Printf("Starting download %d: %s\n", id, url)

	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("Error downloading %s: %v\n", url, err)
		return
	}
	defer resp.Body.Close()

	filename := filepath.Join("downloads", fmt.Sprintf("file_%d.json", id))
	outFile, err := os.Create(filename)
	if err != nil {
		fmt.Printf("Error creating file %s: %v\n", filename, err)
		return
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, resp.Body)
	if err != nil {
		fmt.Printf("Error saving file %s: %v\n", filename, err)
		return
	}

	fmt.Printf("Completed download %d: %s\n", id, filename)
}

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
		if line != "" {
			urls = append(urls, line)
		}
	}
	return urls, scanner.Err()
}
