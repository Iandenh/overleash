package server

import (
	"fmt"
	"github.com/sahilm/fuzzy"
	"overleash/overleash"
	"slices"
	"strings"
)

func fuzzyFeatureFlags(search string, o *overleash.OverleashContext) overleash.FeatureFlags {
	search = strings.TrimSpace(search)

	if search == "" {
		return o.FeatureFile().Features
	}

	results := fuzzy.FindFrom(search, o.FeatureFile().Features)

	flags := make(overleash.FeatureFlags, 0, len(results))

	for _, r := range results {
		flag := o.FeatureFile().Features[r.Index]
		name := ""

		for i := 0; i < len(r.Str); i++ {
			if slices.Contains(r.MatchedIndexes, i) {
				name += fmt.Sprintf("<strong>%s</strong>", string(r.Str[i]))
			} else {
				name += string(r.Str[i])
			}
		}

		flag.SearchTerm = name
		flags = append(flags, flag)
	}

	return flags
}
