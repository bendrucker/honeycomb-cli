package command

import "testing"

func TestValidateEnum(t *testing.T) {
	allowed := []string{"ingest", "configuration"}

	for _, tc := range []struct {
		name    string
		value   string
		wantErr string
	}{
		{
			name:  "empty passes",
			value: "",
		},
		{
			name:  "allowed value passes",
			value: "ingest",
		},
		{
			name:    "invalid value reports allowed set",
			value:   "bogus",
			wantErr: `invalid --key-type "bogus": must be one of ingest, configuration`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateEnum("key-type", tc.value, allowed)
			if tc.wantErr == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("expected error %q, got nil", tc.wantErr)
			}
			if err.Error() != tc.wantErr {
				t.Errorf("error = %q, want %q", err.Error(), tc.wantErr)
			}
		})
	}
}
