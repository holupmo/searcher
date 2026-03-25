package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
)

type FileJob struct {
	Path string
	Size int64
}

type FileResult struct {
	Path string
	Size int64
	Hash string
	Err  error
}

func hashFile(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hasher := sha256.New()

	_, err = io.Copy(hasher, file)
	if err != nil {
		return "", err
	}

	hash := fmt.Sprintf("%x", hasher.Sum(nil))
	return hash, nil
}

func worker(jobs <-chan FileJob, results chan<- FileResult, wg *sync.WaitGroup) {
	defer wg.Done()

	for job := range jobs {
		hash, err := hashFile(job.Path)

		result := FileResult{
			Path: job.Path,
			Size: job.Size,
			Hash: hash,
			Err:  err,
		}

		results <- result
	}
}

func main() {
	root := ""

	jobs := make(chan FileJob, 100)
	results := make(chan FileResult, 100)

	var wg sync.WaitGroup

	workerCount := 4

	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go worker(jobs, results, &wg)
	}

	go func() {
		filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}

			if !info.IsDir() {
				jobs <- FileJob{
					Path: path,
					Size: info.Size(),
				}
			}

			return nil
		})

		close(jobs)
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	totalFiles := 0
	var totalSize float64
	var totalSize1 int64


	for result := range results {
		sizeMB := float64(result.Size) / (1024 * 1024)
		if result.Err != nil {
			fmt.Println("Error:", result.Err)
			continue
		}

		fmt.Printf("File: %s\n", result.Path)
		fmt.Printf("Size: %.2f MB\n", sizeMB)
		fmt.Printf("SHA256: %s\n\n", result.Hash)

		totalFiles++
		totalSize += sizeMB
		totalSize1 += result.Size
	}

	fmt.Println("-----")
	fmt.Println("Total files:", totalFiles)
	fmt.Println("Total size:", totalSize1, "bytes, or", totalSize, "MB")
}
