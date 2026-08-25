package server

import (
	"testing"
	"time"
)

func TestParseRequest(t *testing.T) {
	tests := []struct {
		input   string
		command string
		key     string
		value   string
		ttl     time.Duration
		wantErr bool
	}{
		{
			input:   "SET name Atithi",
			command: "SET",
			key:     "name",
			value:   "Atithi",
		},
		{
			input:   "SET name Atithi 10",
			command: "SET",
			key:     "name",
			value:   "Atithi",
			ttl:     10 * time.Second,
		},
		{
			input:   "SET name Atithi invalid",
			wantErr: true,
		},
		{
			input:   "SET name Atithi -5",
			wantErr: true,
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
			t.Errorf("unexpected error for %q: %v", tt.input, err)
		}

		if req.Command != tt.command ||
			req.Key != tt.key ||
			req.Value != tt.value ||
			req.TTL != tt.ttl {
			t.Errorf("unexpected request for %q: %+v", tt.input, req)
		}
	}
}
