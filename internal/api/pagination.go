package api

import (
	"fmt"
	"net/url"
)

// NextPageCursor extracts the page[after] cursor from a links.next URL,
// returning an empty string when there is no further page. Passing the cursor
// that produced this page as prev bounds the loop: a server that keeps
// returning the same cursor would otherwise page forever.
func NextPageCursor(links *PaginationLinks, prev string) (string, error) {
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

	cursor := parsed.Query().Get("page[after]")
	if cursor != "" && cursor == prev {
		return "", fmt.Errorf("next page link repeats cursor %q", cursor)
	}

	return cursor, nil
}
