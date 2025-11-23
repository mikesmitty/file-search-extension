package cmd

import (
	"context"
	"sync"
)

// BatchOptions provides configuration for batch processing.
type BatchOptions struct {
	Concurrency int // Number of parallel operations (default: 5)
	Quiet       bool
	OnProgress  func(current, total int, file string, err error)
}

// BatchResult holds the outcome of a batch processing operation.
type BatchResult struct {
	Succeeded []string
	Failed    map[string]error
	Total     int
}

// processBatch processes a slice of files concurrently.
// It takes a context, a list of files, a processor function for each file, and batch options.
// The processor function should return an error if the processing of a single file fails.
// It returns a BatchResult summarizing the operation.
func processBatch(ctx context.Context, files []string, processor func(ctx context.Context, file string) error, opts *BatchOptions) *BatchResult {
	result := &BatchResult{
		Succeeded: make([]string, 0),
		Failed:    make(map[string]error),
		Total:     len(files),
	}

	if len(files) == 0 {
		return result
	}

	if opts == nil {
		opts = &BatchOptions{}
	}
	if opts.Concurrency <= 0 {
		opts.Concurrency = 5 // Default concurrency
	}

	var (
		wg          sync.WaitGroup
		mu          sync.Mutex // Protects result, fileIdx, and progress updates
		inProgress  = make(chan struct{}, opts.Concurrency)
		processedMu sync.Mutex // Protects processedCount
		processedCount int
	)

	for _, file := range files {
		inProgress <- struct{}{} // Acquire a slot

		wg.Add(1)
		go func(f string) {
			defer func() {
				<-inProgress // Release the slot
				wg.Done()
			}()

			err := processor(ctx, f)

			mu.Lock()
			processedMu.Lock()
			processedCount++
			current := processedCount
			mu.Unlock()
			processedMu.Unlock()
			
			if opts.OnProgress != nil && !opts.Quiet {
				opts.OnProgress(current, result.Total, f, err)
			}

			mu.Lock()
			if err != nil {
				result.Failed[f] = err
			} else {
				result.Succeeded = append(result.Succeeded, f)
			}
			mu.Unlock()
		}(file)
	}

	wg.Wait()
	return result
}
