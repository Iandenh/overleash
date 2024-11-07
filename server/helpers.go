package server

import (
	"fmt"
	"github.com/Iandenh/overleash/internal/version"
	unleash "github.com/Unleash/unleash-client-go/v4/api"
	"net/url"
)

func constraintsOfStrategy(strategy unleash.Strategy, segments map[int][]unleash.Constraint) []unleash.Constraint {
	if len(strategy.Segments) == 0 {
		return strategy.Constraints
	}

	var constraints []unleash.Constraint

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
