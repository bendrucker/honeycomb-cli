package key

import (
	"context"
	"fmt"
	"net/url"
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

		cursor, err := nextEnvironmentPageCursor(list.Links)
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

// nextEnvironmentPageCursor extracts the page[after] cursor from a links.next
// URL. It returns an empty string when there is no further page.
func nextEnvironmentPageCursor(links *api.PaginationLinks) (string, error) {
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
