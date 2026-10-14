package server

import (
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strings"

	"github.com/Iandenh/overleash/overleash"
	"github.com/teal-finance/fuzzy"
)

type featureList struct {
	flags      overleash.FeatureViews
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

	snapshot := o.ActiveFeatureEnvironment().Snapshot()
	nameSorted := q == "" && (sortField == "name-asc" || sortField == "name-desc")
	var allFlags overleash.FeatureViews
	if nameSorted {
		allFlags = snapshot.FeaturesByName(sortField == "name-desc")
	} else {
		allFlags = snapshot.Features()
	}
	flags := fuzzyFeatureFlags(q, allFlags)
	flags = filterFeaturesByOverrideStatus(flags, filterOption, o)

	if !nameSorted {
		sortFeatures(flags, sortField)
	}

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

	finalURL := o.Config.CleanBasePath() + "/"
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
		totalFlags: len(allFlags),
	}
}

func fuzzyFeatureFlags(search string, flags overleash.FeatureViews) overleash.FeatureViews {
	search = strings.TrimSpace(search)

	if search == "" {
		return flags
	}

	results := fuzzy.FindFrom(search, flags)

	matched := make(overleash.FeatureViews, 0, len(results))

	for _, r := range results {
		flag := flags[r.Index]

		var nameBuilder strings.Builder

		for i, char := range r.Str {
			if slices.Contains(r.MatchedIndexes, i) {
				nameBuilder.WriteString(fmt.Sprintf("<strong>%c</strong>", char))
			} else {
				nameBuilder.WriteRune(char)
			}
		}

		flag.SearchTerm = nameBuilder.String()
		matched = append(matched, flag)
	}

	return matched
}

func sortFeatures(flags overleash.FeatureViews, sortOption string) {
	switch sortOption {
	case "name-asc":
		slices.SortFunc(flags, func(a, b overleash.FeatureView) int {
			return strings.Compare(strings.ToLower(a.Name()), strings.ToLower(b.Name()))
		})
	case "name-desc":
		slices.SortFunc(flags, func(a, b overleash.FeatureView) int {
			return strings.Compare(strings.ToLower(b.Name()), strings.ToLower(a.Name()))
		})
	}
}

func filterFeaturesByOverrideStatus(flags overleash.FeatureViews, filterOption string, o *overleash.OverleashContext) overleash.FeatureViews {
	var filteredFlags overleash.FeatureViews
	if filterOption != "overridden" && filterOption != "not-overridden" {
		return flags
	}
	overrides := o.Overrides()

	switch filterOption {
	case "overridden":
		for _, flag := range flags {
			if _, exists := overrides[flag.Name()]; exists {
				filteredFlags = append(filteredFlags, flag)
			}
		}
	case "not-overridden":
		for _, flag := range flags {
			if _, exists := overrides[flag.Name()]; !exists {
				filteredFlags = append(filteredFlags, flag)
			}
		}
	default:
		filteredFlags = flags
	}

	return filteredFlags
}
