package key

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"text/tabwriter"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/config"
	"github.com/bendrucker/honeycomb-cli/internal/output"
	"github.com/spf13/cobra"
)

func NewCmd(opts *options.RootOptions) *cobra.Command {
	var team string

	cmd := &cobra.Command{
		Use:     "key",
		Short:   "Manage API keys",
		Aliases: []string{"keys"},
	}

	cmd.PersistentFlags().StringVar(&team, "team", "", "Team slug (required)")
	_ = cmd.MarkPersistentFlagRequired("team")

	cmd.AddCommand(NewListCmd(opts, &team))
	cmd.AddCommand(NewGetCmd(opts, &team))
	cmd.AddCommand(NewCreateCmd(opts, &team))
	cmd.AddCommand(NewUpdateCmd(opts, &team))
	cmd.AddCommand(NewDeleteCmd(opts, &team))

	return cmd
}

func keyEditor(key string) api.RequestEditorFn {
	return func(_ context.Context, req *http.Request) error {
		config.ApplyAuth(req, config.KeyManagement, key)
		return nil
	}
}

type keyItem struct {
	ID       string `json:"id"                    yaml:"id"`
	Name     string `json:"name"                  yaml:"name"`
	KeyType  string `json:"key_type"              yaml:"key_type"`
	Disabled bool   `json:"disabled"              yaml:"disabled"`
}

type keyDetail struct {
	ID       string `json:"id"                     yaml:"id"`
	Name     string `json:"name"                   yaml:"name"`
	KeyType  string `json:"key_type"               yaml:"key_type"`
	Disabled bool   `json:"disabled"               yaml:"disabled"`
	Secret   string `json:"secret,omitempty"        yaml:"secret,omitempty"`
}

var keyListTable = output.TableDef{
	Columns: []output.Column{
		{Header: "ID", Value: func(v any) string { return v.(keyItem).ID }},
		{Header: "Name", Value: func(v any) string { return v.(keyItem).Name }},
		{Header: "Key Type", Value: func(v any) string { return v.(keyItem).KeyType }},
		{Header: "Disabled", Value: func(v any) string { return fmt.Sprintf("%t", v.(keyItem).Disabled) }},
	},
}

func writeKeyDetail(opts *options.RootOptions, detail keyDetail) error {
	return opts.OutputWriter().WriteValue(detail, func(w io.Writer) error {
		tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
		_, _ = fmt.Fprintf(tw, "ID:\t%s\n", detail.ID)
		_, _ = fmt.Fprintf(tw, "Name:\t%s\n", detail.Name)
		_, _ = fmt.Fprintf(tw, "Key Type:\t%s\n", detail.KeyType)
		_, _ = fmt.Fprintf(tw, "Disabled:\t%t\n", detail.Disabled)
		if detail.Secret != "" {
			_, _ = fmt.Fprintf(tw, "Secret:\t%s\n", detail.Secret)
		}
		return tw.Flush()
	})
}

func objectToItem(obj api.ApiKeyObject) keyItem {
	item := keyItem{}
	if obj.Id != nil {
		item.ID = *obj.Id
	}
	if obj.Attributes != nil {
		fillFromAttributes(&item.Name, &item.KeyType, &item.Disabled, *obj.Attributes)
	}
	return item
}

func objectToDetail(obj api.ApiKeyObject) keyDetail {
	detail := keyDetail{}
	if obj.Id != nil {
		detail.ID = *obj.Id
	}
	if obj.Attributes != nil {
		fillFromAttributes(&detail.Name, &detail.KeyType, &detail.Disabled, *obj.Attributes)
	}
	return detail
}

func fillFromAttributes(name *string, keyType *string, disabled *bool, attrs api.ApiKeyAttributes) {
	if ingest, err := attrs.AsIngestKeyAttributes(); err == nil {
		*name = ingest.Name
		*keyType = string(ingest.KeyType)
		if ingest.Disabled != nil {
			*disabled = *ingest.Disabled
		}
		return
	}
	if cfg, err := attrs.AsConfigurationKeyAttributes(); err == nil {
		*name = cfg.Name
		*keyType = string(cfg.KeyType)
		if cfg.Disabled != nil {
			*disabled = *cfg.Disabled
		}
		return
	}
}

func readBodyFile(ios *options.RootOptions, file string) ([]byte, error) {
	var r io.Reader
	if file == "-" {
		r = ios.IOStreams.In
	} else {
		f, err := openFile(file)
		if err != nil {
			return nil, err
		}
		defer func() { _ = f.Close() }()
		r = f
	}
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}
	return data, nil
}
