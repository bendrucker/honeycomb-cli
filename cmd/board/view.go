package board

import (
	"context"
	"fmt"
	"text/tabwriter"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/config"
	"github.com/bendrucker/honeycomb-cli/internal/output"
	"github.com/spf13/cobra"
)

func NewViewCmd(opts *options.RootOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "view <board-id>",
		Short: "View a board",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runBoardView(cmd.Context(), opts, args[0])
		},
	}
}

func runBoardView(ctx context.Context, opts *options.RootOptions, boardID string) error {
	key, err := opts.RequireKey(config.KeyConfig)
	if err != nil {
		return err
	}

	client, err := api.NewClientWithResponses(opts.ResolveAPIUrl())
	if err != nil {
		return fmt.Errorf("creating API client: %w", err)
	}

	resp, err := client.GetBoardWithResponse(ctx, boardID, keyEditor(key))
	if err != nil {
		return fmt.Errorf("getting board: %w", err)
	}

	if err := api.CheckResponse(resp.StatusCode(), resp.Body); err != nil {
		return err
	}

	if resp.JSON200 == nil {
		return fmt.Errorf("unexpected response: %s", resp.Status())
	}

	detail := boardToDetail(*resp.JSON200)

	format := opts.ResolveFormat()
	if format != "table" {
		return opts.OutputWriter().Write(detail, output.TableDef{})
	}

	tw := tabwriter.NewWriter(opts.IOStreams.Out, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintf(tw, "ID:\t%s\n", detail.ID)
	_, _ = fmt.Fprintf(tw, "Name:\t%s\n", detail.Name)
	_, _ = fmt.Fprintf(tw, "Description:\t%s\n", detail.Description)
	_, _ = fmt.Fprintf(tw, "Type:\t%s\n", detail.Type)
	_, _ = fmt.Fprintf(tw, "Column Layout:\t%s\n", detail.ColumnLayout)
	_, _ = fmt.Fprintf(tw, "URL:\t%s\n", detail.URL)

	if resp.JSON200.Panels != nil {
		_, _ = fmt.Fprintf(tw, "\nPanels:\n")
		for i, panel := range *resp.JSON200.Panels {
			panelType, err := panel.Discriminator()
			if err != nil {
				_, _ = fmt.Fprintf(tw, "  %d:\t(unknown type)\n", i+1)
				continue
			}
			switch panelType {
			case "query":
				qp, _ := panel.AsQueryPanel()
				_, _ = fmt.Fprintf(tw, "  %d:\tquery\tquery_id=%s\n", i+1, qp.QueryPanel.QueryId)
			case "slo":
				sp, _ := panel.AsSLOPanel()
				sloID := ""
				if sp.SloPanel.SloId != nil {
					sloID = *sp.SloPanel.SloId
				}
				_, _ = fmt.Fprintf(tw, "  %d:\tslo\tslo_id=%s\n", i+1, sloID)
			case "text":
				tp, _ := panel.AsTextPanel()
				_, _ = fmt.Fprintf(tw, "  %d:\ttext\t%s\n", i+1, truncate(tp.TextPanel.Content, 40))
			default:
				_, _ = fmt.Fprintf(tw, "  %d:\t%s\n", i+1, panelType)
			}
		}
	}

	return tw.Flush()
}
