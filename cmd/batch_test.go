package cmd

import (
	"context"
	"fmt"
	"sort"
	"sync/atomic"
	"testing"
	"time"
)

func TestProcessBatch(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		files       []string
		concurrency int
		processor   func(ctx context.Context, file string) error
		wantSucceed int
		wantFailed  int
		wantError   bool
	}{
		{
			name:        "empty file list",
			files:       []string{},
			concurrency: 1,
			processor:   func(ctx context.Context, file string) error { return nil },
			wantSucceed: 0,
			wantFailed:  0,
			wantError:   false,
		},
		{
			name:        "single file success",
			files:       []string{"file1.txt"},
			concurrency: 1,
			processor:   func(ctx context.Context, file string) error { return nil },
			wantSucceed: 1,
			wantFailed:  0,
			wantError:   false,
		},
		{
			name:        "single file failure",
			files:       []string{"file1.txt"},
			concurrency: 1,
			processor:   func(ctx context.Context, file string) error { return fmt.Errorf("failed") },
			wantSucceed: 0,
			wantFailed:  1,
			wantError:   true,
		},
		{
			name:        "multiple files all success",
			files:       []string{"file1.txt", "file2.txt", "file3.txt"},
			concurrency: 2,
			processor:   func(ctx context.Context, file string) error { return nil },
			wantSucceed: 3,
			wantFailed:  0,
			wantError:   false,
		},
		{
			name:        "multiple files all failure",
			files:       []string{"file1.txt", "file2.txt", "file3.txt"},
			concurrency: 2,
			processor:   func(ctx context.Context, file string) error { return fmt.Errorf("failed: %s", file) },
			wantSucceed: 0,
			wantFailed:  3,
			wantError:   true,
		},
		{
			name:        "multiple files mixed success and failure",
			files:       []string{"file1.txt", "file2.txt", "file3.txt", "file4.txt"},
			concurrency: 2,
			processor: func(ctx context.Context, file string) error {
				if file == "file2.txt" || file == "file4.txt" {
					return fmt.Errorf("failed: %s", file)
				}
				return nil
			},
			wantSucceed: 2,
			wantFailed:  2,
			wantError:   true,
		},
		{
			name:        "high concurrency",
			files:       []string{"f1", "f2", "f3", "f4", "f5", "f6", "f7", "f8", "f9", "f10"},
			concurrency: 10,
			processor:   func(ctx context.Context, file string) error { time.Sleep(10 * time.Millisecond); return nil },
			wantSucceed: 10,
			wantFailed:  0,
			wantError:   false,
		},
		{
			name:        "low concurrency",
			files:       []string{"f1", "f2", "f3", "f4", "f5", "f6", "f7", "f8", "f9", "f10"},
			concurrency: 1,
			processor:   func(ctx context.Context, file string) error { time.Sleep(10 * time.Millisecond); return nil },
			wantSucceed: 10,
			wantFailed:  0,
			wantError:   false,
		},
		{
			name:        "concurrency boundary - exact number of files",
			files:       []string{"f1", "f2", "f3", "f4", "f5"},
			concurrency: 5,
			processor:   func(ctx context.Context, file string) error { return nil },
			wantSucceed: 5,
			wantFailed:  0,
			wantError:   false,
		},
		{
			name:        "concurrency boundary - more concurrency than files",
			files:       []string{"f1", "f2", "f3"},
			concurrency: 5,
			processor:   func(ctx context.Context, file string) error { return nil },
			wantSucceed: 3,
			wantFailed:  0,
			wantError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var progressCalls int32
			opts := &BatchOptions{
				Concurrency: tt.concurrency,
				Quiet:       false,
				OnProgress: func(current, total int, file string, err error) {
					atomic.AddInt32(&progressCalls, 1)
					if current <= 0 || current > total || total != len(tt.files) {
						t.Errorf("OnProgress called with invalid current/total: %d/%d for file %s", current, total, file)
					}
				},
			}

			result := processBatch(ctx, tt.files, tt.processor, opts)

			if result.Total != len(tt.files) {
				t.Errorf("processBatch() Total = %v, want %v", result.Total, len(tt.files))
			}
			if len(result.Succeeded) != tt.wantSucceed {
				t.Errorf("processBatch() Succeeded = %v, want %v", len(result.Succeeded), tt.wantSucceed)
			}
			if len(result.Failed) != tt.wantFailed {
				t.Errorf("processBatch() Failed = %v, want %v", len(result.Failed), tt.wantFailed)
			}

			// Check if OnProgress was called for each file if not quiet
			if !opts.Quiet && len(tt.files) > 0 {
				if int(atomic.LoadInt32(&progressCalls)) != len(tt.files) {
					t.Errorf("OnProgress callback count mismatch: got %d, want %d", atomic.LoadInt32(&progressCalls), len(tt.files))
				}
			}

			// Verify contents of Succeeded and Failed maps/slices
			var sortedSucceeded []string
			for _, s := range result.Succeeded {
				sortedSucceeded = append(sortedSucceeded, s)
			}
			sort.Strings(sortedSucceeded)

			var expectedSucceeded []string
			for _, f := range tt.files {
				if _, ok := tt.processor(ctx, f).(error); !ok {
					expectedSucceeded = append(expectedSucceeded, f)
				}
			}
			sort.Strings(expectedSucceeded)

			if fmt.Sprintf("%v", sortedSucceeded) != fmt.Sprintf("%v", expectedSucceeded) {
				t.Errorf("processBatch() Succeeded files = %v, want %v", sortedSucceeded, expectedSucceeded)
			}

			var sortedFailed []string
			for f := range result.Failed {
				sortedFailed = append(sortedFailed, f)
			}
			sort.Strings(sortedFailed)

			var expectedFailed []string
			for _, f := range tt.files {
				if tt.processor(ctx, f) != nil {
					expectedFailed = append(expectedFailed, f)
				}
			}
			sort.Strings(expectedFailed)

			if fmt.Sprintf("%v", sortedFailed) != fmt.Sprintf("%v", expectedFailed) {
				t.Errorf("processBatch() Failed files = %v, want %v", sortedFailed, expectedFailed)
			}

		})
	}
}

func TestProcessBatch_ContextCancellation(t *testing.T) {
	files := []string{"f1", "f2", "f3", "f4", "f5"}
	slowProcessor := func(ctx context.Context, file string) error {
		select {
		case <-time.After(100 * time.Millisecond):
			// Simulate work
		case <-ctx.Done():
			return ctx.Err()
		}
		return nil
	}

	t.Run("cancellation stops pending tasks", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		var processedCount atomic.Int32
		opts := &BatchOptions{
			Concurrency: 1, // Ensure sequential processing for predictable cancellation
			OnProgress: func(current, total int, file string, err error) {
				processedCount.Add(1)
				if current == 2 { // Cancel after the second file starts processing
					cancel()
				}
			},
		}

		result := processBatch(ctx, files, slowProcessor, opts)

		// Expect f1 and f2 to be processed (f2 might be cancelled mid-way, or just before returning)
		// It's hard to precisely predict how many will *succeed* when cancelled
		// but we expect not all to succeed.
		if len(result.Succeeded)+len(result.Failed) != int(processedCount.Load()) {
			t.Errorf("Expected total processed files %d, got %d", processedCount.Load(), len(result.Succeeded)+len(result.Failed))
		}
		if len(result.Succeeded) == len(files) {
			t.Errorf("Expected some files to be cancelled, but all succeeded")
		}
	})

	t.Run("no cancellation finishes all tasks", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel() // Ensure context is cancelled eventually, but not prematurely

		var processedCount int32
		opts := &BatchOptions{
			Concurrency: 1,
			OnProgress: func(current, total int, file string, err error) {
				atomic.AddInt32(&processedCount, 1)
			},
		}

		result := processBatch(ctx, files, slowProcessor, opts)

		if len(result.Succeeded) != len(files) {
			t.Errorf("Expected all files to succeed, but got %d succeeded", len(result.Succeeded))
		}
		if int(processedCount) != len(files) {
			t.Errorf("Expected OnProgress to be called for all files, but got %d calls", processedCount)
		}
	})
}
