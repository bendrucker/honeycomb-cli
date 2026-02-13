package key

import (
	"fmt"
	"io"
	"strconv"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
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

type keyItem struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	KeyType  string `json:"key_type"`
	Disabled bool   `json:"disabled"`
}

type keyDetail struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	KeyType  string `json:"key_type"`
	Disabled bool   `json:"disabled"`
	Secret   string `json:"secret,omitempty"`
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
	fields := []output.Field{
		{Label: "ID", Value: detail.ID},
		{Label: "Name", Value: detail.Name},
		{Label: "Key Type", Value: detail.KeyType},
		{Label: "Disabled", Value: strconv.FormatBool(detail.Disabled)},
	}
	if detail.Secret != "" {
		fields = append(fields, output.Field{Label: "Secret", Value: detail.Secret})
	}
	return opts.OutputWriter().WriteFields(detail, fields)
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
