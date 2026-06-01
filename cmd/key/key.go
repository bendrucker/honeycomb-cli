package key

import (
	"strings"

	"github.com/bendrucker/honeycomb-cli/cmd/command"
	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/deref"
	"github.com/bendrucker/honeycomb-cli/internal/output"
	"github.com/spf13/cobra"
)

func NewCmd(opts *options.RootOptions) *cobra.Command {
	var team string

	cmd := &cobra.Command{
		Use:     "key",
		Short:   "Manage API keys",
		Aliases: []string{"keys"},
		Example: `  # List API keys
  honeycomb key list

  # Get an API key by ID
  honeycomb key get abc123`,
	}

	cmd.PersistentFlags().StringVar(&team, "team", "", "Team slug")

	cmd.AddCommand(NewListCmd(opts, &team))
	cmd.AddCommand(NewGetCmd(opts, &team))
	cmd.AddCommand(NewCreateCmd(opts, &team))
	cmd.AddCommand(NewUpdateCmd(opts, &team))
	cmd.AddCommand(NewDeleteCmd(opts, &team))

	return command.Group(cmd)
}

type keyItem struct {
	ID       string `json:"id" col:"ID"`
	Name     string `json:"name" col:"Name"`
	KeyType  string `json:"key_type" col:"Key Type"`
	Disabled bool   `json:"disabled" col:"Disabled"`
}

type keyDetail struct {
	ID          string   `json:"id" detail:"ID"`
	Name        string   `json:"name" detail:"Name"`
	KeyType     string   `json:"key_type" detail:"Key Type"`
	Disabled    bool     `json:"disabled" detail:"Disabled"`
	Environment string   `json:"environment,omitempty"`
	Permissions []string `json:"permissions,omitempty"`
	Secret      string   `json:"secret,omitempty"`
}

var keyListTable = output.TableFromTags[keyItem]()

func writeKeyDetail(opts *options.RootOptions, detail keyDetail) error {
	fields := output.FieldsFromTags(detail)
	if detail.Environment != "" {
		fields = append(fields, output.Field{Label: "Environment", Value: detail.Environment})
	}
	if len(detail.Permissions) > 0 {
		fields = append(fields, output.Field{Label: "Permissions", Value: strings.Join(detail.Permissions, ", ")})
	}
	if detail.Secret != "" {
		fields = append(fields, output.Field{Label: "Secret", Value: detail.Secret})
	}
	return opts.OutputWriter().WriteFields(detail, fields)
}

func objectToItem(obj api.ApiKeyObject) keyItem {
	item := keyItem{
		ID: deref.String(obj.Id),
	}
	if obj.Attributes != nil {
		fillFromAttributes(&item.Name, &item.KeyType, &item.Disabled, *obj.Attributes)
	}
	return item
}

func objectToDetail(obj api.ApiKeyObject) keyDetail {
	detail := keyDetail{
		ID: deref.String(obj.Id),
	}
	if obj.Attributes != nil {
		fillFromAttributes(&detail.Name, &detail.KeyType, &detail.Disabled, *obj.Attributes)
		if cfg, err := obj.Attributes.AsConfigurationKeyAttributes(); err == nil {
			detail.Permissions = grantedPermissions(cfg.Permissions)
		}
	}
	if obj.Relationships != nil {
		detail.Environment = obj.Relationships.Environment.Data.Id
	}
	return detail
}

func fillFromAttributes(name *string, keyType *string, disabled *bool, attrs api.ApiKeyAttributes) {
	if ingest, err := attrs.AsIngestKeyAttributes(); err == nil {
		*name = ingest.Name
		*keyType = string(ingest.KeyType)
		*disabled = deref.Bool(ingest.Disabled)
		return
	}
	if cfg, err := attrs.AsConfigurationKeyAttributes(); err == nil {
		*name = cfg.Name
		*keyType = string(cfg.KeyType)
		*disabled = deref.Bool(cfg.Disabled)
		return
	}
}
