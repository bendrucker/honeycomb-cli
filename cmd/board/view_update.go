package board

import (
	"bytes"
	"context"
	"fmt"

	"github.com/bendrucker/honeycomb-cli/cmd/command"
	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/deref"
	"github.com/spf13/cobra"
)

func NewViewUpdateCmd(opts *options.RootOptions, board *string) *cobra.Command {
	var (
		file       string
		name       string
		filterArgs []string
	)

	cmd := &cobra.Command{
		Use:   "update <view-id>",
		Short: "Update a board view",
		Example: `  # Rename a view
  honeycomb board view update view123 --board abc123 --name "Errors"

  # Replace a view's filters
  honeycomb board view update view123 --board abc123 \
    --filter "status_code:>=:500"`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runViewUpdate(cmd, opts, *board, args[0], file, name, filterArgs)
		},
	}

	cmd.Flags().StringVarP(&file, "file", "f", "", "Path to JSON file (- for stdin)")
	cmd.Flags().StringVar(&name, "name", "", "View name")
	cmd.Flags().StringArrayVar(&filterArgs, "filter", nil, "Filter: column:operation:value (repeatable)")

	cmd.MarkFlagsMutuallyExclusive("file", "name")
	cmd.MarkFlagsMutuallyExclusive("file", "filter")

	return cmd
}

func runViewUpdate(cmd *cobra.Command, opts *options.RootOptions, boardID, viewID, file, name string, filterArgs []string) error {
	client, err := opts.ClientFor(nil, options.AuthConfig)
	if err != nil {
		return err
	}

	ctx := cmd.Context()

	if file != "" {
		return updateViewFromFile(ctx, client, opts, boardID, viewID, file)
	}

	if !command.AnyChanged(cmd, "name", "filter") {
		return fmt.Errorf("--file, --name, or --filter is required")
	}

	current, err := getView(ctx, client, boardID, viewID)
	if err != nil {
		return err
	}

	body := api.UpdateBoardViewJSONRequestBody{
		Name:    name,
		Filters: []api.BoardViewFilter{},
	}

	if !cmd.Flags().Changed("name") {
		body.Name = deref.String(current.Name)
	}

	if cmd.Flags().Changed("filter") {
		body.Filters, err = parseViewFilters(filterArgs)
		if err != nil {
			return err
		}
	} else if current.Filters != nil {
		body.Filters = *current.Filters
	}

	resp, err := client.UpdateBoardViewWithResponse(ctx, boardID, viewID, body)
	if err != nil {
		return fmt.Errorf("updating board view: %w", err)
	}

	updated, err := api.Decode(resp.StatusCode(), resp.Status(), resp.Body, resp.JSON200)
	if err != nil {
		return err
	}

	return writeViewDetail(opts, viewResponseToDetail(*updated))
}

func updateViewFromFile(ctx context.Context, client *api.ClientWithResponses, opts *options.RootOptions, boardID, viewID, file string) error {
	data, err := command.ReadDefinitionFile(opts.IOStreams, file)
	if err != nil {
		return err
	}

	resp, err := client.UpdateBoardViewWithBodyWithResponse(ctx, boardID, viewID, "application/json", bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("updating board view: %w", err)
	}

	updated, err := api.Decode(resp.StatusCode(), resp.Status(), resp.Body, resp.JSON200)
	if err != nil {
		return err
	}

	return writeViewDetail(opts, viewResponseToDetail(*updated))
}

func getView(ctx context.Context, client *api.ClientWithResponses, boardID, viewID string) (*api.BoardViewResponse, error) {
	resp, err := client.GetBoardViewWithResponse(ctx, boardID, viewID)
	if err != nil {
		return nil, fmt.Errorf("getting board view: %w", err)
	}

	return api.Decode(resp.StatusCode(), resp.Status(), resp.Body, resp.JSON200)
}
