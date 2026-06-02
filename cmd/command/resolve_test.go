package command

import (
	"errors"
	"strings"
	"testing"

	"github.com/bendrucker/honeycomb-cli/internal/iostreams"
)

func TestResolve(t *testing.T) {
	nonInteractive := errors.New("--name is required in non-interactive mode")
	empty := errors.New("name is required")

	for _, tc := range []struct {
		name        string
		value       string
		promptable  bool
		input       string
		field       Field
		want        string
		wantErr     error
		wantErrText string
		wantPrompt  string
		wantStream  Stream
	}{
		{
			name:  "value present skips prompt",
			value: "set",
			field: Field{
				Prompt:            "Name: ",
				Required:          true,
				NonInteractiveErr: nonInteractive,
				EmptyErr:          empty,
			},
			want: "set",
		},
		{
			name:       "required empty non-interactive returns error",
			value:      "",
			promptable: false,
			field: Field{
				Prompt:            "Name: ",
				Required:          true,
				NonInteractiveErr: nonInteractive,
				EmptyErr:          empty,
			},
			wantErr: nonInteractive,
		},
		{
			name:       "required empty prompts and returns answer",
			value:      "",
			promptable: true,
			input:      "prompted\n",
			field: Field{
				Prompt:            "Name: ",
				Required:          true,
				NonInteractiveErr: nonInteractive,
				EmptyErr:          empty,
			},
			want:       "prompted",
			wantPrompt: "Name: ",
			wantStream: StreamErr,
		},
		{
			name:       "required prompted empty returns empty error",
			value:      "",
			promptable: true,
			input:      "\n",
			field: Field{
				Prompt:            "Name: ",
				Required:          true,
				NonInteractiveErr: nonInteractive,
				EmptyErr:          empty,
			},
			wantErr: empty,
		},
		{
			name:       "optional empty non-interactive returns empty no error",
			value:      "",
			promptable: false,
			field: Field{
				Prompt: "Description (optional): ",
			},
			want: "",
		},
		{
			name:       "optional empty prompts and accepts empty",
			value:      "",
			promptable: true,
			input:      "\n",
			field: Field{
				Prompt: "Description (optional): ",
			},
			want:       "",
			wantPrompt: "Description (optional): ",
			wantStream: StreamErr,
		},
		{
			name:       "choice path resolves matching option",
			value:      "",
			promptable: true,
			input:      "slack\n",
			field: Field{
				Prompt:            "Type: ",
				Required:          true,
				Choices:           []string{"email", "slack"},
				NonInteractiveErr: nonInteractive,
				EmptyErr:          empty,
			},
			want:       "slack",
			wantPrompt: "Type: ",
			wantStream: StreamErr,
		},
		{
			name:       "prompt read error propagates",
			value:      "",
			promptable: true,
			input:      "",
			field: Field{
				Prompt:            "Name: ",
				Required:          true,
				NonInteractiveErr: nonInteractive,
				EmptyErr:          empty,
			},
			wantErrText: "unexpected end of input",
		},
		{
			name:       "stream out writes prompt to stdout",
			value:      "",
			promptable: true,
			input:      "prompted\n",
			field: Field{
				Prompt: "Dataset name: ",
				Stream: StreamOut,
			},
			want:       "prompted",
			wantPrompt: "Dataset name: ",
			wantStream: StreamOut,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var ts *iostreams.TestStreams
			if tc.promptable {
				ts = iostreams.TestPromptable(t)
			} else {
				ts = iostreams.Test(t)
			}
			if tc.input != "" {
				ts.InBuf.WriteString(tc.input)
			}

			got, err := Resolve(ts.IOStreams, tc.value, tc.field)

			if tc.wantErrText != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantErrText) {
					t.Fatalf("error = %v, want containing %q", err, tc.wantErrText)
				}
				return
			}
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("error = %v, want %v", err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Errorf("value = %q, want %q", got, tc.want)
			}

			if tc.wantPrompt != "" {
				stream := ts.ErrBuf.String()
				other := ts.OutBuf.String()
				if tc.wantStream == StreamOut {
					stream, other = other, stream
				}
				if stream != tc.wantPrompt {
					t.Errorf("prompt stream = %q, want %q", stream, tc.wantPrompt)
				}
				if other != "" {
					t.Errorf("unexpected output on other stream = %q", other)
				}
			}
		})
	}
}
