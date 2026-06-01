package environment

import (
	"context"
	"fmt"
	"net/url"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/config"
	"github.com/bendrucker/honeycomb-cli/internal/output"
	"github.com/spf13/cobra"
)

var environmentListTable = output.TableFromTags[environmentItem]()

func NewListCmd(opts *options.RootOptions, team *string) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List environments",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := opts.RequireTeam(team); err != nil {
				return err
			}
			return runEnvironmentList(cmd.Context(), opts, *team)
		},
	}
}

func runEnvironmentList(ctx context.Context, opts *options.RootOptions, team string) error {
	client, err := opts.Client(config.KeyManagement)
	if err != nil {
		return err
	}

	var items []environmentItem
	params := &api.ListEnvironmentsParams{}
	for {
		resp, err := client.ListEnvironmentsWithResponse(ctx, team, params)
		if err != nil {
			return fmt.Errorf("listing environments: %w", err)
		}

		list, err := api.Decode(resp.StatusCode(), resp.Status(), resp.Body, resp.ApplicationvndApiJSON200)
		if err != nil {
			return err
		}

		for _, e := range list.Data {
			items = append(items, envToItem(e))
		}

		cursor, err := nextPageCursor(list.Links)
		if err != nil {
			return err
		}
		if cursor == "" {
			break
		}
		params.PageAfter = &cursor
	}

	return opts.OutputWriterList().WriteList(items, environmentListTable, "No environments found.")
}

// nextPageCursor extracts the page[after] cursor from a links.next URL.
// It returns an empty string when there is no further page.
func nextPageCursor(links *api.PaginationLinks) (string, error) {
	if links == nil || !links.Next.IsSpecified() || links.Next.IsNull() {
		return "", nil
	}

	next, err := links.Next.Get()
	if err != nil {
		return "", fmt.Errorf("reading next page link: %w", err)
	}
	if next == "" {
		return "", nil
	}

	parsed, err := url.Parse(next)
	if err != nil {
		return "", fmt.Errorf("parsing next page link: %w", err)
	}

	return parsed.Query().Get("page[after]"), nil
}
