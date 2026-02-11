package poll

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestPoll(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		check   func(ctx context.Context) (string, bool, error)
		want    string
		wantErr string
	}{
		{
			name: "ImmediateSuccess",
			cfg:  Config{Interval: 10 * time.Millisecond, Timeout: time.Second},
			check: func(_ context.Context) (string, bool, error) {
				return "done", true, nil
			},
			want: "done",
		},
		{
			name: "SuccessAfterRetries",
			cfg:  Config{Interval: 10 * time.Millisecond, Timeout: time.Second},
			check: func() func(context.Context) (string, bool, error) {
				calls := 0
				return func(_ context.Context) (string, bool, error) {
					calls++
					if calls >= 3 {
						return "done", true, nil
					}
					return "", false, nil
				}
			}(),
			want: "done",
		},
		{
			name: "Error",
			cfg:  Config{Interval: 10 * time.Millisecond, Timeout: time.Second},
			check: func(_ context.Context) (string, bool, error) {
				return "", false, errors.New("check failed")
			},
			wantErr: "check failed",
		},
		{
			name: "Timeout",
			cfg:  Config{Interval: 10 * time.Millisecond, Timeout: 50 * time.Millisecond},
			check: func(_ context.Context) (string, bool, error) {
				return "", false, nil
			},
			wantErr: "polling timed out",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Poll(context.Background(), tt.cfg, tt.check)
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("error = %q, want containing %q", err.Error(), tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.want {
				t.Errorf("result = %q, want %q", result, tt.want)
			}
		})
	}
}

func TestPoll_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	cfg := Config{Interval: 10 * time.Millisecond, Timeout: time.Second}
	_, err := Poll(ctx, cfg, func(_ context.Context) (string, bool, error) {
		return "", false, nil
	})

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("error = %v, want context.Canceled", err)
	}
}
