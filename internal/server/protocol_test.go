package server

import "testing"

func TestParseRequest(t *testing.T) {
	tests := []struct {
		input   string
		command string
		key     string
		value   string
		wantErr bool
	}{
		{
			input:   "SET name Atithi",
			command: "SET",
			key:     "name",
			value:   "Atithi",
		},
		{
			input:   "GET name",
			command: "GET",
			key:     "name",
		},
		{
			input:   "DELETE name",
			command: "DELETE",
			key:     "name",
		},
		{
			input:   "GET",
			wantErr: true,
		},
		{
			input:   "UNKNOWN name",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		req, err := ParseRequest(tt.input)

		if tt.wantErr {
			if err == nil {
				t.Errorf("expected error for %q", tt.input)
			}
			continue
		}

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if req.Command != tt.command ||
			req.Key != tt.key ||
			req.Value != tt.value {
			t.Errorf("unexpected request: %+v", req)
		}
	}
}
