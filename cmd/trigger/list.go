package trigger

import (
	"context"
	"fmt"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/config"
	"github.com/bendrucker/honeycomb-cli/internal/output"
	"github.com/spf13/cobra"
)

var triggerListTable = output.TableFromTags[triggerItem]()

func NewListCmd(opts *options.RootOptions, dataset *string) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List triggers",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runList(cmd.Context(), opts, *dataset)
		},
	}
}

func runList(ctx context.Context, opts *options.RootOptions, dataset string) error {
	client, err := opts.Client(config.KeyConfig)
	if err != nil {
		return err
	}

	resp, err := client.ListTriggersWithResponse(ctx, dataset)
	if err != nil {
		return fmt.Errorf("listing triggers: %w", err)
	}

	triggers, err := api.Decode(resp.StatusCode(), resp.Status(), resp.Body, resp.JSON200)
	if err != nil {
		return err
	}

	items := make([]triggerItem, len(*triggers))
	for i, t := range *triggers {
		items[i] = toItem(t)
	}

	return opts.OutputWriterList().WriteList(items, triggerListTable, "No triggers found.")
}
