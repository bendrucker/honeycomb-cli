---
name: add-command
description: Add a new cobra subcommand to the CLI
user_invocable: true
---

# Add Command

Add a new cobra subcommand following the project's patterns.

## Steps

1. Read existing commands in `cmd/` to understand patterns
2. Create `cmd/<name>.go` with a `New<Name>Cmd` factory function
3. The factory accepts a parent options struct or `*iostreams.IOStreams`
4. Register the command with its parent via `AddCommand` in the parent's factory
5. Add `--format` output support (json, table, yaml)
6. Implement both interactive (TTY prompts) and non-interactive (flags-only) paths
7. Create `cmd/<name>_test.go` with tests using `iostreams.Test()`

## Command Pattern

```go
func NewExampleCmd(ios *iostreams.IOStreams) *cobra.Command {
	opts := &ExampleOptions{IOStreams: ios}

	cmd := &cobra.Command{
		Use:   "example",
		Short: "One-line description",
		RunE: func(_ *cobra.Command, _ []string) error {
			return runExample(opts)
		},
	}

	cmd.Flags().StringVar(&opts.Format, "format", "", "Output format: json, table, yaml")

	return cmd
}

func runExample(opts *ExampleOptions) error {
	// Implementation
	return nil
}
```

## Checklist

- [ ] `New*Cmd` factory, no `init()`
- [ ] `SilenceUsage: true` on subcommands that have `RunE`
- [ ] Registered with parent command
- [ ] `--format` flag if command produces output
- [ ] Interactive path with prompts (when `ios.CanPrompt()`)
- [ ] Non-interactive path requiring all flags
- [ ] Test file with `iostreams.Test()`
