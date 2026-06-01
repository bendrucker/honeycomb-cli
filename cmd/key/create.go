package key

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"slices"

	"github.com/bendrucker/honeycomb-cli/cmd/command"
	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/deref"
	"github.com/bendrucker/honeycomb-cli/internal/prompt"
	"github.com/spf13/cobra"
)

var knownPermissions = []string{
	"create_datasets",
	"manage_boards",
	"manage_columns",
	"manage_markers",
	"manage_private_boards",
	"manage_recipients",
	"manage_slos",
	"manage_triggers",
	"read_service_maps",
	"run_queries",
	"send_events",
}

func NewCreateCmd(opts *options.RootOptions, team *string) *cobra.Command {
	var (
		file           string
		name           string
		keyType        string
		environment    string
		permissions    []string
		allPermissions bool
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create an API key",
		Example: `  # Create an ingest key in an environment
  honeycomb key create --name "ingest" --key-type ingest \
    --environment env-abc --all-permissions

  # Create a configuration key with specific permissions
  honeycomb key create --name "config" --key-type configuration \
    --environment env-abc --permission boards --permission triggers`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := opts.ClientFor(team, options.AuthManagement)
			if err != nil {
				return err
			}
			return runKeyCreate(cmd, opts, client, *team, file, name, keyType, environment, permissions, allPermissions)
		},
	}

	cmd.Flags().StringVarP(&file, "file", "f", "", "Path to JSON:API request body (- for stdin)")
	cmd.Flags().StringVar(&name, "name", "", "Key name")
	cmd.Flags().StringVar(&keyType, "key-type", "", "Key type (ingest or configuration)")
	cmd.Flags().StringVar(&environment, "environment", "", "Environment ID")
	cmd.Flags().StringSliceVar(&permissions, "permission", nil, "Permission to grant (repeatable)")
	cmd.Flags().BoolVar(&allPermissions, "all-permissions", false, "Grant all permissions")

	cmd.MarkFlagsMutuallyExclusive("file", "name")
	cmd.MarkFlagsMutuallyExclusive("file", "key-type")
	cmd.MarkFlagsMutuallyExclusive("file", "environment")
	cmd.MarkFlagsMutuallyExclusive("file", "permission")
	cmd.MarkFlagsMutuallyExclusive("file", "all-permissions")
	cmd.MarkFlagsMutuallyExclusive("permission", "all-permissions")

	return cmd
}

func runKeyCreate(cmd *cobra.Command, opts *options.RootOptions, client *api.ClientWithResponses, team, file, name, keyType, environment string, permissions []string, allPermissions bool) error {
	if file != "" {
		return runKeyCreateFromFile(cmd.Context(), opts, client, team, file)
	}
	return runKeyCreateFromFlags(cmd, opts, client, team, name, keyType, environment, permissions, allPermissions)
}

func runKeyCreateFromFile(ctx context.Context, opts *options.RootOptions, client *api.ClientWithResponses, team, file string) error {
	data, err := command.ReadDefinitionFile(opts.IOStreams, file)
	if err != nil {
		return err
	}

	resp, err := client.CreateApiKeyWithBodyWithResponse(ctx, api.TeamSlug(team), "application/vnd.api+json", bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("creating API key: %w", err)
	}

	return handleCreateResponse(opts, resp)
}

func runKeyCreateFromFlags(cmd *cobra.Command, opts *options.RootOptions, client *api.ClientWithResponses, team, name, keyType, environment string, permissions []string, allPermissions bool) error {
	var err error

	if name == "" {
		if !opts.IOStreams.CanPrompt() {
			return fmt.Errorf("--name is required in non-interactive mode")
		}
		name, err = prompt.Line(opts.IOStreams.Err, opts.IOStreams.In, "Key name: ")
		if err != nil {
			return err
		}
		if name == "" {
			return fmt.Errorf("key name is required")
		}
	}

	if keyType == "" {
		if !opts.IOStreams.CanPrompt() {
			return fmt.Errorf("--key-type is required in non-interactive mode")
		}
		keyType, err = prompt.Choice(opts.IOStreams.Err, opts.IOStreams.In, "Key type (ingest, configuration): ", []string{"ingest", "configuration"})
		if err != nil {
			return err
		}
	}

	if keyType != "ingest" && keyType != "configuration" {
		return fmt.Errorf("--key-type must be ingest or configuration")
	}

	if (len(permissions) > 0 || allPermissions) && keyType != "configuration" {
		return fmt.Errorf("--permission and --all-permissions are only valid with --key-type configuration")
	}

	if err := validatePermissions(permissions); err != nil {
		return err
	}

	if environment == "" {
		if !opts.IOStreams.CanPrompt() {
			return fmt.Errorf("--environment is required in non-interactive mode")
		}
		environment, err = prompt.Line(opts.IOStreams.Err, opts.IOStreams.In, "Environment ID: ")
		if err != nil {
			return err
		}
		if environment == "" {
			return fmt.Errorf("environment ID is required")
		}
	}

	body := api.ApiKeyCreateRequest{}
	body.Data.Type = api.ApiKeyCreateRequestDataTypeApiKeys
	body.Data.Relationships.Environment = api.EnvironmentRelationship{
		Data: struct {
			Id   string                              `json:"id"`
			Type api.EnvironmentRelationshipDataType `json:"type"`
		}{
			Id:   environment,
			Type: api.EnvironmentRelationshipDataTypeEnvironments,
		},
	}

	switch keyType {
	case "ingest":
		err = body.Data.Attributes.FromIngestKeyAttributes(api.IngestKeyAttributes{
			Name: name,
		})
	case "configuration":
		attrs := api.ConfigurationKeyAttributes{
			Name: name,
		}
		if allPermissions {
			permissions = knownPermissions
		}
		if len(permissions) > 0 {
			setPermissions(&attrs, permissions)
		}
		err = body.Data.Attributes.FromConfigurationKeyAttributes(attrs)
	}
	if err != nil {
		return fmt.Errorf("building request: %w", err)
	}

	resp, err := client.CreateApiKeyWithApplicationVndAPIPlusJSONBodyWithResponse(cmd.Context(), api.TeamSlug(team), body)
	if err != nil {
		return fmt.Errorf("creating API key: %w", err)
	}

	return handleCreateResponse(opts, resp)
}

func validatePermissions(permissions []string) error {
	known := make(map[string]bool, len(knownPermissions))
	for _, p := range knownPermissions {
		known[p] = true
	}
	for _, p := range permissions {
		if !known[p] {
			return fmt.Errorf("unknown permission %q; valid permissions: %v", p, knownPermissions)
		}
	}
	return nil
}

func setPermissions(attrs *api.ConfigurationKeyAttributes, permissions []string) {
	t := true
	attrs.Permissions = &struct {
		CreateDatasets      *bool `json:"create_datasets,omitempty"`
		ManageBoards        *bool `json:"manage_boards,omitempty"`
		ManageColumns       *bool `json:"manage_columns,omitempty"`
		ManageMarkers       *bool `json:"manage_markers,omitempty"`
		ManagePrivateBoards *bool `json:"manage_privateBoards,omitempty"`
		ManageRecipients    *bool `json:"manage_recipients,omitempty"`
		ManageSlos          *bool `json:"manage_slos,omitempty"`
		ManageTriggers      *bool `json:"manage_triggers,omitempty"`
		ReadServiceMaps     *bool `json:"read_service_maps,omitempty"`
		RunQueries          *bool `json:"run_queries,omitempty"`
		SendEvents          *bool `json:"send_events,omitempty"`
		VisibleTeamMembers  *bool `json:"visible_team_members,omitempty"`
	}{}
	for _, p := range permissions {
		switch p {
		case "create_datasets":
			attrs.Permissions.CreateDatasets = &t
		case "manage_boards":
			attrs.Permissions.ManageBoards = &t
		case "manage_columns":
			attrs.Permissions.ManageColumns = &t
		case "manage_markers":
			attrs.Permissions.ManageMarkers = &t
		case "manage_private_boards":
			attrs.Permissions.ManagePrivateBoards = &t
		case "manage_recipients":
			attrs.Permissions.ManageRecipients = &t
		case "manage_slos":
			attrs.Permissions.ManageSlos = &t
		case "manage_triggers":
			attrs.Permissions.ManageTriggers = &t
		case "read_service_maps":
			attrs.Permissions.ReadServiceMaps = &t
		case "run_queries":
			attrs.Permissions.RunQueries = &t
		case "send_events":
			attrs.Permissions.SendEvents = &t
		}
	}
}

func grantedPermissions(perms *struct {
	CreateDatasets      *bool `json:"create_datasets,omitempty"`
	ManageBoards        *bool `json:"manage_boards,omitempty"`
	ManageColumns       *bool `json:"manage_columns,omitempty"`
	ManageMarkers       *bool `json:"manage_markers,omitempty"`
	ManagePrivateBoards *bool `json:"manage_privateBoards,omitempty"`
	ManageRecipients    *bool `json:"manage_recipients,omitempty"`
	ManageSlos          *bool `json:"manage_slos,omitempty"`
	ManageTriggers      *bool `json:"manage_triggers,omitempty"`
	ReadServiceMaps     *bool `json:"read_service_maps,omitempty"`
	RunQueries          *bool `json:"run_queries,omitempty"`
	SendEvents          *bool `json:"send_events,omitempty"`
	VisibleTeamMembers  *bool `json:"visible_team_members,omitempty"`
}) []string {
	if perms == nil {
		return nil
	}
	var granted []string
	for name, flag := range map[string]*bool{
		"create_datasets":       perms.CreateDatasets,
		"manage_boards":         perms.ManageBoards,
		"manage_columns":        perms.ManageColumns,
		"manage_markers":        perms.ManageMarkers,
		"manage_private_boards": perms.ManagePrivateBoards,
		"manage_recipients":     perms.ManageRecipients,
		"manage_slos":           perms.ManageSlos,
		"manage_triggers":       perms.ManageTriggers,
		"read_service_maps":     perms.ReadServiceMaps,
		"run_queries":           perms.RunQueries,
		"send_events":           perms.SendEvents,
	} {
		if deref.Bool(flag) {
			granted = append(granted, name)
		}
	}
	slices.Sort(granted)
	return granted
}

func handleCreateResponse(opts *options.RootOptions, resp *api.CreateApiKeyResp) error {
	created, err := api.Decode(resp.StatusCode(), resp.Status(), resp.Body, resp.ApplicationvndApiJSON201)
	if err != nil {
		return err
	}

	detail := createResponseToDetail(created)

	if detail.Secret != "" {
		_, _ = fmt.Fprintf(opts.IOStreams.Err, "Save this key now — it cannot be retrieved again.\n")
	}

	return writeKeyDetail(opts, detail)
}

func createResponseToDetail(resp *api.ApiKeyCreateResponse) keyDetail {
	detail := keyDetail{
		ID:          deref.String(resp.Data.Id),
		Secret:      deref.String(resp.Data.Attributes.Secret),
		Environment: resp.Data.Relationships.Environment.Data.Id,
	}

	if ingest, err := resp.Data.Attributes.AsIngestKeyAttributes(); err == nil {
		detail.Name = ingest.Name
		detail.KeyType = string(ingest.KeyType)
		detail.Disabled = deref.Bool(ingest.Disabled)
	} else if cfg, err := resp.Data.Attributes.AsConfigurationKeyAttributes(); err == nil {
		detail.Name = cfg.Name
		detail.KeyType = string(cfg.KeyType)
		detail.Disabled = deref.Bool(cfg.Disabled)
		detail.Permissions = grantedPermissions(cfg.Permissions)
	} else {
		var raw struct {
			Name     string `json:"name"`
			KeyType  string `json:"key_type"`
			Disabled bool   `json:"disabled"`
		}
		rawBytes, _ := json.Marshal(resp.Data.Attributes)
		_ = json.Unmarshal(rawBytes, &raw)
		detail.Name = raw.Name
		detail.KeyType = raw.KeyType
		detail.Disabled = raw.Disabled
	}

	return detail
}
