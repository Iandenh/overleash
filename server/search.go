package server

import (
	"fmt"
	"github.com/Iandenh/overleash/overleash"
	"github.com/teal-finance/fuzzy"
	"net/http"
	"net/url"
	"slices"
	"sort"
	"strings"
)

type featureList struct {
	flags      overleash.FeatureFlags
	searchTerm string
	sort       string
	filter     string
	url        string
	totalFlags int
}

func searchTerm(r *http.Request) string {
	err := r.ParseForm()

	if err == nil {
		term := r.Form.Get("search")

		if term != "" {
			return term
		}
	}

	return strings.TrimSpace(r.URL.Query().Get("q"))
}

func search(r *http.Request, o *overleash.OverleashContext) featureList {
	query := r.URL.Query()

	sortField := strings.ToLower(query.Get("sort"))
	filterOption := strings.ToLower(r.URL.Query().Get("filter"))

	q := searchTerm(r)

	flags := fuzzyFeatureFlags(q, o)
	flags = filterFeaturesByOverrideStatus(flags, filterOption, o)

	sortFeatures(flags, sortField)

	urlValues := url.Values{}

	if q != "" {
		urlValues.Set("q", q)
	}

	if sortField != "" {
		urlValues.Set("sort", sortField)
	}

	if filterOption != "" {
		urlValues.Set("filter", filterOption)
	}

	finalURL := "/"
	encodedQuery := urlValues.Encode()
	if encodedQuery != "" {
		finalURL += "?" + encodedQuery
	}

	return featureList{
		flags:      flags,
		searchTerm: q,
		sort:       sortField,
		filter:     filterOption,
		url:        finalURL,
		totalFlags: o.FeatureFile().Features.Len(),
	}
}

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

func sortFeatures(flags overleash.FeatureFlags, sortOption string) {
	switch sortOption {
	case "name-asc":
		sort.Slice(flags, func(i, j int) bool {
			return strings.ToLower(flags[i].Name) < strings.ToLower(flags[j].Name)
		})
	case "name-desc":
		sort.Slice(flags, func(i, j int) bool {
			return strings.ToLower(flags[i].Name) > strings.ToLower(flags[j].Name)
		})
	default:
		sort.Slice(flags, func(i, j int) bool {
			return strings.ToLower(flags[i].Name) < strings.ToLower(flags[j].Name)
		})
	}
}

func filterFeaturesByOverrideStatus(flags overleash.FeatureFlags, filterOption string, o *overleash.OverleashContext) overleash.FeatureFlags {
	var filteredFlags overleash.FeatureFlags

	switch filterOption {
	case "overridden":
		for _, flag := range flags {
			if override, exists := o.Overrides()[flag.Name]; exists && override.Enabled {
				filteredFlags = append(filteredFlags, flag)
			}
		}
	case "not-overridden":
		for _, flag := range flags {
			if override, exists := o.Overrides()[flag.Name]; !exists || !override.Enabled {
				filteredFlags = append(filteredFlags, flag)
			}
		}
	default:
		filteredFlags = flags
	}

	return filteredFlags
}
