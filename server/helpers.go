package server

import (
	"fmt"
	"html"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/Iandenh/overleash/internal/version"
	"github.com/Iandenh/overleash/overleash"
)

func constraintsOfStrategy(strategy overleash.Strategy, segments map[int][]overleash.Constraint) []overleash.Constraint {
	if len(strategy.Segments) == 0 {
		return strategy.Constraints
	}

	var constraints []overleash.Constraint

	copy(constraints, strategy.Constraints)

	for _, segmentId := range strategy.Segments {
		constraints = append(segments[segmentId], strategy.Constraints...)
	}

	return constraints
}

func getVersion() string {
	v := version.Version

	if v == "DEV" || v == "" {
		return "Development"
	}

	return fmt.Sprintf("v%s", v)
}

type shortcut struct {
	character   string
	description string
	alt         bool
}

func getShortcuts() []shortcut {
	return []shortcut{
		{
			character:   "↓",
			description: "Move down the flag selection",
			alt:         false,
		},
		{
			character:   "↑",
			description: "Move up the flag selection",
			alt:         false,
		},
		{
			character:   "e",
			description: "Enable selected flag",
			alt:         false,
		},
		{
			character:   "d",
			description: "Disable selected flag",
			alt:         false,
		},
		{
			character:   "q",
			description: "Remove selected flag",
			alt:         false,
		},
		{
			character:   "i",
			description: "Toggle constraints info on selected flag",
			alt:         false,
		},
		{
			character:   "/",
			description: "Focus search input",
			alt:         false,
		},
		{
			character:   "h",
			description: "Show this help dialog",
			alt:         true,
		},
		{
			character:   "r",
			description: "Refresh all flags",
			alt:         true,
		},
		{
			character:   "p",
			description: "Pause all overrides",
			alt:         true,
		},
		{
			character:   "t",
			description: "Toggle Dark/Light mode",
			alt:         true,
		},
		{
			character:   "↓",
			description: "Move down the remote selection",
			alt:         true,
		},
		{
			character:   "↑",
			description: "Move up the remote selection",
			alt:         true,
		},
	}
}

func (list *featureList) generateUrl(key, value string) string {
	urlValues := url.Values{}

	urlValues.Set("q", list.searchTerm)
	urlValues.Set("sort", list.sort)
	urlValues.Set("filter", list.filter)

	urlValues.Set(key, value)

	return "/?" + urlValues.Encode()
}

func (list *featureList) isSelected(key, value string) bool {
	urlValues := url.Values{}

	urlValues.Set("q", list.searchTerm)
	urlValues.Set("sort", list.sort)
	urlValues.Set("filter", list.filter)

	return urlValues.Get(key) == value
}

func renderJiraLink(text string) string {
	text = html.EscapeString(strings.TrimSpace(text))

	re := regexp.MustCompile(`https://[a-zA-Z0-9\-]+\.atlassian\.net/browse/([A-Z]+-\d+)`)

	return re.ReplaceAllStringFunc(text, func(match string) string {
		subMatch := re.FindStringSubmatch(match)
		if len(subMatch) < 2 {
			return match
		}
		ticket := subMatch[1]
		return fmt.Sprintf(`<a target="_black" href="%s">%s</a>`, match, ticket)
	})
}

func (c *Config) featureEnvironmentFromRequest(r *http.Request) *overleash.FeatureEnvironment {
	if c.envFromToken == false {
		return c.Overleash.ActiveFeatureEnvironment()
	}

	token := r.Header.Get("Authorization")

	if token == "" {
		return c.Overleash.ActiveFeatureEnvironment()
	}

	envName, err := overleash.ExtractEnvironment(token)

	if err != nil {
		return c.Overleash.ActiveFeatureEnvironment()
	}

	for _, f := range c.Overleash.FeatureEnvironments() {
		if f.Environment() == envName {
			return f
		}
	}

	return c.Overleash.ActiveFeatureEnvironment()
}
