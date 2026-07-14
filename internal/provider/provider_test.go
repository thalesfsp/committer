package provider

import (
	"context"
	"expvar"
	"testing"
	"time"

	"github.com/thalesfsp/inference/provider"
	"github.com/thalesfsp/sypl/v2"
	"github.com/thalesfsp/sypl/v2/level"
)

// mockProvider implements provider.IProvider for testing.
type mockProvider struct {
	completionFunc func(ctx context.Context, options ...provider.Func) (string, error)
}

func (m *mockProvider) Completion(ctx context.Context, options ...provider.Func) (string, error) {
	return m.completionFunc(ctx, options...)
}

func (m *mockProvider) GetClient() any                          { return nil }
func (m *mockProvider) GetLogger() sypl.ISypl                   { return sypl.NewDefault("test", level.Info) }
func (m *mockProvider) GetName() string                         { return "mock" }
func (m *mockProvider) GetType() string                         { return "mock" }
func (m *mockProvider) GetCounterCompletion() *expvar.Int       { return expvar.NewInt("mock_completion") }
func (m *mockProvider) GetCounterCompletionFailed() *expvar.Int { return expvar.NewInt("mock_failed") }

// TestCallLLM_RespectsParentContext verifies that CallLLM propagates the
// parent context rather than discarding it.
//
// This is a regression test for a bug where CallLLM used context.Background()
// instead of the passed ctx parameter, breaking context cancellation.
func TestCallLLM_RespectsParentContext(t *testing.T) {
	t.Run("cancelled parent context propagates to provider", func(t *testing.T) {
		// Create a context that is already cancelled.
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately.

		mock := &mockProvider{
			completionFunc: func(ctx context.Context, options ...provider.Func) (string, error) {
				// The context passed to Completion should be derived from
				// the parent context. Since the parent is cancelled, this
				// derived context should also be done.
				select {
				case <-ctx.Done():
					return "", ctx.Err()
				default:
					t.Error("expected context to be cancelled, but it was not")

					return "should not reach", nil
				}
			},
		}

		_, err := CallLLM(ctx, mock, 5*time.Second, "test prompt")
		if err == nil {
			t.Error("expected error from cancelled context, got nil")
		}
	})

	t.Run("context values propagate to provider", func(t *testing.T) {
		type ctxKey string
		key := ctxKey("test-key")
		expectedVal := "test-value"

		ctx := context.WithValue(context.Background(), key, expectedVal)

		mock := &mockProvider{
			completionFunc: func(ctx context.Context, options ...provider.Func) (string, error) {
				// The context should carry values from the parent context.
				val, ok := ctx.Value(key).(string)
				if !ok || val != expectedVal {
					t.Errorf("expected context value %q, got %q (ok=%v)", expectedVal, val, ok)
				}

				return "commit message", nil
			},
		}

		result, err := CallLLM(ctx, mock, 5*time.Second, "test prompt")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result != "commit message" {
			t.Errorf("expected 'commit message', got %q", result)
		}
	})

	t.Run("timeout is applied on top of parent context", func(t *testing.T) {
		ctx := context.Background()
		shortTimeout := 50 * time.Millisecond

		mock := &mockProvider{
			completionFunc: func(ctx context.Context, options ...provider.Func) (string, error) {
				// Verify deadline is set.
				deadline, ok := ctx.Deadline()
				if !ok {
					t.Error("expected context to have a deadline")

					return "", nil
				}

				// The deadline should be roughly shortTimeout from now.
				remaining := time.Until(deadline)
				if remaining > shortTimeout {
					t.Errorf("deadline too far in the future: %v", remaining)
				}

				return "result", nil
			},
		}

		result, err := CallLLM(ctx, mock, shortTimeout, "test")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result != "result" {
			t.Errorf("expected 'result', got %q", result)
		}
	})
}

// TestChunkDiff verifies diff chunking behavior.
func TestChunkDiff(t *testing.T) {
	t.Run("small diff returns single chunk", func(t *testing.T) {
		diff := "small change"
		chunks, err := ChunkDiff(1000, diff)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(chunks) != 1 {
			t.Errorf("expected 1 chunk, got %d", len(chunks))
		}
		if chunks[0] != diff {
			t.Errorf("expected chunk to equal input diff")
		}
	})
}
