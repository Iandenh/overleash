package server

import unleash "github.com/Unleash/unleash-client-go/v4/api"

func constraintsOfStrategy(strategy unleash.Strategy, segments map[int][]unleash.Constraint) []unleash.Constraint {
	if len(strategy.Segments) == 0 {
		return strategy.Constraints
	}

	var Constraints []unleash.Constraint

	copy(Constraints, strategy.Constraints)

	for _, segmentId := range strategy.Segments {
		Constraints = append(segments[segmentId], strategy.Constraints...)
	}

	return Constraints
}
