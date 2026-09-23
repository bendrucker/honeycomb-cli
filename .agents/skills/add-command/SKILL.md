---
name: add-command
description: Add a new cobra subcommand to the CLI
user-invocable: true
---

# Add Command

Add a cobra subcommand that matches its siblings. The command is done when it builds, its tests pass, and it appears under its parent in `honeycomb --help`.

## Steps

1. Read the sibling commands in the resource's package (`cmd/<resource>/`) and copy their shape. `cmd/marker` covers full CRUD plus a nested sub-resource.
2. Create `cmd/<resource>/<verb>.go` with a `New<Verb>Cmd(opts *options.RootOptions, ...)` factory:
   - Take any parent-scoped flag, such as `--dataset`, as a pointer argument.
   - Give every input a flag. Resolve missing required inputs with `command.Resolve`, which prompts when it can and returns the non-interactive error otherwise.
   - Build the client with `opts.ClientFor`.
   - Write output with `opts.OutputWriter()` for detail commands or `opts.OutputWriterList()` for lists. `--format` is a root persistent flag, so the command defines no format flag.
3. Register it with `AddCommand` in the resource's `NewCmd`.
4. Create `cmd/<resource>/<verb>_test.go` using the package's existing test helper (usually `setupTest`), which wires `iostreams.Test`, `httptest.NewServer`, and the mock keyring.

## Command Pattern

```go
func NewListCmd(opts *options.RootOptions, dataset *string) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List markers",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runMarkerList(cmd.Context(), opts, *dataset)
		},
	}
}
```
