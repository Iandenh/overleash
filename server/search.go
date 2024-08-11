package server

import (
	"fmt"
	"github.com/Iandenh/overleash/overleash"
	"github.com/teal-finance/fuzzy"
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

		var nameBuilder strings.Builder

		for i, char := range r.Str {
			if slices.Contains(r.MatchedIndexes, i) {
				nameBuilder.WriteString(fmt.Sprintf("<strong>%c</strong>", char))
			} else {
				nameBuilder.WriteRune(char)
			}
		}

		flag.SearchTerm = nameBuilder.String()
		flags = append(flags, flag)
	}

	return flags
}
