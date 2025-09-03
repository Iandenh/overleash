package overleash

import (
	"testing"
)

func Test_fromString(t *testing.T) {
	t.Run("Should create from String", func(t *testing.T) {
		token := "*:development.key"

		got, ok := fromString(token)

		if !ok {
			t.Errorf("fromString() didn't correctly finisch")
		}

		if got.Token != token {
			t.Errorf("fromString() token = %v, want %v", got.Token, token)
		}

		if got.Environment != "development" {
			t.Errorf("fromString() Environment = %v, want %v", got.Environment, "development")
		}
	})

	t.Run("Should create from String with extra spaces", func(t *testing.T) {
		token := "        *:development.key           "

		got, ok := fromString(token)

		if !ok {
			t.Errorf("fromString() didn't correctly finisch")
		}

		if got.Token != "*:development.key" {
			t.Errorf("fromString() token = %v, want %v", got.Token, token)
		}

		if got.Environment != "development" {
			t.Errorf("fromString() Environment = %v, want %v", got.Environment, "development")
		}
	})
}

func TestExtractEnvironment(t *testing.T) {
	tests := []struct {
		name      string
		token     string
		want      string
		expectErr bool
	}{
		{
			name:      "valid token with development env",
			token:     "default:development.unleash-insecure-api-token",
			want:      "development",
			expectErr: false,
		},
		{
			name:      "valid token with production env",
			token:     "proj:production.abcdef123456",
			want:      "production",
			expectErr: false,
		},
		{
			name:      "missing colon",
			token:     "invalidtokenwithoutcolon",
			want:      "",
			expectErr: true,
		},
		{
			name:      "missing dot after environment",
			token:     "proj:staging",
			want:      "staging",
			expectErr: false,
		},
		{
			name:      "empty token",
			token:     "",
			want:      "",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ExtractEnvironment(tt.token)
			if (err != nil) != tt.expectErr {
				t.Errorf("ExtractEnvironment() error = %v, expectErr %v for test %s", err, tt.expectErr, tt.name)
				return
			}
			if got != tt.want {
				t.Errorf("ExtractEnvironment() = %v, want %v for test %s", got, tt.want, tt.name)
			}
		})
	}
}
