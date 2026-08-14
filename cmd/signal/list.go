package signal

import (
	"context"
	"fmt"
	"net/url"

	"github.com/bendrucker/honeycomb-cli/cmd/command"
	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/output"
	"github.com/spf13/cobra"
)

var signalListTable = output.TableFromTags[signalItem]()

func NewListCmd(opts *options.RootOptions) *cobra.Command {
	var (
		service        string
		dataset        string
		measuredSignal string
		status         string
		anomalous      bool
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List signals",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := command.ValidateEnum("measured-signal", measuredSignal, measuredSignals); err != nil {
				return err
			}
			if err := command.ValidateEnum("status", status, statuses); err != nil {
				return err
			}

			params := &api.ListSignalsParams{}
			if service != "" {
				params.ServiceName = &service
			}
			if dataset != "" {
				params.DatasetSlug = &dataset
			}
			if measuredSignal != "" {
				m := api.AnomalySignal(measuredSignal)
				params.MeasuredSignal = &m
			}
			if status != "" {
				s := api.ListSignalsParamsStatus(status)
				params.Status = &s
			}
			if cmd.Flags().Changed("anomalous") {
				params.CurrentlyAnomalous = &anomalous
			}

			return runList(cmd.Context(), opts, params)
		},
	}

	cmd.Flags().StringVar(&service, "service", "", "Filter by service name")
	cmd.Flags().StringVar(&dataset, "dataset", "", "Filter by dataset slug")
	cmd.Flags().StringVar(&measuredSignal, "measured-signal", "", "Filter by measured signal: "+enumUsage(measuredSignals))
	cmd.Flags().StringVar(&status, "status", "", "Filter by status: "+enumUsage(statuses))
	cmd.Flags().BoolVar(&anomalous, "anomalous", false, "Only list signals that are currently anomalous")

	return cmd
}

func runList(ctx context.Context, opts *options.RootOptions, params *api.ListSignalsParams) error {
	client, err := opts.ClientFor(nil, options.AuthConfig)
	if err != nil {
		return err
	}

	var items []signalItem
	for {
		resp, err := client.ListSignalsWithResponse(ctx, params)
		if err != nil {
			return fmt.Errorf("listing signals: %w", err)
		}

		page, err := api.Decode(resp.StatusCode(), resp.Status(), resp.Body, resp.JSON200)
		if err != nil {
			return err
		}

		for _, s := range page.Signals {
			items = append(items, toItem(s))
		}

		cursor, err := nextPageCursor(page.Links)
		if err != nil {
			return err
		}
		if cursor == "" {
			break
		}
		params.PageAfter = &cursor
	}

	return opts.OutputWriterList().WriteList(items, signalListTable, "No signals found.")
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
