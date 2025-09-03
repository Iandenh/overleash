package overleash

import (
	"fmt"
	"strings"
)

type TokenType string

const (
	Frontend TokenType = "frontend"
	Client   TokenType = "overleashClient"
	Admin    TokenType = "admin"
	Unknown  TokenType = "unknown"
)

type EdgeToken struct {
	Token       string    `json:"token"`
	TokenType   TokenType `json:"type"`
	Environment string    `json:"environment"`
	Projects    []string  `json:"projects"`
}

func fromString(token string) (*EdgeToken, bool) {
	token = strings.TrimSpace(token)

	if strings.Contains(token, ":") && strings.Contains(token, ".") {
		tokenParts := strings.SplitN(token, ":", 2)

		project := tokenParts[0]

		tokenParts = strings.SplitN(tokenParts[1], ".", 2)

		return &EdgeToken{
			Token:       token,
			TokenType:   Unknown,
			Environment: tokenParts[0],
			Projects:    []string{project},
		}, true
	}

	return nil, false
}

func ExtractEnvironment(token string) (string, error) {
	// Split on ":"
	parts := strings.SplitN(token, ":", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid token format: missing ':'")
	}

	// The second part looks like "development.unleash-insecure-api-token"
	subParts := strings.SplitN(parts[1], ".", 2)
	if len(subParts) < 1 {
		return "", fmt.Errorf("invalid token format: missing '.'")
	}

	return subParts[0], nil
}
