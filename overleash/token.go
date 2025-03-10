package overleash

import "strings"

type TokenType string

const (
	Frontend TokenType = "frontend"
	Client   TokenType = "overleashclient"
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
