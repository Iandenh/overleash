package server

import (
	"fmt"
	"github.com/Iandenh/overleash/internal/version"
	unleash "github.com/Unleash/unleash-client-go/v4/api"
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
