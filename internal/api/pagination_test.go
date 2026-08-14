package api

import (
	"strings"
	"testing"

	"github.com/oapi-codegen/nullable"
)

func TestNextPageCursor(t *testing.T) {
	for _, tc := range []struct {
		name    string
		links   *PaginationLinks
		prev    string
		want    string
		wantErr string
	}{
		{
			name:  "nil links",
			links: nil,
		},
		{
			name:  "unspecified next",
			links: &PaginationLinks{},
		},
		{
			name:  "null next",
			links: &PaginationLinks{Next: nullable.NewNullNullable[string]()},
		},
		{
			name:  "empty next",
			links: &PaginationLinks{Next: nullable.NewNullableWithValue("")},
		},
		{
			name:  "next without cursor",
			links: &PaginationLinks{Next: nullable.NewNullableWithValue("/1/signals")},
		},
		{
			name:  "next with cursor",
			links: &PaginationLinks{Next: nullable.NewNullableWithValue("/1/signals?page%5Bafter%5D=cursor-2")},
			want:  "cursor-2",
		},
		{
			name:  "cursor advances past previous",
			links: &PaginationLinks{Next: nullable.NewNullableWithValue("/1/signals?page%5Bafter%5D=cursor-2")},
			prev:  "cursor-1",
			want:  "cursor-2",
		},
		{
			name:    "cursor repeats previous",
			links:   &PaginationLinks{Next: nullable.NewNullableWithValue("/1/signals?page%5Bafter%5D=cursor-1")},
			prev:    "cursor-1",
			wantErr: "repeats cursor",
		},
		{
			name:    "unparseable next",
			links:   &PaginationLinks{Next: nullable.NewNullableWithValue("://nope")},
			wantErr: "parsing next page link",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := NextPageCursor(tc.links, tc.prev)

			if tc.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q", tc.wantErr)
				}
				if !strings.Contains(err.Error(), tc.wantErr) {
					t.Errorf("error = %q, want it to contain %q", err.Error(), tc.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatal(err)
			}
			if got != tc.want {
				t.Errorf("cursor = %q, want %q", got, tc.want)
			}
		})
	}
}
