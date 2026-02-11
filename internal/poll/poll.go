package poll

import (
	"context"
	"fmt"
	"time"

	"github.com/charmbracelet/huh/spinner"
)

type Config struct {
	// Interval between poll attempts. Default: 1s.
	Interval time.Duration
	// Timeout for the entire polling operation. Default: 2m.
	Timeout time.Duration
	// Title shown in the spinner (interactive mode only).
	Title string
	// Interactive enables the huh/spinner for TTY sessions.
	Interactive bool
}

func (c *Config) defaults() {
	if c.Interval == 0 {
		c.Interval = time.Second
	}
	if c.Timeout == 0 {
		c.Timeout = 2 * time.Minute
	}
	if c.Title == "" {
		c.Title = "Loading..."
	}
}

func Poll[T any](ctx context.Context, cfg Config, check func(ctx context.Context) (T, bool, error)) (T, error) {
	cfg.defaults()

	ctx, cancel := context.WithTimeout(ctx, cfg.Timeout)
	defer cancel()

	if cfg.Interactive {
		return pollInteractive(ctx, cfg, check)
	}
	result, _, err := pollLoop(ctx, cfg, check)
	return result, err
}

func pollInteractive[T any](ctx context.Context, cfg Config, check func(ctx context.Context) (T, bool, error)) (T, error) {
	var result T
	var pollErr error

	s := spinner.New().
		Title(cfg.Title).
		Context(ctx).
		ActionWithErr(func(ctx context.Context) error {
			var done bool
			result, done, pollErr = pollLoop(ctx, cfg, check)
			if pollErr != nil {
				return pollErr
			}
			if done {
				return nil
			}
			return ctx.Err()
		})

	if err := s.Run(); err != nil {
		if pollErr != nil {
			return result, pollErr
		}
		return result, err
	}
	return result, pollErr
}

func pollLoop[T any](ctx context.Context, cfg Config, check func(ctx context.Context) (T, bool, error)) (T, bool, error) {
	ticker := time.NewTicker(cfg.Interval)
	defer ticker.Stop()

	for {
		result, done, err := check(ctx)
		if err != nil {
			return result, false, err
		}
		if done {
			return result, true, nil
		}

		select {
		case <-ctx.Done():
			var zero T
			if ctx.Err() == context.DeadlineExceeded {
				return zero, false, fmt.Errorf("polling timed out")
			}
			return zero, false, ctx.Err()
		case <-ticker.C:
		}
	}
}
