package overleash

import (
	"errors"
	"time"
)

type ParameterMap map[string]interface{}

type FeatureFlags []Feature

func (fr FeatureFile) SegmentsMap() map[int][]Constraint {
	segments := make(map[int][]Constraint, len(fr.Segments))
	for _, segment := range fr.Segments {
		segments[segment.Id] = segment.Constraints
	}

	return segments
}

type FeatureFile struct {
	Version int `json:"version"`

	Features FeatureFlags `json:"features"`
	Segments []Segment    `json:"segments"`
}

type Segment struct {
	Id          int          `json:"id"`
	Name        string       `json:"name,omitempty"`
	Constraints []Constraint `json:"constraints"`
}

type Feature struct {
	// Name is the name of the feature toggle.
	Name string `json:"name"`

	// Type is the type of the feature toggle.
	Type string `json:"type"`

	// Enabled indicates whether the feature was enabled or not.
	Enabled bool `json:"enabled"`

	Project string `json:"project"`

	Stale *bool `json:"stale,omitempty"`

	// Strategies is a list of names of the strategies supported by the overleashclient.
	Strategies []Strategy `json:"strategies"`

	CreatedAt  *time.Time `json:"createdAt,omitzero"`
	LastSeenAt *time.Time `json:"lastSeenAt,omitzero"`

	// Strategy is the strategy of the feature toggle.
	Strategy string `json:"strategy,omitempty"`

	// Variants is a list of variants of the feature toggle.
	Variants []Variant `json:"variants"`

	Description string `json:"description"`

	// Dependencies is a list of feature toggle dependency objects
	Dependencies *[]Dependency `json:"dependencies,omitempty"`

	// ImpressionData indicates whether the overleashclient SDK should emit an impression event
	ImpressionData bool `json:"impressionData"`

	SearchTerm string `json:"-"`
}

type Dependency struct {
	Feature  string    `json:"feature"`
	Variants *[]string `json:"variants"`
	Enabled  *bool     `json:"enabled"`
}

type Variant struct {
	Name       string            `json:"name"`
	Weight     int               `json:"weight"`
	WeightType string            `json:"weightType"`
	Stickiness string            `json:"stickiness"`
	Payload    Payload           `json:"payload"`
	Overrides  []VariantOverride `json:"overrides"`
}

type VariantOverride struct {
	ContextName string   `json:"contextName"`
	Values      []string `json:"values"`
}

type Payload struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

type Strategy struct {
	// Name is the name of the strategy.
	Name string `json:"name"`

	SortOrder *int `json:"sortOrder,omitempty"`

	Segments []int `json:"segments,omitempty"`

	// Constraints is the constraints of the strategy.
	Constraints []Constraint `json:"constraints"`

	// Parameters is the parameters of the strategy.
	Parameters ParameterMap `json:"parameters"`

	// Variants for a strategy
	Variants []StrategyVariant `json:"variants"`
}

type StrategyVariant struct {
	Name       string  `json:"name"`
	Weight     int     `json:"weight"`
	Payload    Payload `json:"payload"`
	Stickiness string  `json:"stickiness"`
}

// Operator is a type representing a constraint operator
type Operator string

const (
	// OperatorIn indicates that the context values must be
	// contained within those specified in the constraint.
	OperatorIn Operator = "IN"

	// OperatorNotIn indicates that the context values must
	// NOT be contained within those specified in the constraint.
	OperatorNotIn Operator = "NOT_IN"

	// OperatorStrContains indicates that the context value
	// must contain the specified substring.
	OperatorStrContains Operator = "STR_CONTAINS"

	// OperatorStrStartsWith indicates that the context value
	// must have the specified prefix.
	OperatorStrStartsWith Operator = "STR_STARTS_WITH"

	// OperatorStrEndsWith indicates that the context value
	// must have the specified suffix.
	OperatorStrEndsWith Operator = "STR_ENDS_WITH"

	// OperatorNumEq indicates that the context value
	// must be equal to the specified number.
	OperatorNumEq Operator = "NUM_EQ"

	// OperatorNumLt indicates that the context value
	// must be less than the specified number.
	OperatorNumLt Operator = "NUM_LT"

	// OperatorNumLte indicates that the context value
	// must be less than or equal to the specified number.
	OperatorNumLte Operator = "NUM_LTE"

	// OperatorNumGt indicates that the context value
	// must be greater than the specified number.
	OperatorNumGt Operator = "NUM_GT"

	// OperatorNumGte indicates that the context value
	// must be greater than or equal to the specified number.
	OperatorNumGte Operator = "NUM_GTE"

	// OperatorDateBefore indicates that the context value
	// must be before the specified date.
	OperatorDateBefore Operator = "DATE_BEFORE"

	// OperatorDateAfter indicates that the context value
	// must be after the specified date.
	OperatorDateAfter Operator = "DATE_AFTER"

	// OperatorSemverEq indicates that the context value
	// must be equal to the specified SemVer version.
	OperatorSemverEq Operator = "SEMVER_EQ"

	// OperatorSemverLt indicates that the context value
	// must be less than the specified SemVer version.
	OperatorSemverLt Operator = "SEMVER_LT"

	// OperatorSemverGt indicates that the context value
	// must be greater than the specified SemVer version.
	OperatorSemverGt Operator = "SEMVER_GT"
)

// Constraint represents a constraint on a particular context value.
type Constraint struct {
	// ContextName is the context name of the constraint.
	ContextName string `json:"contextName"`

	// Operator is the operator of the constraint.
	Operator Operator `json:"operator"`

	// Values is the list of target values for multi-valued constraints.
	Values []string `json:"values"`

	// Value is the target value single-value constraints.
	Value string `json:"value"`

	// CaseInsensitive makes the string operators case-insensitive.
	CaseInsensitive bool `json:"caseInsensitive"`

	// Inverted flips the constraint check result.
	Inverted bool `json:"inverted"`
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
