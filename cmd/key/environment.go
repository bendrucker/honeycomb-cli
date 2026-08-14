package key

import (
	"context"
	"fmt"
	"strings"

	"github.com/bendrucker/honeycomb-cli/internal/api"
)

// environmentIDPrefix is the prefix Honeycomb uses for environment IDs.
const environmentIDPrefix = "hcaen_"

// resolveEnvironment turns a user-supplied --environment value into an
// environment ID. A value carrying the environment ID prefix is returned
// unchanged. Otherwise the value is treated as an environment name and looked
// up against the Management API for the team.
func resolveEnvironment(ctx context.Context, client *api.ClientWithResponses, team, value string) (string, error) {
	if strings.HasPrefix(value, environmentIDPrefix) {
		return value, nil
	}

	params := &api.ListEnvironmentsParams{}
	var cursor string
	for {
		resp, err := client.ListEnvironmentsWithResponse(ctx, api.TeamSlug(team), params)
		if err != nil {
			return "", fmt.Errorf("listing environments: %w", err)
		}

		list, err := api.Decode(resp.StatusCode(), resp.Status(), resp.Body, resp.ApplicationvndApiJSON200)
		if err != nil {
			return "", err
		}

		for _, e := range list.Data {
			if e.Attributes.Name == value {
				return e.Id, nil
			}
		}

		cursor, err = api.NextPageCursor(list.Links, cursor)
		if err != nil {
			return "", err
		}
		if cursor == "" {
			break
		}
		params.PageAfter = &cursor
	}

	return "", fmt.Errorf("no environment found with name %q", value)
}
