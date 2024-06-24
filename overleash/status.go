package overleash

import (
	"fmt"
	"github.com/Unleash/unleash-client-go/v4/api"
	"slices"
)

type FeatureFlagStatus struct {
	Strategy string
	Status   string
}

func (fr FeatureFile) FeatureFlagStatus(featureFlag string) []FeatureFlagStatus {
	statuses := make([]FeatureFlagStatus, 0, 1)
	idx := slices.IndexFunc(fr.Features, func(f Feature) bool { return f.Name == featureFlag })

	if idx == -1 {
		return statuses
	}

	flag := fr.Features[idx]

	for _, strategy := range flag.Strategies {
		name, status := parseFromStrategy(strategy)

		suffix := ""
		if len(strategy.Constraints) > 0 {
			suffix = " (with constraints)"
		}

		statuses = append(statuses, FeatureFlagStatus{
			Strategy: name,
			Status:   status + suffix,
		})
	}

	return statuses
}

func (fr FeatureFile) FeatureFlagEnabled(featureFlag string) bool {
	idx := slices.IndexFunc(fr.Features, func(f Feature) bool { return f.Name == featureFlag })

	if idx == -1 {
		return false
	}

	flag := fr.Features[idx]

	return flag.Enabled
}

func parseFromStrategy(strategy api.Strategy) (string, string) {
	switch strategy.Name {
	case "default":
		return "Default", "On"

	case "flexibleRollout":
		return "Rollout", fmt.Sprintf("%s%%", rollout(strategy.Parameters))

	case "gradualRolloutRandom":
		return "Random rollout", fmt.Sprintf("%s%%", percentage(strategy.Parameters))

	case "gradualRolloutSessionId":
		return "Session Id Rollout", fmt.Sprintf("%s%%", percentage(strategy.Parameters))

	case "gradualRolloutUserId":
		return "User Id Rollout", fmt.Sprintf("%s%%", percentage(strategy.Parameters))

	case "userWithId":
		return "User ids", fmt.Sprintf("%s%%", userIds(strategy.Parameters))

	case "remoteAddress":
		return "IP", ips(strategy.Parameters)

	case "applicationHostname":
		return "Hosts", hostNames(strategy.Parameters)
	}

	return "", ""
}

func rollout(parameterMap api.ParameterMap) string {
	return parameterMap["rollout"].(string)
}

func ips(parameterMap api.ParameterMap) string {
	return parameterMap["IPs"].(string)
}

func hostNames(parameterMap api.ParameterMap) string {
	return parameterMap["hostNames"].(string)
}

func percentage(parameterMap api.ParameterMap) string {
	return parameterMap["percentage"].(string)
}

func userIds(parameterMap api.ParameterMap) string {
	return parameterMap["userIds"].(string)
}
