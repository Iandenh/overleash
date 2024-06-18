package server

import (
	"fmt"
	"github.com/sahilm/fuzzy"
	"overleash/overleash"
	"strings"
)

func fuzzyFeatureFlags(search string, o *overleash.OverleashContext) overleash.FeatureFlags {
	search = strings.TrimSpace(search)

	if search == "" {
		return o.FeatureFile().Features
	}

	results := fuzzy.FindFrom(search, o.FeatureFile().Features)

	flags := overleash.FeatureFlags{}

	for _, r := range results {
		flag := o.FeatureFile().Features[r.Index]
		name := ""

		for i := 0; i < len(r.Str); i++ {
			if contains(i, r.MatchedIndexes) {
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

func contains(needle int, haystack []int) bool {
	for _, i := range haystack {
		if needle == i {
			return true
		}
	}
	return false
}
