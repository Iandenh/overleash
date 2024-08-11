package overleash

import (
	"fmt"
	"github.com/Unleash/unleash-client-go/v4/api"
	"slices"
	"strings"
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
		return "Standard", "On"

	case "flexibleRollout":
		return "Gradual rollout", fmt.Sprintf("%s%%", rollout(strategy.Parameters))

	case "gradualRolloutRandom":
		return "Randomized", fmt.Sprintf("%s%%", percentage(strategy.Parameters))

	case "gradualRolloutSessionId":
		return "Sessions", fmt.Sprintf("%s%%", percentage(strategy.Parameters))

	case "gradualRolloutUserId":
		return "Users", fmt.Sprintf("%s%%", percentage(strategy.Parameters))

	case "userWithId":
		return "UserIDs", fmt.Sprintf("%s%%", userIds(strategy.Parameters))

	case "remoteAddress":
		return "IPs", ips(strategy.Parameters)

	case "applicationHostname":
		return "Hosts", hostNames(strategy.Parameters)
	}

	return "", ""
}

func ToStrategyName(strategy api.Strategy) string {
	switch strategy.Name {
	case "default":
		return "Standard"

	case "flexibleRollout":
		return "Gradual rollout"

	case "gradualRolloutRandom":
		return "Randomized"

	case "gradualRolloutSessionId":
		return "Sessions"

	case "gradualRolloutUserId":
		return "Users"

	case "userWithId":
		return "UserIDs"

	case "remoteAddress":
		return "IPs"

	case "applicationHostname":
		return "Hosts"
	}

	return ""
}

func ToLabelText(strategy api.Strategy) string {
	switch strategy.Name {
	case "default":
		return "The standard strategy is <span>ON</span> for all users."
	case "flexibleRollout":
		extra := ""
		if len(strategy.Constraints) > 0 {
			extra = "who match constraints "
		}

		return fmt.Sprintf("<span>%s%%</span> of your base %sis included", rollout(strategy.Parameters), extra)

	case "remoteAddress":
		count := len(strings.Split(ips(strategy.Parameters), ","))

		return fmt.Sprintf("%d IPs will get access: %s", count, ips(strategy.Parameters))
	}

	return ""
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
