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
