package overleash

import (
	"errors"
	unleash "github.com/Unleash/unleash-client-go/v4/api"
)

type FeatureFile struct {
	unleash.FeatureResponse

	Features FeatureFlags `json:"features"`
}

type FeatureFlags []Feature

type Feature struct {
	unleash.Feature
	SearchTerm string `json:"-"`
}

func (f FeatureFlags) String(i int) string {
	return f[i].Name
}

func (f FeatureFlags) Len() int {
	return len(f)
}

func (f FeatureFlags) Get(key string) (Feature, error) {
	for _, flag := range f {
		if flag.Name == key {
			return flag, nil
		}
	}

	return Feature{}, errors.New("not found")
}
